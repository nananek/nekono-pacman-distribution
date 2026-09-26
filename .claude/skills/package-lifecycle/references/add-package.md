# 新規 package の追加

## Step 0: 追加してよいか

- user が名指しで依頼した package か。
- **Arch 公式 (core/extra/multilib) に既にあれば入れる意味が薄い**: `LC_ALL=C pacman -Si <name> | grep -E '^(Repository|Version)'`
  (build host は `[nekono]` も持つので、Repository が `nekono` のものは既に登録済み)。
- 命名: AUR の package 名をそのまま使う。`<name>` と `<name>-bin` は基本どちらか片方。**`-git` (rolling) は避け**、
  upstream の tag を pin する。自家 fork は `<name>-nekono` の前例あり。
- 追加しようとしている物の性質を user に確認したほうがよい点 (prebuilt binary を配るか、from-source か等) は
  依頼文で決まっていなければ聞く。

## Step 1: AUR から fork

```sh
git checkout -b pkg/<name>-<pkgver> origin/master
cd pkgs && git clone https://aur.archlinux.org/<name>.git <name>
# .git を消す前に、REVIEW.md に書く材料を控える
git -C <name> log --format='%h %an %ad %s' --date=short | head -20       # maintainer 交代・不審な履歴の有無
git -C <name> rev-parse --short HEAD                                     # fork 元の AUR commit
rm -rf <name>/.git                                                       # AUR との直接連動を切る
```

`.SRCINFO` も含める。AUR の履歴で **maintainer が最近交代している / orphan からの再取得 / 直近で source URL の
domain が変わった**等があれば supply-chain の警戒信号。REVIEW.md に記録し、user に報告する。

## Step 2: review (`/security-review` skill + 自分の独立検証)

CLAUDE.md「Review 手順」の全項目。**AUR の値を鵜呑みにしない**:

- [ ] `source` の URL が upstream の official path (typosquat / domain spoof 無し)。upstream の公式 release ページと照合
- [ ] sha256 (or sha512) で pin されている。**`SKIP` 不可**。値は自分で download して実測 (`curl -fsSL <url> | sha256sum`)、
      upstream 公開 checksum があれば照合
