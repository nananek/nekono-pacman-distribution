# antigravity-cli review

## 状態

**review 済み、approve (注意事項あり)** (2026-09-24)

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
  - release 識別子: `1.2.9-5905287731871744` (= `<version>-<build id>`)

## 検証結果

- [x] `source_x86_64` URL = `https://storage.googleapis.com/antigravity-public/antigravity-cli/1.2.9-5905287731871744/linux-x64/cli_linux_x64.tar.gz`
  - upstream の auto-updater release server
    (`antigravity-cli-auto-updater-974169037036.us-central1.run.app/manifests/linux_amd64.json`)
    が返す `url` と **完全一致**。同 bucket / path 構造は AUR 履歴 (2026-05 初回
    submit 〜 現在) を通じて不変で、domain の付け替え / typosquat は無い
  - バイナリ内に埋め込まれた release server URL も同じ Cloud Run endpoint
- [x] `sha256sums_x86_64` を独立実測 (`curl | sha256sum`) し AUR 記載値と一致
  - 実測 / PKGBUILD 値: `d9850373f3df866011024a961fa9740cc4adaac060eebe9c70fbf263ac6b2624`
  - **upstream が manifest で公開している sha512 とも一致** (= AUR maintainer の
    値を鵜呑みにしない独立照合):
    `75e83d46cb1632da4aca29b785017753c8c2dfb0320166da6b2c8007678982943515e64c8824ff2d8510d9b52eb2b3c27f306aeec58c4a4f4f8aa5550cb0cb12`
  - (参考) 本 repo では使わない aarch64 tarball も sha256 / sha512 とも
    manifest と一致を確認済み
- [x] tarball の中身は `antigravity` 1 ファイル (217,424,080 byte、root/root、
  0755) のみ。 ELF 64-bit x86-64 PIE、stripped、Go 製
- [x] `package()`: `install -Dm755 antigravity → /usr/bin/agy` と
  `install -Dm644 LICENSE → /usr/share/licenses/antigravity-cli/LICENSE` のみ。
  curl / wget / eval / pipe-to-shell 無し。`prepare()` / `build()` 無し
- [x] `depends=('glibc')` は妥当: バイナリの `DT_NEEDED` は
  `libc / libm / libdl / libpthread / librt / libresolv / ld-linux` の glibc 系のみ、
  要求 symbol version は最大 `GLIBC_2.26` (Arch の glibc 2.44 で充足)
- [x] バイナリ内に埋め込まれた宿主名は Google 管理ドメイン
  (`antigravity.google`, `googleapis.com`, `cloud.google.com`,
  `antigravity-unleash.goog` 等) と github.com / 仕様書 URL のみ。
  不審な第三者ドメイン無し (`strings` による静的確認)
- [x] `provides=('agy')` / `conflicts=('agy')`: Arch 公式 repo に `agy` /
  `antigravity*` は無く衝突なし。`optdepends=('antigravity')` は AUR 側の
  desktop app を指すヒントで [nekono] 外の pkg (無くても install 可)
- [x] `antigravity-cli.install`: `post_install` で `agy install` を案内する
  `echo` のみ。自動実行はしない
- [x] `options=('!strip')`: upstream 側で strip 済みの単一バイナリのため妥当
  (claude-code と同方針)
- [x] ローカル test build (`makepkg -f --nosign`) 成功、makepkg の lint 警告無し。
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

## 更新方針

upstream の新 release が出たら (`upstream-version-issue.yml` が Issue を立てる):
1. manifest (`.../manifests/linux_amd64.json`) の `url` / `sha512` を確認
2. 本 dir の PKGBUILD の `pkgver` (= `<version>_<build id>`) と
   `sha256sums_x86_64` を更新 (aarch64 行は持ち込まない)
3. tarball を独立に取得し sha256 を実測、manifest の sha512 とも照合
4. tarball の中身が `antigravity` 1 ファイルのままか、`DT_NEEDED` が glibc のみ
   のままか (= `depends` の変更要否) を確認
5. `.SRCINFO` 再生成、REVIEW.md に 1 行追記
