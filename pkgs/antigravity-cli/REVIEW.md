# antigravity-cli review

## 状態

**review 済み、approve (注意事項あり)** (最新: 2026-09-27 / 1.2.11、初回: 2026-09-24)

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
  - release 識別子: `1.2.11-6016716732497920` (= `<version>-<build id>`)

## 検証結果

- [x] `source_x86_64` URL = `https://storage.googleapis.com/antigravity-public/antigravity-cli/1.2.11-6016716732497920/linux-x64/cli_linux_x64.tar.gz`
  - upstream の auto-updater release server
    (`antigravity-cli-auto-updater-974169037036.us-central1.run.app/manifests/linux_amd64.json`)
    が返す `url` と **完全一致**。同 bucket / path 構造は AUR 履歴 (2026-05 初回
    submit 〜 現在) を通じて不変で、domain の付け替え / typosquat は無い
  - バイナリ内に埋め込まれた release server URL も同じ Cloud Run endpoint
- [x] `sha256sums_x86_64` を独立実測 (`curl | sha256sum`)
  - 実測 / PKGBUILD 値 (1.2.11): `c91c62c5e6fa954f5a7e1d7b9ad417d749db4aa60a4ba0b3d604dec1b645d190`
  - **upstream が manifest で公開している sha512 と一致** (= AUR maintainer の
    値に頼らない独立照合。AUR は 1.2.10 止まりで 1.2.11 は未反映のため、
    1.2.11 については AUR 値との照合は不可):
    `ca12c262343f29a2b87423d1ff1e4244989e936e37fe1f8e56056b0f91cd02f93f133321229ce35c86c2d84b977e919937a3fd9430cfd769a16d6b03ede25081`
  - (参考) 集約対象の 1.2.10 tarball も実測し、AUR 記載値
    `77cb6925...` と一致 (= 版の系列が upstream 公式 bucket 由来であることの確認)
  - (初回 1.2.9 時の記録) aarch64 tarball も sha256 / sha512 とも manifest と
    一致を確認済み (本 repo では使わない)
- [x] tarball の中身は `antigravity` 1 ファイル (219,545,808 byte、root/root、
  0755) のみ (1.2.9 は 217,424,080 byte、構成同一)。 ELF 64-bit x86-64 PIE、
  stripped、Go 製。 tarball 内バイナリの sha256:
  `ec7cf797ecb0e1d91ddf3b6d9d6c1d616bb89f78a5b0e43536b72a7fce695f56`
- [x] `package()`: `install -Dm755 antigravity → /usr/bin/agy` と
  `install -Dm644 LICENSE → /usr/share/licenses/antigravity-cli/LICENSE` のみ。
  curl / wget / eval / pipe-to-shell 無し。`prepare()` / `build()` 無し
- [x] `depends=('glibc')` は妥当: バイナリの `DT_NEEDED` は
  `libc / libm / libdl / libpthread / librt / libresolv / ld-linux` の glibc 系のみ、
  要求 symbol version は最大 `GLIBC_2.26` (Arch の glibc 2.44 で充足)。
  1.2.11 でも `DT_NEEDED` 集合・最大 symbol version とも 1.2.9 と同一
- [x] バイナリ内に埋め込まれた宿主名は Google 管理ドメイン
  (`antigravity.google`, `googleapis.com`, `cloud.google.com`,
  `antigravity-unleash.goog` 等) と github.com / 仕様書 URL のみ。
  不審な第三者ドメイン無し (`strings` による静的確認)。 1.2.11 でも scheme 付き URL の
  集合 (277 件) は 1.2.9 と同一で、新規の接続先は無い (bare な宿主名で新規に見えるものは
  Go の symbol 名 `codeassistclient.apply...` / `render.commandSummaryItem` や
  Google 内部の build 元 path の断片で、ネットワーク接続先ではない)
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
- [⚠] **バイナリは実行していない**。build host は YubiKey の gpg-agent を SSH
  forward で受けているため、出所を独立検証できていても未実行の 217MB
  closed-source バイナリをここで走らせない方針。上記は静的解析 (ELF header /
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

## 更新方針

upstream の新 release が出たら (`upstream-version-issue.yml` が Issue を立てる):
1. manifest (`.../manifests/linux_amd64.json`) の `url` / `sha512` を確認
2. 本 dir の PKGBUILD の `pkgver` (= `<version>_<build id>`) と
   `sha256sums_x86_64` を更新 (aarch64 行は持ち込まない)
3. tarball を独立に取得し sha256 を実測、manifest の sha512 とも照合
4. tarball の中身が `antigravity` 1 ファイルのままか、`DT_NEEDED` が glibc のみ
   のままか (= `depends` の変更要否) を確認
5. `.SRCINFO` 再生成、REVIEW.md に 1 行追記
