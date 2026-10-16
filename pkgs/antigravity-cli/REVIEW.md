# antigravity-cli review

## 状態

**review 済み、approve (注意事項あり)** (最新: 2026-10-16 / 1.4.0、初回: 2026-09-24)

2026-10-02 に `0ae3052` で retire されたものを 2026-10-07 に復活 (user 依頼)。復元元は retire 直前版。

AUR の `antigravity-cli` PKGBUILD (pkgver=1.2.9_5905287731871744, pkgrel=1,
AUR commit `a154fb7` = 2026-09-23) を fork。
`arch=('x86_64')` のみに絞る + nvchecker 設定を `regex` source に変える
(下記「依存方針」) 以外は無改変。`LICENSE` / `antigravity-cli.install` は
AUR と byte 一致。

## Source

- AUR: https://aur.archlinux.org/packages/antigravity-cli
  - AUR 上の maintainer / submitter: `viridivn` (投票 28、初回 submit 2026-05)
  - PKGBUILD 冒頭の `# Maintainer:` コメントは `Coraline Shuryn` で AUR の
    maintainer 名と異なるが、初回 commit から一貫して同じ記載で途中変更は無い
    (= 履歴上の異変ではなく記載上の食い違い。source URL には影響しない)
- Upstream: https://antigravity.google/product/antigravity-cli (Google)
  - distribution: `storage.googleapis.com/antigravity-public/antigravity-cli/`
    (Google Cloud Storage、 GitHub 等の公開 git repo / tag は無い closed-source)
  - release 識別子: `1.2.12-5784551402897408` (= `<version>-<build id>`)

## 検証結果

- [x] `source_x86_64` URL = `https://storage.googleapis.com/antigravity-public/antigravity-cli/1.4.0-5281842794987520/linux-x64/cli_linux_x64.tar.gz`
  - upstream の auto-updater release server
    (`antigravity-cli-auto-updater-974169037036.us-central1.run.app/manifests/linux_amd64.json`)
    が返す `url` と **完全一致** (manifest の `version` は `1.4.0`)。同 bucket /
    path 構造は AUR 履歴 (2026-05 初回 submit 〜 現在) を通じて不変で、
    domain の付け替え / typosquat は無い
  - バイナリ内に埋め込まれた release server URL も同じ Cloud Run endpoint
  - GitHub `google-antigravity/antigravity-cli` の tag `1.4.0` / commit `e8a400403c82b7eb27a333784dc3c37985aa622e`、
    publisher `haofan` は従来と同一、 release は draft / prerelease でない (published 2026-10-16)
- [x] `sha256sums_x86_64` を独立実測 (`curl | sha256sum`)
  - 実測 / PKGBUILD 値 (1.4.0): `fd96967583c1eed94733c039638fa59b71ddbd823c75ebc8dd0e677463565aad`
  - **upstream が manifest で公開している sha512 と一致** (= AUR maintainer の
    値に頼らない独立照合):
    `6c4b298c2a69b6bd0f58567031cc56d8097a32be54c4b633aa80ad395871b389618ab3860d4ba0a12599831b8226b51b03e258ee2ec7fae105a45b8e0fce2188`
  - **GitHub Release `1.4.0` の asset `agy_cli_linux_x64.tar.gz` の digest とも
    完全一致** (size 63,650,174 / `sha256:fd969675...`)。旧 pin の 1.3.5 でも
    asset digest (`sha256:007a5a44...`) が本 repo の pin と一致し、GCS manifest
    配信物 = GitHub Release asset であることを複数版で確認
  - AUR は 1.3.5 に追随済み (x86_64 sha256 `007a5a44...` は本 repo の 1.3.5 pin と
    一致)。1.4.0 は AUR 未追随のため cross-check 不可
    (upstream 先行は 1.3.1〜1.3.5 時の前例あり)
  - (初回 1.2.9 時の記録) aarch64 tarball も sha256 / sha512 とも manifest と
    一致を確認済み (本 repo では使わない)
