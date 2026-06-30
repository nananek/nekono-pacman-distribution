# nekono-pacman-distribution

自作 pacman repo。AUR で配布されている package のうち、self-host したい
ものだけを Claude review + Nekono GPG 署名で固めて Tailscale 経由で配信
する。クライアント (ayaka 等 Arch 機) は `/etc/pacman.conf` の `[nekono]`
セクション経由で `pacman -S` できる。詳細は README.md。

## 規約

### Package 命名

- AUR の package 名をそのまま使う (`vesktop-bin`, `moonlight-qt`, etc.)
- AUR にある `<name>` と `<name>-bin` の両方を抱える必要があれば両方
  入れる。基本は片方だけ採用。
- `<name>-git` (rolling) は避ける。upstream の tag 切れたバージョンを
  pin して再現性を確保。

### PKGBUILD のソース

- AUR から fork:
  ```sh
  cd pkgs && git clone https://aur.archlinux.org/<name>.git <name>
  rm -rf <name>/.git                 # AUR との直接連動を切る
  ```
- `.SRCINFO` も含める (将来 AUR との diff 取りやすい)
- Claude review 通過後に commit

### Review 手順 (1 package = 1 commit)

1. AUR から `pkgs/<name>/PKGBUILD` を fork
2. `/security-review` skill で以下を確認:
   - `source` URL が upstream の official path (typosquat / domain spoof
     無し)
   - `sha256sums` (or `sha512sums`) で tarball が pin されているか
   - `build()` / `package()` step に怪しい curl / wget / exec / shell
     injection 無し
   - `depends` / `makedepends` が想定通り
   - `prepare()` の patch (もしあれば) が中身合理的
3. 結果を `pkgs/<name>/REVIEW.md` に記録:
   - review 日付
   - review した PKGBUILD の本 repo 内 SHA
   - upstream source の commit / release tag (検証して特定したもの)
   - findings: 受入 / 改変要 / 却下
4. 受入 or 改変版を commit (Nekono GPG `-S` 署名必須)

### Pre-push review gate (旧 claude-review.yml の代替)

PR の supply-chain review はかつて GitHub Actions の `claude-review.yml` が
回していたが、**遅すぎ・verdict が check の緑/赤に出ない (verdict は post
された comment 側) 等の理由で廃止**し、**push 前のローカル gate** に移行した。

- 実体: `.githooks/pre-push` → `bin/prepush-review`。有効化は
  `git config core.hooksPath .githooks` (新規 clone / build host では 1 度
  この config を実行すること)。
- `bin/prepush-review` は **決定的に検出できる supply-chain 地雷だけ**を弾く
  (= push 中止): ① secret 混入 ② PKGBUILD の SKIP checksum 新規導入
  ③ build 関数への curl/wget/pipe-to-shell/eval/`pip install` (--no-index 無し)
  追加 ④ `source=`/`*sums`/`pkgver` を変えたのに REVIEW.md 未更新。
- 「source URL が本当に upstream official か」「build step が合理的か」 等の
  **判断系 review は従来どおり人間 (Claude) が PR 処理時に実施**する。script は
  その前段の機械 gate であり、**通過 ≠ 承認**。
- 緊急 override: `git push --no-verify`。

### PKGBUILD でよく踏む落とし穴 (review で繰り返し指摘される項目)

過去 PR の build 失敗 / review 指摘から得た学習。 同種 PKGBUILD
を fork する時は最初から下記を踏まえる:

#### 1. `pip install` を build/package() 内で動的取得しない (supply-chain)

- AUR には `pip install foo bar` (= sha 検証なし) でビルド時 PyPI 取得する
  PKGBUILD がある (= voicevox-engine 等)。 [nekono] は **不可**
- 対処: PyPI distfile (sdist / wheel) を **source[] に vendor + sha256 pin**
  → `pip install --no-index --no-deps --no-build-isolation --target ...` で
  オフライン install
