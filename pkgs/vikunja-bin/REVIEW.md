# vikunja-bin review

## 状態

**review 済み、approve** (2026-08-19)

AUR の `vikunja-bin` PKGBUILD (pkgver=2.5.0, pkgrel=1) を fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/vikunja-bin
  - maintainer: Michael Clayfield
- Upstream: https://github.com/go-vikunja/vikunja (= Vikunja organization、
  AGPL-3.0、self-hosted To-Do list / task manager、13k+ stars)
- 配布 CDN: https://dl.vikunja.io/ (= vikunja.io/install/ からリンクされている
  公式 download server、typosquat / mirror spoof なし)

## 検証結果

- [x] `source_x86_64` / `source_armv7h` / `source_aarch64` URL =
      `dl.vikunja.io/vikunja/v${pkgver}/vikunja-v${pkgver}-linux-{amd64,arm-7,arm64}-full.zip`
  - upstream 公式 download server、typosquatting / mirror spoof なし
- [x] `sha256sums_x86_64` / `sha256sums_armv7h` / `sha256sums_aarch64` が
      実測値と一致
  - `vikunja-v2.5.0-linux-amd64-full.zip`:
    `8843de18f5f297bac83db010a54064a45033f82cffdf53421f6ce39f12a8ad98`
  - `vikunja-v2.5.0-linux-arm-7-full.zip`:
    `c98905c277baca10092c15ef65c1f0c02c0da4cdd63a2ac1ef04f2a3e7115151`
  - `vikunja-v2.5.0-linux-arm64-full.zip`:
    `299b90a2b9c5a54a2901f0c585f91cdc48042d59d2355ef5d058d6f27c5af45a`
  - すべて curl + sha256sum で独立に再計算し PKGBUILD 値と一致確認済み
- [x] 自前 file (`vikunja.service` / `vikunja.sysusers` / `vikunja.tmpfiles`)
      の sha256 も実測し PKGBUILD の (arch 非依存) `sha256sums` 値と一致確認
- [x] tag `v2.5.0` の git commit (`ef2200e9429c5cc42f5c1811433418bfcc72b3aa`)
      は **GPG verified** (author: kolaente = Vikunja 主開発者)、tampering
      の兆候なし。GitHub Releases API 上の latest release とも一致
      (= 現時点で最新版)
- [x] `package()`: `install -Dm755` / `install -Dm640` / `install -Dm644` の
      標準コマンドのみ。network fetch / eval / curl / pip 等の動的取得なし
  - `/usr/bin/vikunja`: release zip 内の prebuilt Go 静的バイナリ (CARCH 別
    に amd64/arm-7/arm64 の 3 種を分岐、非決定 build ではなく単純な
    arch dispatch)
  - `/usr/lib/systemd/system/vikunja.service`: `vikunja.service` を vendor
    (hardening directive 多数、`ProtectSystem=full` 等で妥当な構成)
  - `/etc/vikunja/config.yml.sample`: upstream zip 同梱の sample config、
    そのまま install のみ (実 config ではない)
  - `/usr/lib/sysusers.d/vikunja.conf`: `vikunja.sysusers` を vendor
    (system user `vikunja` を作成するのみ)
  - `/usr/lib/tmpfiles.d/vikunja.conf`: `vikunja.tmpfiles` を vendor
    (`/var/lib/vikunja` を 0750 vikunja:vikunja で作成)
- [x] `depends`: 未指定 (= Go 静的バイナリで実行時 runtime dep 不要、
      `ldd` 相当の確認は zip 展開後の実行ファイルサイズ (~60MB) からも
      static build であることと整合)
- [x] `makedepends` / `checkdepends`: 未指定、build() 自体が存在しない
      (prebuilt binary の install のみ)
- [x] license `AGPL3` — upstream 一致

## 結論

**approve** — そのまま build host で
`makepkg -s --sign --key 483DC691DF9F29327EA106BD030130E2F156CD74` 可。

self-hosted service 系 (systemd unit + sysusers + tmpfiles) の構成だが、
build/install step は upstream 公式 CDN からの prebuilt binary 展開のみで
動的取得・任意コード実行の余地なし。運用開始する場合は
`/etc/vikunja/config.yml` を sample から作成し DB 接続先等を設定した上で
`systemctl enable --now vikunja` する運用が必要 (= package 自体は enable
しない、AUR 標準の作法に準拠)。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-08-19 | 2.5.0-1 | (this PR) | `ef2200e9429c5cc42f5c1811433418bfcc72b3aa` | 初回 review、AUR からの改変なしで受入 |

## 更新方針

upstream の新 release (v2.5.1 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立に再計算 (= `curl -fsSL <url> | sha256sum` で照合、3 arch 分)
4. tag commit が GPG verified か `gh api repos/go-vikunja/vikunja/commits/<tag>`
   で確認
5. REVIEW.md に確認日 + 結論 update