- [x] tarball の中身は `antigravity` 1 ファイル (220,553,472 byte、root/root、
  0755) のみ (1.3.5 は 215,060,688 byte、構成同一)。 ELF 64-bit x86-64 PIE、
  stripped、Go 製。 tarball 内バイナリの sha256:
  `60e72242f07c131f75afcd9a9e425d6ee097a1a1b3190ea98ae78fa4180bbfa8`
- [x] `package()`: `install -Dm755 antigravity → /usr/bin/agy` と
  `install -Dm644 LICENSE → /usr/share/licenses/antigravity-cli/LICENSE` のみ。
  curl / wget / eval / pipe-to-shell 無し。`prepare()` / `build()` 無し
- [x] `depends=('glibc')` は妥当: バイナリの `DT_NEEDED` は
  `libc / libm / libdl / libpthread / librt / libresolv / ld-linux` の glibc 系のみ、
  要求 symbol version は最大 `GLIBC_2.26` (Arch の glibc 2.44 で充足)。
  1.4.0 でも `DT_NEEDED` 集合・最大 symbol version とも 1.3.5 と同一
- [x] バイナリ内に埋め込まれた宿主名は Google 管理ドメイン
  (`antigravity.google`, `googleapis.com`, `cloud.google.com`,
  `antigravity-unleash.goog` 等) と github.com / 仕様書 URL のみ。
  不審な第三者ドメイン無し (`strings` による静的確認)。 1.4.0 では scheme 付き
  URL の host 集合に `nodejs.org` (完全新規) と `registry.npmjs.org`
  (URL target としては新規) が加わった。いずれも 1.4.0 で内蔵された
  `chrome-devtools-mcp` ランチャーが実行時に download する先で、version は
  pin 済み・公式配布元 (下記 注意事項)。 これ以外の新規 host は無い。
  `https://opentelemetry.io/schemas/1.43.0` は OpenTelemetry SDK の schema
  識別子 (属性値の文字列定数で fetch 先ではない)。 strings の連結ノイズを除いた
  実差分はこの 3 件のみ
- [x] `provides=('agy')` / `conflicts=('agy')`: Arch 公式 repo に `agy` /
  `antigravity*` は無く衝突なし。`optdepends=('antigravity')` は AUR 側の
  desktop app を指すヒントで [nekono] 外の pkg (無くても install 可)
- [x] `antigravity-cli.install`: `post_install` で `agy install` を案内する
  `echo` のみ。自動実行はしない
- [x] `options=('!strip')`: upstream 側で strip 済みの単一バイナリのため妥当
  (claude-code と同方針)
- [x] (初回 1.2.9 時の記録) ローカル test build (`makepkg -f --nosign`) 成功、makepkg の lint 警告無し。
  pkg 内 `/usr/bin/agy` の sha256 は upstream tarball 内バイナリと一致
  (`1dbb10f8295cc1ad2e558bd006c7808fe53b6c7f678a887eb557b576bb591711`)、
  `.PKGINFO` に `$srcdir` の絶対 path 混入無し

## 注意事項 (受入の上で把握しておくこと)

- [⚠] **バイナリに background auto-updater が内蔵**されている
  (`third_party/jetski/cli/updater/auto_updater.go`)。既定の更新先は上記と同じ
  Google の Cloud Run release server。 無効化用 env は
  `AGY_CLI_DISABLE_AUTO_UPDATE`。
  - `/usr/bin/agy` は root 所有なので、一般 user 権限の updater が in-place で
    上書きすることはできない
  - ただし updater が user 領域 (`~/.local` 等) に別 copy を置いて pacman 管理外の
    version が動く可能性は **未検証** (= 下記の通りバイナリを実行していないため)。
    pacman 管理下の version pin を厳密にしたい client では `AGY_CLI_DISABLE_AUTO_UPDATE=1`
    を設定すること