- `--no-build-isolation` を入れないと、 PEP 517 が isolated venv に
  `setuptools>=64` 等を fetch しようとして `--no-index` と衝突して fail
  (= 実例: voicevox-engine-cuda PR #43)
- makedepends に build dep を明示的に列挙 (= `python-setuptools` `python-wheel`
  `python-build` `cython` `gcc` 等)。 sdist が C extension を含むなら gcc / make /
  autoconf / automake も
- Arch 公式 repo に無い Python module (= AUR-only。 例: `python-pyworld` 等) は
  vendor 投入する (= AUR helper 前提を取らない方針)

#### 2. prebuilt bundle 同梱の Python モジュールと system Python ABI 不一致

- upstream の prebuilt bundle (`*.7z`、 `*.AppImage` 等) に `libpythonX.Y.so`
  + 同 X.Y で build した site-packages が同梱されている場合あり (= voicevox-engine の NVIDIA bundle は Python 3.11 build を抱えている)
- Arch の system Python は rolling、 bundle と minor version が違うと
  extension module (`*.so`) の import で ABI 不一致
- 対処: bundle move 後、 PyPI vendor を pip install --target する **前に**
  bundle 由来の Python module dir + `*.dist-info` を `rm -rf` で除去
- `pip install --target` は既存 dir に対し **warn + skip** (= 何もしない)。
  上書きしたいなら事前削除 or `--upgrade` フラグ必須
- 実例: voicevox-engine-cuda PR #46

#### 3. multi-volume 7z (`.7z.001` + `.7z.002`) と 7zip 26.01 の挙動差異

- 古い 7z は `7z x archive.7z.001` で split merge + nested .7z 再帰展開を
  1-step で行ったが、 **7zip 26.01 は split を merge して中間 .7z を出すだけ**
  だったり、 出力 filename が予測不能だったりする (= voicevox-bin PR #44 / #45)
- 確実な方法:
  ```sh
  cat foo.7z.001 foo.7z.002 > _merged.7z
  7z -y x _merged.7z
  rm _merged.7z
  ```
- 出力 filename も bash glob で対応:
  ```sh
  _arr=( PATTERN* ); _file="${_arr[0]}"
  [[ -f "$_file" ]] || { echo "[FATAL] not found"; ls -la >&2; exit 1; }
  ```
- makedepends に `7zip` (= `7z` binary)。 libarchive (`bsdtar`) は
  multi-volume 7z を扱えない

#### 4. 環境依存で source[] を分岐させる PKGBUILD は禁止 (非決定 build)

- `if [ -c /dev/nvidia0 ]; then source=(...NVIDIA AppImage...)` のように
  build host の GPU 有無で source を動的切替する PKGBUILD は禁止 (= sha256
  pin が片方しか定義されない supply-chain 欠陥)
- 必要なら **variant pkg を別 pkgname で分ける** (= `<name>-cuda` 等)。
  `provides=<name>` / `conflicts=<name>` で AUR ベース pkg を満たす
- 実例: voicevox-engine-cuda (PR #39)、 voicevox-bin (PR #40 で if 削除)

#### 5. 細かい lint (review で頻繁に指摘される項目)

- **prepare-only tool は makedepends のみ**、 depends に重複させない (= 例:
  `7zip` を AppImage 展開のために `prepare()` で使うだけなら makedepends 専用)
- **`sed -i.bak`** で生成される `*.bak` ファイルを `rm -f` で後片付け (=
  srcdir 残骸の review 指摘回避)
- **自前 file (service unit / sysusers.d / tmpfiles.d 等) の sha256** を
  source[] と sha256sums に必ず pin。 ファイル内容変更時は再計算を忘れない
- **REVIEW.md 更新履歴 section に当該 release の 1 行を必ず追記** (= PR で
  PKGBUILD 触ったら追記、 build fix 系も "X.Y.Z (build fix)" 行を入れる)。
  `source=`/`*sums`/`pkgver` を変えた場合は pre-push gate (`bin/prepush-review`)
  が REVIEW.md 未更新を block する
- **AUR との意図的 diff** は REVIEW.md 「依存方針」 等の section に
  「なぜ AUR と違うのか」 を記載。 そうしないと bump の度に同じ指摘が再発する

#### 6. `pip install --target` 後の cleanup と prebuilt .so 同梱 pkg の options

- **PEP 610 の `direct_url.json` を削除**: `pip install --target ... <local-file>`
  は dist-info 配下に `direct_url.json` を生成し、 そこに $srcdir 配下の
  絶対 path を埋め込む。 published .pkg.tar.zst に build host path がリーク
  + makepkg lint が「パッケージは $srcdir へのリファレンスを含んでいます」 警告を出す
  → pip install 後に `rm -f "$pkgdir"/.../<target>/*.dist-info/direct_url.json`
- **`options=(!strip !debug)` を入れる**: bundle 同梱の prebuilt `.so` (= CUDA /
  cuDNN / onnxruntime 系) は upstream で既に strip 済み + debug symbol 無し。
  指定しないと makepkg の debug pkg 生成 + gdb-add-index 試行が大量に warning
  を吐く (= `gdb-add-index: No index was created for ...`)。 voicevox-bin /
  voicevox-engine-cuda はこの指定あり
- 実例: voicevox-engine-cuda (PR #48)

### Commit policy

- commit は **必ず `-S` 署名** (Nekono GPG)。verify-commit で trust chain
  が切れない状態を保つ。
- 1 PKGBUILD update = 1 commit。複数 package を 1 commit に混ぜない (review
  履歴を package 単位で追えるように)。
- 大規模な `build-all` 変更等は別 commit (PKGBUILD 系と切り分け)。

### Task name は ASCII で

ansible-nekonodesk と同じ理由 (TTY 文字化け防止)。bin/ script の echo / log
出力は英語で書く。コメント・README は日本語 OK。

### Secret は commit しない

GPG 秘密鍵 / Tailscale 認証 / build host hostname の機微なものは commit
しない。`bin/serve` の Server URL 等は環境変数 or build host 側設定で
注入する。

### Build artifact は repo/ で gitignore

`pkgs/<name>/src/` / `pkgs/<name>/pkg/` / `pkgs/<name>/*.pkg.tar.zst` 等の
makepkg 中間 / 成果物は `.gitignore` で除外。`repo/` 配下も同様。
GitHub には PKGBUILD と build 用スクリプトのみ載る。

## 自動化 (GitHub Actions)

`.github/workflows/` に cron で回る workflow を 2 本置いて、PKGBUILD 管理の
退屈な部分を半自動化する。

### `upstream-version-issue.yml` (daily, 02:50 UTC = JST 11:50)

**何をする**: 各 pkg の **upstream の新 version** を検知 → Claude が事前
調査 → **GitHub Issue を立てる** (PR ではない)。

1. `nvchecker.toml` を元に upstream の latest version を fetch
2. PKGBUILD の現 `pkgver` と比較 → 差分のある pkg list を出す
   (`.github/scripts/detect_upstream_updates.py`)
3. 各 pkg について matrix job:
   - 冪等性 check: Issue title `[<pkg>] upstream version: <new_pkgver>`
     を fingerprint に `gh issue list --state all` で重複 search、あれば skip
   - Claude code action が走り:
     - 新旧 upstream tarball の build script diff
     - dependency 変更
     - CHANGELOG / Release notes 要約
     - **supply-chain 監査** (source URL / sha256 / install script 追加 /
       maintainer 変化 / 新 dep の typosquat 等)
     を行い、結果を markdown で Issue 本文に書いて `gh issue create`
4. **PKGBUILD は触らない**。人間が Issue を読んで build host で手作業更新
   する。その後 PR を立てると、push 時の pre-push gate (`bin/prepush-review`)
   が機械チェックし、PR 処理時に人間が judgment review する (= 2 段審査)。

**新規 pkg を足したら**: ルート `nvchecker.toml` に section 追加が必要。
ファイルの head コメント参照。

### `dep-version-pr.yml` (daily, 02:55 UTC = JST 11:55)

**何をする**: 各 pkg の **依存パッケージ (Arch 公式 repo) の version 変化**
を検知 → **`pkgrel` を +1** + `.deps.lock` 更新 → **PR を出す** (機械的処理、
ABI 変化に伴う rebuild 指示)。

1. `pkgs/<pkg>/.deps.lock` (= 前回の `pacman -Si` snapshot) と現在の
   `pacman -Si` 出力を比較
2. 差分のある pkg について:
   - PKGBUILD の `pkgrel` を +1
   - `.deps.lock` を新値で書き直す
   - branch `deps/<pkg>-pkgrel-<N+1>` を切って commit + push
   - 冪等性 check: `gh pr list --head <branch> --state all` で重複 PR
     を search、あれば skip
   - `gh pr create --base master`
3. **bot commit は unsigned**。merge 前に人間が build host で local に
   pull → 当該 branch を checkout → `git commit --amend -S --no-edit` で
   sign し直して force-push → merge (CLAUDE.md「commit は必ず -S 署名」
   の整合性確保)。

**新規 pkg を足したら**: `pkgs/<pkg>/.deps.lock` を手動で初期化 (ローカル
Arch host で .SRCINFO の depends + makedepends について `pacman -Si <dep>`
の Version 行を集める)。詳細は script `.github/scripts/dep_version_check.py`
の冒頭 docstring 参照。

### 共通の前提

- どちらの workflow も `container: archlinux:latest` で動かす (pacman /
  nvchecker / makepkg と同じ環境を runner 上に確保)
- `.deps.lock` の `# MISSING ...` 行は「Arch 公式 repo に無い (virtual
  provides / AUR-only) ので監視対象外」を意味、`dep_version_check.py` は
  これを silently skip
- secrets: `CLAUDE_CODE_OAUTH_TOKEN` (workflow A の Claude 起動)、`GITHUB_TOKEN`
  (default 付与、Issue / PR / git push 用)

### PR review 時の個別事情 (人間判断要)

`dep-version-pr` は機械的に **「depends.X の version が変わった → pkgrel
+1 で rebuild」** を仮定する。しかし PKGBUILD によっては「depends の
特定 pkg と pkgver が sync しているので、本当は pkgver も上げる必要
がある」というケースがある。bot ロジック側に特例を入れず、PR review
時に人間がここの一覧と照らし合わせて読み替える運用とする。

| pkg | sync 対象 dep | 来た PR をどう処理するか |
|---|---|---|
| `docker-rootless-extras` | `docker` | source URL が `docker-v${pkgver}` で moby/moby tag に pinned。`depends.docker` が動いたら本来 `pkgver` も同 version に上げて sha256sums 再計算が必要。来た pkgrel bump PR は **close**、`upstream-version-issue.yml` が moby/moby の release を nvchecker で拾って Issue を立てる経路で対応する。Issue が立つまでに時差があれば手動で `workflow_dispatch` で trigger。 |

新規 pkg を足すとき、`pkgver` が depends のどれかと version sync する
構造なら **必ずこの表に追記** すること。書かないと将来「来た pkgrel
bump PR をうっかり merge → 旧 pkgver のまま rebuild されて配布パッケージ
が更新されない」事故を踏む。

## Claude への委任フロー (全自動)

「PR 消化して」「Issue 確認して」と言われたら、以下のフローを **ユーザ確認
なしで自律実行** する。途中で判断に迷う場合 (block verdict、PKGBUILD 大改修
が必要な場合等) のみ報告して止まる。

### bot PR (dep-version-pr) の全自動フロー

```
gh pr list --state open  →  deps/<pkg>-pkgrel-<N> の PR を列挙
```

各 PR に対して順番に:

1. **個別事情テーブル確認** (後述「PR review 時の個別事情」)。該当する pkg
   (= `docker-rootless-extras` 等) なら `gh pr close <N>` して次へ。

2. **3 点チェックを事前修正** (repo 規約の hygiene。 pkgrel bump は
   pre-push gate を素通りするので、 下記は人手で揃える):
   - `pkgs/<pkg>/REVIEW.md` の「更新履歴」テーブルに 5 列行を追記
     (`| 日付 | release | SHA | — (pkgrel bump のみ) | findings |`)
   - `pkgs/<pkg>/.SRCINFO` の `pkgrel` を PKGBUILD と一致させる
   - `pkgs/<pkg>/.deps.lock` の `# MISSING ...` 行が消えていたら復元
     (bot は AUR-only / virtual provide の MISSING 行を削除してしまう)

3. `git commit --amend -S --no-edit` + `git push --force-with-lease`
   (push 時に pre-push gate = `bin/prepush-review` が走る。 block されたら
   内容を直して再 push。 pkgrel bump は素通りが正常)

4. push が通れば **待つべき CI は無い**。 そのまま
   `gh pr merge <N> --merge --delete-branch`

5. 全 PR 処理後、build host で実行するコマンドをユーザに提示:
   ```sh
   git pull --ff-only && bin/build-all --pending
   ```

### upstream Issue (upstream-version-issue) の全自動フロー

```
gh issue list --state open  →  [<pkg>] upstream version: <new_pkgver> の Issue を列挙
```

各 Issue に対して:

| verdict | Claude の動作 |
|---|---|
| `safe-to-bump` | Issue 本文の "Suggested PKGBUILD changes" に従い PKGBUILD/sha256/.SRCINFO/REVIEW.md を更新 → `git commit -S` → `gh pr create` (push 時に pre-push gate 通過) → `gh pr merge` → ユーザに `bin/build-all <pkg>` を提示 |
| `needs-attention` | Issue 本文の build script / dep 変化 section を読み込み、PKGBUILD を適切に改修 → あとは safe-to-bump と同じ |
| `block` | ユーザに内容を要約して報告し、判断を仰ぐ。自動では何もしない |
| 誤検知疑い (nvchecker の誤検知等) | 内容を示しユーザに確認してから close |

### build host 作業 (Claude には実行できない)

`bin/build-all` / `bin/update-repo` は build host (nekono-pacman0) 上での実行が必要。
Claude は merge 後に実行コマンドを提示するのみ:

```sh
cd ~/nekono-pacman-distribution
git pull --ff-only
bin/build-all --pending
```

---

## デイリー運用 (Issue / PR 消化)

`upstream-version-issue` / `dep-version-pr` の 2 本の cron workflow が毎日
JST 11:50〜11:55 (= 02:50〜02:55 UTC) に新規 Issue / PR を生成する。
Claude に「PR 消化して」と伝えれば上記の全自動フローで処理する。

### Issues タブ (upstream-version-issue 経由)

新規 Issue title は `[<pkg>] upstream version: <new_pkgver>`。本文の 1 行目
に Claude が付ける **verdict**:

| verdict | 対応 |
|---|---|
| `safe-to-bump` | PKGBUILD の pkgver / sha256sums を本文の Suggested PKGBUILD changes に従って差し替え → `.SRCINFO` 同期 → `git commit -S` → PR open (push 時に pre-push gate 通過) → judgment review して merge → `bin/build-all <pkg>` → Issue close。 |
| `needs-attention` | build script / depends 変化あり。Claude が事前調査 section を読み込み PKGBUILD を改修。あとは safe-to-bump と同じ流れ。 |
| `block` | supply-chain 上の懸念あり。Claude がユーザに報告して止まる。 |

verdict 判定に疑義がある場合 (例: nvchecker の誤検知で立った Issue) は、
内容を読んで close する。冪等性 key は title なので、close されていても
同 title の再検知は skip される (= 二重立て防止)。

### Pull Requests タブ

#### bot 自動 PR (`dep-version-pr` 経由)

branch 名 `deps/<pkg>-pkgrel-<N+1>`、title `<pkg>: pkgrel bump to <N+1>
(deps changed)`。

1. **「PR review 時の個別事情」 表 (上記) で該当 pkg を確認**。該当する
   なら `gh pr close <N>` して別経路 (upstream-version-issue) で対応。
2. 3 点チェック事前修正 (REVIEW.md / .SRCINFO / .deps.lock)。
3. `git commit --amend -S --no-edit` + `git push --force-with-lease`
   (push 時に pre-push gate が走る)。
4. push が通れば CI 待ちは無い。`gh pr merge <N> --merge --delete-branch`。

#### 手動 bump PR (= upstream-version-issue 由来の PR)

`pkgs/<pkg>` ディレクトリの PKGBUILD / .SRCINFO / .deps.lock を更新して
立てる PR。CLAUDE.md 規約「1 PKGBUILD update = 1 commit」「`-S` 署名」を
守る。

merge 前のチェックリスト:

- [ ] push 時の pre-push gate (`bin/prepush-review`) が通っていること
      (block されたら内容を直してから push。 override は緊急時の
      `git push --no-verify` のみ)
- [ ] **judgment review** (機械 gate の通過 ≠ 承認): diff を読んで source URL
      が upstream official か / build 関数に怪しい step が無いか / depends が
      整合するかを自分で確認
- [ ] `REVIEW.md` の「更新履歴」 section に当該 release の 1 行を追記済み
      (review 日付、PKGBUILD repo SHA、upstream tag commit SHA、findings)
- [ ] AUR との意図的 diff (例: `slirp4netns` を depends に維持) があれば
      REVIEW.md の「依存方針」 section に理由を書く
- [ ] build host で `makepkg --verifysource` で sha256 検証、build 出力に
      新規 install path / 警告がないか目視確認
- [ ] PR title / body で `Closes #N` を入れ、merge で対応 Issue が auto
      close されること

### 朝 routine (推奨)

毎朝 1 度、Claude に「PR/Issue 消化して」と伝える。Claude が全自動で
Issues/PRs を処理する。終了後、build host で:

```sh
cd ~/nekono-pacman-distribution
git pull --ff-only
bin/build-all --pending   # 末尾で repo-add まで完了。"nothing to build" が出れば消化完了
```

`bin/build-all` 完了後、 build host 自身を含む client 側で
`sudo pacman -Sy && pacman -Si <pkg>` で新 version が解決できれば配信反映済み
(= nginx は Tailscale 越し file 配信のみ、 reload 不要)。

## Package の退役 (retirement)

不要になった package を [nekono] から外す手順。 `pkgs/<name>/` を消すだけでは
**db と repo/x86_64/ の `.pkg.tar.zst` が残る** ので、別途 `repo-remove` +
`bin/prune` が必要。 `bin/build-all` が両方を自動でやるように整えてある。

1. ローカル (この repo, master):
   ```sh
   git rm -r pkgs/<pkg>/
   # README.md「配布パッケージ一覧」 から該当行を削除
   # nvchecker.toml の [<pkg>] section を削除
   git commit -S -m "<pkg>: retire (<reason>)"
   git push
   ```
2. Build host:
   ```sh
   cd ~/nekono-pacman-distribution
   git pull --ff-only
   bin/build-all --pending
   # build-all は pkgs/ に PKGBUILD が無く db に entry が残っている pkg を
   # 検出し、 repo-remove --sign で db から外したあと bin/prune を呼んで
   # `.pkg.tar.zst{,.sig}` も削除する。 retirement は --pending / 個別 pkg
   # 指定とは独立に常に走る。
   ```
3. Client 側 (ayaka 等):
   ```sh
   sudo pacman -Rns <pkg>     # まだ install されていれば
   sudo pacman -Sy            # repo db refresh
   ```
4. ansible-nekonodesk: `roles/os_packages/vars/Archlinux.yml` の
   "Self-hosted (nekono repo)" block から該当行を消す PR を別途出す
   (= 次回 ansible run で client から消える)。

## ansible-nekonodesk との分担

| 領域 | 置き場所 |
|---|---|
| PKGBUILD / build script / signing pipeline | nekono-pacman-distribution (この repo) |
| build host 自体のセットアップ (apt 系・gpg-agent・pcscd・nginx 設定) | ansible-nekonodesk の Arch path 経由で別 host として管理 |
| client 側 `/etc/pacman.conf` の `[nekono]` 設定 + pacman-key 信頼 | ansible-nekonodesk の `pacman_nekono_repo` role (新規) |
| self-hosted apps の install (`pacman -S vesktop-bin` 等) | ansible-nekonodesk の `apt_packages/vars/Archlinux.yml` の "Self-hosted (nekono repo)" block |

## トラブル時の参照ルール

| 症状 | 疑う場所 |
|---|---|
| `pacman -Sy` で `[nekono]` の signature 検証失敗 | (1) nekono.db.tar.gz.sig が無い (= bin/update-repo が `--sign` 付きで走っていない) (2) client 側 pacman-key の Nekono key trust が切れている (3) build host で署名した GPG key と client が信頼している key の fingerprint 不一致 |
| `[nekono]` repo にアクセスできない (404 / 接続不可) | (1) build host の nginx が落ちている (2) Tailscale が切れている (3) pacman.conf の Server URL が古い (4) repo/x86_64/ にファイルが無い |
| `makepkg` で signature 検証失敗 | upstream の release tarball が差し替わった可能性 → PKGBUILD の sha256sums を再 review して更新、または PKGBUILD のバージョンを bump |
| Claude review 後の PKGBUILD が build host で fail | makedepends 不足 / upstream の API change / 等。build host で `makepkg --verifysource` だけ走らせて切り分け |
