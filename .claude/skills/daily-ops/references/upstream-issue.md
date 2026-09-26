# Phase 2: upstream Issue → 調査 → bump PR

Issue は `.github/scripts/create_upstream_issue.py` が nvchecker の差分から**機械的に**立てる。
title `[<pkg>] upstream version: <new_pkgver>`、本文は現 PKGBUILD の `source=` 引用 + human review
checklist のみ。**verdict も supply-chain 監査も付いていない**。調査と判断は自分がここで行う。

## Step 0: 列挙とグルーピング

```sh
gh issue list --state open --json number,title --jq '.[] | "\(.number)\t\(.title)"'
```

- 同じ pkg の Issue が複数ある (例: claude-code 2.1.281 / .282 / .283) ときは、**調査時点の最新版に集約して 1 PR**。
  最新版が実在・安定していること (取り下げ/yank/pre-release でない) を確認してから。Issue title の version より
  新しい版を使ってよい (PR に明記)。
- **PR 本文の close 構文は Issue ごとに keyword を付ける**: `Closes #a, closes #b, closes #c`。`Closes #a, #b, #c`
  だと GitHub は **最初の 1 件しか自動 close しない** (2026-09-27 に実測。`#b`/`#c` は open のまま残った)。
  取りこぼしたら `gh issue close <N> --comment "PR #<PR> で解消 (最新版に集約)"` で閉じる。merge 後に
  `gh issue view <N> --json state` で全件 CLOSED を確認する。
- 冪等性 key は title。close 済みの同 title は再検知されない。

## Step 1: その pkg の過去のやり方を読む

**`pkgs/<pkg>/REVIEW.md` が pkg 別知識の唯一の置き場**。「更新履歴」と bump 手順相当の section を読み、
前回と**同じ検証方法**をとる (pkg ごとに違う: claude-code は npm tarball の file 単位 diff + 公式 `manifest.json`、
docker-rootless-extras は moby の tag、AUR 由来の -bin は upstream 公開 checksum との照合、等)。
「AUR との意図的 diff」が書かれていたら**維持する** (AUR の PKGBUILD をそのままコピーしない。
例: AUR は `SKIP` でも本 repo は実測 pin)。

## Step 2: 調査 (Issue の checklist を具体化)

1. **公式 release/tag と source URL**: PKGBUILD の source の origin (domain / owner / repo) が従来と同じ upstream か。
   typosquat / domain spoof / repo transfer が無いか。release author / publisher が従来と同じか。
   (`gh release view <tag> -R <owner>/<repo>`、`gh api repos/<o>/<r>/releases/tags/<tag>`、`git ls-remote --tags`、
   `npm view <pkg> versions dist-tags` 等)
2. **sha256 を自分で実測**: `curl -fsSL "<url>" | sha256sum`。AUR の値・upstream が公開する checksum
   (manifest / `.sha256` / release asset) と照合し、**Issue や AUR の値を鵜呑みにしない**。
   `updpkgsums` があれば計算には使ってよいが、それ単体は検証にならない (cross-check が本体)。
   **`SKIP` は不可** (pre-push gate が block)。arch 別 sums (`sha256sums_x86_64` 等) は全 arch。
3. **旧新の差分**: `build()` / `package()` / `prepare()` / install script に効くファイル (Makefile、Cargo.toml、
   pyproject.toml、package.json の scripts、`*.install` 等) を旧新 tarball で diff。依存の追加・変更。
4. **release notes**: breaking change、security fix、packaging / install 経路の変更。
5. **depends の妥当性**: 新版で要る依存が変わっていないか。prebuilt binary は**実行せず** `readelf` で見る
   (`LC_ALL=C readelf -d <bin> | awk '/NEEDED/{print $NF}'`、`LC_ALL=C readelf -V <bin>` の最大 `GLIBC_*`)。
   **`LC_ALL=C` を付ける** (この host は日本語 locale で、付けないと出力が翻訳され awk の抽出が空になる)。
   複数の `.so` を含む bundle は NEEDED の和集合を旧新で `diff`。`strings` の宿主名比較は Go/Flutter の symbol 名が
   ノイズになるので、scheme 付き URL (`https?://...`) の集合で比べる。native lib の内容が変わっていても、
   upstream の toolchain / 依存更新 (`gh api repos/<o>/<r>/compare/<old>...<new>` で `.fvmrc` / lockfile / manifest
   を確認) で説明がつくかを見る。lockfile の diff では、新規 package/crate の追加、git 依存の owner・URL の変更、
   registry 外 source の増減に注目する。
6. **AUR の最新 PKGBUILD との diff**: `git clone https://aur.archlinux.org/<pkg>.git` を scratch dir に取り、
   AUR 側で何が変わったか確認。取り込むべき変更か、本 repo の意図的 diff かを切り分ける。
7. CLAUDE.md「PKGBUILD でよく踏む落とし穴」#1–7 に触れる変更でないか
   (build 時の動的取得、bundle の Python ABI、multi-volume 7z、環境依存 source 分岐、自前 tarball の `--mtime` 等)。

## Step 3: verdict

| 調査結果 | 動作 |
|---|---|
| pkgver + sha256 の機械的更新のみ (safe-to-bump) | Step 4–5 へ |
| build/depends の手当が要る (needs-attention)。既存 pkg を今まで通り build 可能に保つ範囲 | PKGBUILD を適切に改修して Step 4–5 へ |
| supply-chain 上の懸念 / breaking change (block) | **何も merge せず**、調査結果を user に要約して報告。Issue は open のまま `gh issue comment <N> --body "保留: <理由>"` |
| 誤検知疑い (nvchecker の regex 崩れ、pre-release 拾い等) | 根拠を user に示して確認してから close |