- [⚠] **1.4.0 で `chrome-devtools-mcp` ランチャーが内蔵された** (browser tool 系
  MCP を使用した時に走る embed bash script)。 system に `npx` + Node.js >= 20 が
  あれば `npx -y chrome-devtools-mcp@1.10.1` を exec し、無ければ nodejs.org から
  Node.js 22.20.0 (`node-v22.20.0-linux-{x64,arm64}.tar.gz`) を `/tmp` /
  `$HOME/.cache` 配下へ download・展開した上で registry.npmjs.org の
  `chrome-devtools-mcp@1.10.1` tarball を取得して実行する。 **Node.js 本体の
  download は Node.js >= 20 があれば回避されるが、 chrome-devtools-mcp の
  tarball は npx 経路も含め初回に registry.npmjs.org から取得される**
  (1.10.1 に固定。 curl + tar 経路の checksum 検証はランチャー内には無く、 npx
  経路は npm の取得処理に委ねる)。 download 先は公式配布元で不審ドメインでは
  ない。 配置先は `CHROME_DEVTOOLS_MCP_DIR` / `CHROME_DEVTOOLS_MCP_NODE_DIR` で
  変更可。 本 pkg は `/usr/bin/agy` 1 file を install するのみで、 build / install
  時の動的取得は無い
- [⚠] **バイナリは実行していない**。build host は YubiKey の gpg-agent を SSH
  forward で受けているため、出所を独立検証できていても未実行の closed-source
  バイナリ (現行 1.4.0 で 220,553,472 byte ≒ 210 MiB) をここで走らせない方針。
  上記は静的解析 (ELF header /
  `readelf` / `strings`) と hash 照合に基づく。 `agy install` が shell 環境に
  何を書くかも未解析 (= ユーザが明示的に叩く操作で、pkg は自動実行しない)
- [⚠] license は `custom:proprietary` (Google 所有、 AUR 同梱 LICENSE は
  packaging script 部分のみ 0BSD で、コンパイル済みバイナリには OSS license を
  与えない旨)。再配布許諾の明記は無いが、[nekono] は Tailscale 内の私設 repo で、
  同種の proprietary binary を配る `claude-code` と同じ扱い

## 依存方針 (AUR との意図的 diff)

- `arch=('x86_64')` のみに削減 (AUR 本家は aarch64 も persist)。 [nekono] は
  `repo/x86_64/` のみ配信するため `source_aarch64` / `sha256sums_aarch64` block を
  削除 (electron37-bin と同方針)。 bump 時に AUR の値をコピーする場合は aarch64
  行を持ち込まないこと
- `nvchecker.toml` は AUR の `.nvchecker.toml` (`source = "jq"`) を使わず
  `source = "regex"` にした。 nvchecker の jq source は Python module `jq`
  (Arch: `python-jq`) を import するが、 `upstream-version-issue.yml` はこれを
  install せず、実機 (nvchecker 2.22) で `No module named 'jq'` を確認。
  regex source は追加依存が要らず、同じ manifest から
  `1.2.9_5905287731871744` (= PKGBUILD の `pkgver` と文字列完全一致) を返すことを
  実機で確認済み。 `detect_upstream_updates.py` は pkgver 文字列を直接比較するので
  完全一致が必須

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可 (`bin/build-all antigravity-cli`)。