- [ ] `prepare()` / `build()` / `package()` に curl / wget / exec / eval / pipe-to-shell / 動的 fetch が無い
- [ ] `depends` / `makedepends` が妥当 (prebuilt binary なら `readelf -d` の DT_NEEDED で実際の依存を確認)
- [ ] `prepare()` の patch の中身が合理的
- [ ] `*.install` の hook の中身 (post_install で何をするか)
- [ ] secret 混入なし。`.git` 削除済み
- [ ] 落とし穴 (CLAUDE.md #1–7): `pip install` の動的取得 (→ PyPI dist を vendor + sha256 pin、`--no-index --no-deps
      --no-build-isolation`)、prebuilt bundle の Python ABI 不一致、multi-volume 7z、環境依存の source 分岐 (禁止 →
      variant pkg を別 pkgname に)、prepare 専用 tool は makedepends のみ、`sed -i.bak` の残骸、自前 file の sha256 pin、
      `direct_url.json` 削除、prebuilt `.so` 同梱なら `options=(!strip !debug)`、自前 tarball の `--mtime` に固定 UTC 日付を書かない
- [ ] `arch=('x86_64')` に絞る ([nekono] は `repo/x86_64/` のみ配信。前例: antigravity-cli)。`pkgver` は **literal で書く**
      (`pkgver="${_var}"` のように `$` を含むと `.github/scripts/detect_upstream_updates.py` が比較できず、Issue が永久に立たない)

懸念があれば**ここで止まって user に報告**。改変が必要な箇所 (AUR との意図的 diff) は PKGBUILD にコメントで理由を残し、
REVIEW.md の「依存方針」 (等) に「なぜ AUR と違うか」を書く。

## Step 3: pkgs/<name>/ の一式を整える

| file | 作り方 |
|---|---|
| `PKGBUILD` | AUR fork + 必要な改変 (上記) |
| `.SRCINFO` | `( cd pkgs/<name> && makepkg --printsrcinfo > .SRCINFO )`。**build-all が [nekono] 内 cross-dep と split pkg の pkgname をここから読む**ので、必ず PKGBUILD と同期 |
| `LICENSE`, `*.install`, 自前 service/sysusers 等 | AUR のまま (byte 一致を確認)。自前 file は sha256 を `sha256sums` に pin |
| `REVIEW.md` | 下記 |
| `.deps.lock` | 下記 |

### REVIEW.md

似た性質の既存 pkg の REVIEW.md を雛形にする (prebuilt binary: `antigravity-cli` / `claude-code`、C の from-source:
`icecast`、fork: `sunshine-nekono` 等)。節: **状態** (review 済み/approve の日付) / **用途** / **Source** (AUR URL と
commit、maintainer、upstream URL、tag commit) / **検証結果** (Step 2 の checkbox と、実測した sha256 の値、
根拠) / **依存方針** (AUR との意図的 diff とその理由) / **更新履歴** (表: `日付 | release | review した PKGBUILD repo SHA |
upstream tag commit | findings`。初回行は SHA 欄に `(this PR)`)。

review 日付、review した PKGBUILD の SHA、upstream の commit / release tag (検証して特定したもの)、
findings (受入 / 改変要 / 却下) を必ず書く。将来の bump 手順 (何をどう検証するか) も書いておくと次の担当が楽。

### `.deps.lock` の初期化

これが無い / 空だと `dep-version-pr` はその pkg を**監視しない**。`.SRCINFO` の depends + makedepends の
**pkgname (version 制約を除く)** について、Arch 公式 repo にあるものだけ `name=version` を書き、それ以外は
`# MISSING <name>  -- <理由>` 行にする。build host の pacman.conf には `[nekono]` があり `pacman -Si` が
nekono 内 pkg も解決してしまうので、**Repository を core/extra/multilib に限る**:

```sh
lock_line() { local dep=$1 out repo ver
  out=$(LC_ALL=C pacman -Si "$dep" 2>/dev/null) || { echo "# MISSING $dep  -- not found by pacman -Si (virtual provide / AUR-only)"; return; }
  repo=$(sed -n 's/^Repository *: *//p' <<<"$out" | head -1); ver=$(sed -n 's/^Version *: *//p' <<<"$out" | head -1)   # epoch の ':' を切らない
  case "$repo" in core|extra|multilib) echo "$dep=$ver";; *) echo "# MISSING $dep  -- repo=$repo (not Arch official)";; esac; }
{
  echo "# Snapshot of Arch official repo versions for depends + makedepends of <name>."
  echo "# Auto-generated and updated by .github/workflows/dep-version-pr.yml."
  echo "# Source: pkgs/<name>/.SRCINFO (depends + makedepends entries)."
  echo "# Format: 1 line per package: <name>=<full-version>"
  echo
  awk '/^[[:space:]]*(depends|makedepends) = /{sub(/^[[:space:]]*(depends|makedepends) = /,""); sub(/[<>=].*$/,""); print}' pkgs/<name>/.SRCINFO \
    | sort -u | while read -r d; do lock_line "$d"; done
} > pkgs/<name>/.deps.lock
```

出力を目で確認する (epoch 付き version が `avahi=1:0.9rc5-1` のように丸ごと入っていること)。`# MISSING` 行は
監視対象外の理由を残す**情報用コメント** (`dep_version_check.py` はコメントを無視)。既存 lock は virtual provide
(`libfoo.so` 等) の MISSING 注記の有無が pkg 間で不揃いなので、有っても無くてもよい。理由は実態に合わせて書き換える
(前例: `# MISSING libigloo  -- nekono repo package (not in Arch official repos)`)。

## Step 4: repo への配線

- **`nvchecker.toml`**: `[<name>]` section を追加 (section 名 = `pkgs/` の dir 名。ファイル冒頭コメント参照)。
  - source: `github` + `use_latest_release = true` (+ `prefix = "v"` 等で tag から prefix を除く)、release を切らない repo は
    `use_max_tag = true`、`npm`、`regex`、scraper は `cmd`。
  - **nvchecker が返す文字列が PKGBUILD の `pkgver` と完全一致すること** (検出は文字列比較)。`_`/`-` 等の差は
    `from_pattern` / `to_pattern` で揃える。
  - CI (`upstream-version-issue.yml`) には `python-jq` 等の追加 module が**無い** (`source = "jq"` は失敗する。
    antigravity-cli で踏んだ)。github / npm / regex / cmd を使う。
  - ローカルに `nvchecker` は無い。URL / regex を `curl` で試して、抽出結果が `pkgver` と一致することを手で確認し、
    その根拠を section のコメントに書く。
  - 自家 fork (`*-nekono`) は section を持たない前例がある (upstream 追従を手動運用する pkg)。
- **`README.md`**「配布パッケージ一覧」: 該当カテゴリの表に 1 行 (`| package | source 方針 | 役割 |`)。
- **`.gitignore`**: download した source が既存 pattern に拾われない形式 (拡張子なし binary、git clone dir) なら
  「makepkg の source= ダウンロードキャッシュ」block に pkg 個別 pattern を追加。
- **`CLAUDE.md` の「PR review 時の個別事情」表**: `pkgver` が depends のどれかの version と sync する構造なら**必ず追記**
  (書かないと将来 bot の pkgrel bump PR を merge して旧 pkgver のまま rebuild される事故になる)。

## Step 5: build 前に検証 (repo/ を汚さずに)

```sh
( cd pkgs/<name> && makepkg --verifysource )              # sha256 検証
( cd pkgs/<name> && makepkg -sfc --noconfirm )            # 署名なしの test build。-s は makedepends を sudo で入れる
rm -f pkgs/<name>/*.pkg.tar.zst                           # 成果物は repo/ に入れない (署名・db 登録は merge 後に build-all が行う)
git status --short                                        # stray file が add 対象に混ざっていないこと
```

`bin/build-all <name>` は **merge 前には実行しない** (署名して `repo/x86_64/` の db に登録してしまい、未承認の pkg が
次の `bin/publish` で配信され得る)。sudo で止まったら user に伝える。

## Step 6: commit / push / PR → **承認待ち**

```sh
git add pkgs/<name>/ README.md nvchecker.toml [.gitignore CLAUDE.md]     # 実際に触った path を明示 (pkgs/<name>/ の stray に注意)
git status --short                                                       # 意図しない untracked が無いこと
git commit -S -F - <<'EOF'
<name>: add new package

<何の pkg か 1 行>。AUR `<name>` (commit <hash>, pkgver=<ver>) を fork。<PKGBUILD の要点>。

- source URL は upstream の <公式 path> と一致。sha256 を独立実測し <AUR 値/upstream 公開値> と一致を確認
- <改変点とその理由 / 意図的な AUR との diff>
- <depends の根拠>
- <nvchecker の設定と、pkgver との一致確認>

<harness 指定の attribution 行>
EOF
git push -u origin HEAD          # pre-push gate。BLOCK は直す
gh pr create --base master --title "<name>: add new package" --body "<概要 / review 結果 / AUR との diff / 懸念 / Test plan / 「user 承認待ち、merge 前に build・publish はしない」>"
```

PR を出したら **user に review 結果を要約して報告し、承認を待つ**。承認が出たら:

1. `gh pr merge <N> --merge --delete-branch && git checkout master && git pull --ff-only`
2. `daily-ops` Phase 3 (build) → Phase 4 (publish)。新 pkg は artifact が無いので `bin/build-all --pending` が拾う
3. 最終報告で user に依頼: ansible-nekonodesk の "Self-hosted (nekono repo)" 対象リストへの追加 PR、client での
   `sudo pacman -Sy && sudo pacman -S <name>`

## 復活 (撤廃済み pkg を戻す)

```sh
git log --diff-filter=D --format='%h %ad %s' --date=short -- pkgs/<name>/PKGBUILD      # 撤廃 commit を探す
git checkout <撤廃 commit>^ -- pkgs/<name>                                               # 直前の状態を復元
```

- 撤廃 commit は README 行 / `nvchecker.toml` section / `.gitignore` pattern / CLAUDE.md 表の行も消しているので**戻す**。
- 昔の REVIEW.md を信用せず、**現在の AUR と upstream で Step 2 をやり直す**。現在の upstream 版へ bump し、
  `.deps.lock` を Step 3 の手順で作り直す。REVIEW.md に復活の旨と経緯を書く (前例: claude-code の復活 `1263613`)。
- あとは Step 4 以降と同じ。commit は `<name>: revive at <ver> (re-add after retirement)`。