block 相当の具体例と「自律の境界」は `SKILL.md` を参照。needs-attention の範囲を超える改修
(新規 pkg 追加、source origin 変更、install hook 新設) は PR を出して **user 承認を待つ**。

## Step 4: PKGBUILD 等の更新

```sh
git checkout -b pkg/<pkg>-<new_pkgver> origin/master
```

- `pkgver` = 新版、**`pkgrel` は 1 にリセット**。sha256sums 更新。URL に version が埋まっていなければ URL も。
- `.SRCINFO`: `( cd pkgs/<pkg> && makepkg --printsrcinfo > .SRCINFO )` → `git diff pkgs/<pkg>/.SRCINFO` が
  pkgver / source / sha256 (+ 意図した変更) だけであること。
- `.deps.lock`: depends / makedepends を変えた時だけ更新 (`# MISSING ...` 行は維持)。
- `REVIEW.md` (pre-push gate は `source=` / `*sums` / `pkgver` を変えて REVIEW.md を触らないと block):
  - 冒頭「状態」の最新日付・版
  - 「検証結果」の実測値 (sha256 等) を新版に更新 (前回の bump で更新漏れして古い版の値が残っていることがある
    ── aria-bin は 1.5.13 の値が 1.5.12 のままだった。見つけたら現行値に揃えて、その旨を履歴に書く)
  - 「更新履歴」に 1 行 (日付 / release / `(this PR)` / upstream tag commit / findings)。build fix なら `X.Y.Z (build fix)`
  - AUR との意図的 diff があれば「なぜ違うか」を書く (書かないと bump のたびに同じ指摘が再発)
- `( cd pkgs/<pkg> && makepkg --verifysource )` で sha256 検証。source が `pkgs/<pkg>/` に download されるので
  `git status --short` で stray file が add 対象に混ざらないこと (`.gitignore` が拾わない形式の cache は
  `.gitignore` の「makepkg の source= ダウンロードキャッシュ」block に pattern を足す)。
  - 判定は「成功」/「失敗」の表示か rc で行う (`| tail` を付けると rc は `tail` のものになる)。
  - **古い download cache の罠**: source のファイル名に version が入っていない (`name::https://.../${tag}/file` の
    `name` が固定。例: aria-bin の `com.poppingmoon.aria.metainfo.xml`) と、`pkgs/<pkg>/` に残った**前の版の cache を
    makepkg が再利用**し、新しい pin と合わず `整合性チェックをパスしませんでした` で失敗する。`bin/build-all` も同じ
    ところで落ちる。失敗した source について「cache の sha256 = 旧 pin」「今 upstream から取り直した値 = 新 pin」
    を確認して、**cache 側が stale だと分かったら**その 1 ファイルだけ (gitignore 済みの再取得可能な cache)
    削除して再実行する。PKGBUILD の pin を cache の値に合わせてはいけない。
  - `gh pr diff` はパス絞り込みを受け付けない。PKGBUILD 部分だけ見るなら
    `gh pr diff <N> | awk '/^diff --git/{show = ($0 ~ /PKGBUILD/)} show'`。

## Step 5: commit / PR / merge

```sh
git add pkgs/<pkg>/PKGBUILD pkgs/<pkg>/.SRCINFO pkgs/<pkg>/REVIEW.md    # + .deps.lock (変えた時)
git commit -S -F - <<'EOF'
<pkg>: bump to <new_pkgver>

Issue #<N>。<何を独立に確認したか: sha256 実測・公式 manifest との一致、tarball diff の範囲、
release notes に packaging 変更なし、等>。
<AUR との diff があれば 1 行>

<harness 指定の attribution 行>
EOF
git push -u origin HEAD          # pre-push gate。BLOCK は直す。--no-verify は使わない
```

security fix なら title に ` (security fix)` (前例: b60f69a)。

```sh
gh pr create --base master --title "<pkg>: bump to <new_pkgver>" --body "$(cat <<'EOF'
## 概要
<pkg> を <old> → <new> に bump。<safe-to-bump / needs-attention の別と一言>。

Closes #<N>[, closes #<M> ...]      # 各 Issue の前に keyword が要る (`Closes #N, #M` は N しか閉じない)

## 調査結果 (詳細は `pkgs/<pkg>/REVIEW.md`)
- <sha256 の独立実測と照合結果>
- <旧新 diff の範囲>
- <release notes の要点。packaging / install 経路の変更有無>

## AUR との diff
<あれば。無ければ「なし」>

## Test plan
- [x] `makepkg --verifysource`
- [ ] build host で `bin/build-all --pending` (merge 後)

<harness 指定の attribution 行>
EOF
)"
```

merge の前に **judgment review**: `gh pr diff <PR>` を自分で読み、source URL が upstream official か、build 関数に
怪しい step が無いか、depends が整合するかを確認 (pre-push gate の通過は承認ではない)。問題なければ:

```sh
gh pr merge <PR> --merge --delete-branch
git checkout master && git pull --ff-only
```

`Closes #N` で対応 Issue が自動 close される。集約した古い Issue も本文の `Closes` で閉じる。
merge 後は Phase 3 (build) へ。