改変 step は無く、実行 binary は upstream (Google) の release そのまま。
信頼境界に入れるのは「Google の release pipeline が出す prebuilt binary」で、
`claude-code` と同じ構造。 auto-updater の扱い (上記注意事項) だけは client
運用側で意識すること。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-09-24 | 1.2.9_5905287731871744-1 | (this commit) | release `1.2.9-5905287731871744` (Google GCS、公開 git tag 無し。upstream manifest の url と一致) | 初回 add。AUR `antigravity-cli` (commit `a154fb7`) を fork、`arch=('x86_64')` に絞る。 sha256 を独立実測し AUR 値・upstream 公開 sha512 と一致確認。 ELF `DT_NEEDED` が glibc のみで depends 妥当を確認。 nvchecker は jq source が workflow で動かないため regex source に変更。 auto-updater 内蔵 (`AGY_CLI_DISABLE_AUTO_UPDATE` で無効化可) を注意事項に記録。 |
| 2026-09-27 | 1.2.11_6016716732497920-1 | (this PR) | release `1.2.11-6016716732497920` (Google GCS、公開 git tag 無し。upstream manifest の url と一致) | Issue #654 (1.2.10) / #656 (1.2.11) を最新 1.2.11 に集約、approve。 sha256 を独立実測 (`c91c62c5...`) し、upstream manifest 公開の sha512 と一致 (AUR は 1.2.10 止まりで 1.2.11 未反映のため AUR 値とは照合不可。集約対象の 1.2.10 tarball は実測して AUR 値 `77cb6925...` と一致し、版の系列が公式 bucket 由来であることを確認)。 tarball は `antigravity` 1 ファイルのまま (219,545,808 byte)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.2.9 と同一で depends 変更不要。 埋め込み URL 集合 (277 件) 同一で新規接続先なし。 AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、 `LICENSE` / `antigravity-cli.install` は byte 一致。 バイナリは未実行 (静的解析のみ)。 |
| 2026-09-28 | 1.2.12_5784551402897408-1 | (this PR) | release `1.2.12-5784551402897408` (Google GCS、公開 git tag 無し。upstream manifest の url と一致) | Issue #667。 sha256 を独立実測 (`26c7c4c6...`) し、upstream manifest 公開 sha512 と一致 (AUR も 1.2.12 に追随済みで、AUR 記載 sha256 とも一致)。 tarball は `antigravity` 1 ファイルのまま (219,959,504 byte)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.2.11 と同一で depends 変更不要。 埋め込み URL の新規追加は Google 管理ドメインのみで、第三者ドメインの新規接続先なし。 AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、 `LICENSE` / `antigravity-cli.install` は byte 一致。 バイナリは未実行 (静的解析のみ)。 |
| 2026-09-29 | 1.2.12_5784551402897408-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): glibc 2.44+r24+g16be1518495f-1 → 2.44+r50+g1848099f063e-1 |
| 2026-10-07 | 1.3.1_4582356770750464-1 | (this PR) | release `1.3.1-4582356770750464` (Google GCS、upstream manifest の url と一致) | 2026-10-02 retire (`0ae3052`) からの復活、最新 1.3.1 へ bump。sha256 を独立実測 (`0e313b30...`) し manifest 公開 sha512 (`3b8349d7...`) と一致。tarball は `antigravity` 1 ファイル (211,026,152 byte)、`DT_NEEDED` は glibc 系のみ・最大 `GLIBC_2.26` で depends 変更不要。Arch 公式に `agy` なし、`.deps.lock` の glibc は現行と同一。AUR は 1.3.0 止まりで cross-check 不可 (upstream 先行は前例あり)。バイナリ未実行 (静的解析のみ)。 |
| 2026-10-10 | 1.3.2_5813501495738368-1 | (this PR) | release `1.3.2-5813501495738368` (Google GCS、upstream manifest の url と一致) | Issue #700。sha256 を独立実測 (`bf8504c7...`) し manifest 公開 sha512 (`8dc2bb84...`) と一致。tarball は `antigravity` 1 ファイルのまま (211,841,256 byte、1.3.1 比 +815KB)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.3.1 と同一で depends 変更不要。埋め込み URL 集合は 137 host で同一、差分は Google 公式 styleguide の docs URL (`google.github.io`、prompt 文中の参照) のみで第三者ドメインの新規接続先なし。AUR は 1.3.1 止まりで cross-check 不可 (upstream 先行は 1.3.1 時の前例あり、AUR 側の x86_64 sha256 `0e313b30...` は本 repo の 1.3.1 pin と一致)。`package()` / `.install` / `LICENSE` 無変更。バイナリは未実行 (静的解析のみ)。 |
| 2026-10-11 | 1.3.3_5524738307653632-1 | (this PR) | release `1.3.3-5524738307653632` (Google GCS、upstream manifest の url と一致。GitHub `google-antigravity/antigravity-cli` tag `1.3.3` / commit `800852e`、publisher `haofan` は従来と同一) | Issue #705。sha256 を独立実測 (`8ee3ca32...`) し manifest 公開 sha512 (`abe3ddf5...`) と一致 (AUR は 1.3.2 止まりで cross-check 不可、upstream 先行は 1.3.1/1.3.2 時の前例あり。AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、`LICENSE` / `antigravity-cli.install` は byte 一致)。tarball は `antigravity` 1 ファイルのまま (root/root 0755、214,237,392 byte、1.3.2 比 +2.4MB、Go 製 ELF x86-64 PIE stripped)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.3.2 と同一で depends 変更不要。埋め込み URL の新規 host は `agent-plugins.org` のみで、同 release notes の plugin `agents.json` 対応に使う仕様策定サイト (Agent Plugins 1.0 open spec、JSON schema `$id` 参照) の正規ドメイン。`api.github.com` / `169.254.169.254` は旧版にも存在 (strings 断片の接着差のみ)。release notes (機能追加: pinned prompt header、MCP alwaysLoad、plugin agents.json、macOS sandbox 強化等、他 bugfix) に packaging / install 経路の変更なし。`package()` / `.install` / `LICENSE` 無変更。バイナリは未実行 (静的解析のみ)。 |
| 2026-10-13 | 1.3.4_5524158944247808-1 | (this PR) | release `1.3.4-5524158944247808` (Google GCS、upstream manifest の `version` / `url` と一致。GitHub `google-antigravity/antigravity-cli` tag `1.3.4` / commit `9eebbed`、publisher `haofan` は従来と同一、draft / prerelease でない) | Issue #710。sha256 を独立実測 (`db5e81e8...`) し manifest 公開 sha512 (`dbdacfc5...`) と一致。加えて GitHub Release asset `agy_cli_linux_x64.tar.gz` の digest (`sha256:db5e81e8...`、size 62,750,975) とも完全一致 (旧 1.3.3 でも GCS から再取得した実測値 = asset digest `sha256:8ee3ca32...` = pin で、GCS 配信物と GitHub Release asset が同一物であることを両版で確認)。tarball は `antigravity` 1 ファイルのまま (root/root 0755、214,745,296 byte、1.3.3 比 +508KB、Go 製 ELF x86-64 PIE stripped)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.3.3 と同一で depends 変更不要。埋め込み URL host 集合 (139 host) も 1.3.3 と同一で新規接続先なし。GitHub の tag 間 diff は CHANGELOG.md のみ。release notes は plugin list の `--json` 対応 / model picker / Remote Control 改善と bugfix のみで packaging / install 経路の変更なし。`package()` / `.install` / `LICENSE` 無変更。AUR は 1.3.2 止まりで cross-check 不可 (upstream 先行は 1.3.1〜1.3.3 時の前例あり、AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、`LICENSE` / `antigravity-cli.install` は byte 一致)。「検証結果」 section が 1.2.12 時の値のまま stale だったので 1.3.4 の値に更新。バイナリは未実行 (静的解析のみ)。 |
| 2026-10-14 | 1.3.5_6164039479459840-1 | (this PR) | release `1.3.5-6164039479459840` (Google GCS、upstream manifest の `version` / `url` と一致。GitHub `google-antigravity/antigravity-cli` tag `1.3.5` / commit `89081f72e10a98c3689f34b1af7261f43136ab5d`、publisher `haofan` は従来と同一、draft / prerelease でない) | Issue #715。sha256 を独立実測 (`007a5a44...`) し manifest 公開 sha512 (`0fa5ff7c...`) と一致。加えて GitHub Release asset `agy_cli_linux_x64.tar.gz` の digest (`sha256:007a5a44...`、size 62,830,042) とも完全一致 (1.3.4 の asset digest `sha256:db5e81e8...` も本 repo pin と一致済み)。tarball は `antigravity` 1 ファイルのまま (root/root 0755、215,060,688 byte、1.3.4 比 +315KB、Go 製 ELF x86-64 PIE stripped)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.3.4 と同一で depends 変更不要。埋め込み URL host 集合 (139 host) も 1.3.4 と同一で新規接続先なし。GitHub の tag 間 diff は CHANGELOG.md のみ。release notes は Auto mode 追加 / `/compact` コマンド追加 / `/diff` の syntax highlight・folding・性能改善・iTerm2 色修正のみで packaging / install 経路の変更なし。`package()` / `.install` / `LICENSE` 無変更。AUR は 1.3.4 に追随済み (x86_64 sha256 `db5e81e8...` = 本 repo の 1.3.4 pin と一致)、1.3.5 は未追随で cross-check 不可 (upstream 先行は前例あり)。AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、`LICENSE` / `antigravity-cli.install` は byte 一致。バイナリは未実行 (静的解析のみ)。 |
| 2026-10-16 | 1.4.0_5281842794987520-1 | (this PR) | release `1.4.0-5281842794987520` (Google GCS、upstream manifest の `version` / `url` と一致。GitHub `google-antigravity/antigravity-cli` tag `1.4.0` / commit `e8a400403c82b7eb27a333784dc3c37985aa622e`、publisher `haofan` は従来と同一、draft / prerelease でない) | Issue #724 (`Closes #724`)。sha256 を独立実測 (`fd969675...`) し manifest 公開 sha512 (`6c4b298c...`) と一致。加えて GitHub Release asset `agy_cli_linux_x64.tar.gz` の digest (`sha256:fd969675...`、size 63,650,174) とも完全一致 (1.3.5 の asset digest `sha256:007a5a44...` も本 repo pin と一致済み)。tarball は `antigravity` 1 ファイルのまま (root/root 0755、220,553,472 byte、1.3.5 比 +5.5MB)、`DT_NEEDED` 集合・最大 symbol version `GLIBC_2.26` は 1.3.5 と同一で depends 変更不要。埋め込み URL host の実差分は `nodejs.org` / `registry.npmjs.org` の 2 host (+ `https://opentelemetry.io/schemas/1.43.0` は schema 識別子で fetch 先ではない) のみで、`nodejs.org` / `registry.npmjs.org` は 1.4.0 で内蔵された `chrome-devtools-mcp` ランチャー (実行時 bootstrap、version pin 済み、公式配布元) の download 先 (注意事項に追記)。GitHub の tag 間 diff は CHANGELOG.md のみ (+28)。release notes は OSC 7501 terminal status / `alt+m` shortcut / `agy plugin validate` 改善 / plugin MCP header fix / `/diff` commit tree 修正等で packaging / install 経路の変更なし。`package()` / `.install` / `LICENSE` 無変更 (LICENSE sha256 `7bcdb3cf...` も不変)。AUR は 1.3.5 止まりで 1.4.0 は cross-check 不可 (upstream 先行は前例あり)。AUR の PKGBUILD は `arch` 行 (aarch64) 以外が本 repo と同一、`LICENSE` / `antigravity-cli.install` は byte 一致。バイナリは未実行 (静的解析のみ)。 |

## 更新方針

upstream の新 release が出たら (`upstream-version-issue.yml` が Issue を立てる):
1. manifest (`.../manifests/linux_amd64.json`) の `url` / `sha512` を確認
2. 本 dir の PKGBUILD の `pkgver` (= `<version>_<build id>`) と
   `sha256sums_x86_64` を更新 (aarch64 行は持ち込まない)
3. tarball を独立に取得し sha256 を実測、manifest の sha512 とも照合
4. tarball の中身が `antigravity` 1 ファイルのままか、`DT_NEEDED` が glibc のみ
   のままか (= `depends` の変更要否) を確認
5. `.SRCINFO` 再生成、REVIEW.md に 1 行追記
