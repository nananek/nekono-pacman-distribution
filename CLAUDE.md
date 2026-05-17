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

### `upstream-version-issue.yml` (daily, 06:00 UTC)

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
   する。その後 PR を立てれば既存 `claude-review.yml` の supply-chain
   review が走る (= 2 段審査)。

**新規 pkg を足したら**: ルート `nvchecker.toml` に section 追加が必要。
ファイルの head コメント参照。

### `dep-version-pr.yml` (daily, 06:15 UTC)

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
