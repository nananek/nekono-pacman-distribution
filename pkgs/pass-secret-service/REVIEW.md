# pass-secret-service review

## 状態

**review 済み、approve** (2026-05-19)

AUR の `pass-secret-service` PKGBUILD (pkgver=0.7.0, pkgrel=1) を fork。
意図的改変 2 件あり (後述)。

## Source

- AUR: https://aur.archlinux.org/packages/pass-secret-service
  - Maintainer: nathawat
  - Contributor: jakka
- Upstream: https://github.com/grimsteel/pass-secret-service
  - grimsteel (Steed Mountjoy)、GPL-3.0-or-later、Rust 実装

## 検証結果

- [x] source URL = `github.com/grimsteel/pass-secret-service/archive/v0.7.0.tar.gz`
  - upstream 公式 GitHub archive、typosquatting なし
- [x] b2sums が upstream tarball と一致
  - 実測: `e9c5aae7fcfc348d092cb25f98821aac98fe17e746f89e5f0e93412409d887c6bbbc0076badd6e9e11858bb68abeba97a00cf52f99597e2bbd02e804d01ad9b8`
  - PKGBUILD 値: 同上 (一致)
  - sha256 (参考): `2647721628c48eee4cdf472d35ea341397c5720f7b5426a16ae9466e9993ab23`
- [x] v0.7.0 tag commit SHA: `3f9e2925b5a8d9fc81a8095b01dc16b5b663edd1`
- [x] prepare(): `cargo fetch --locked` で Cargo.lock 固定の依存クレートを取得 (= ネットワークアクセスあり、 ただし `--locked` で pinned バージョン以外の解決を禁止、 Cargo.lock 自体は b2sums 検証済み tarball 同梱で固定)
- [x] build(): `cargo build --frozen` でオフラインビルド (= 追加 fetch 不可)
- [x] package(): `install` のみ、 ネットワーク操作・eval・injection なし
- [x] depends: gcc-libs / glibc / dbus / pass — 全て妥当
- [x] makedepends: cargo — Rust ビルドツール (= `rust` package の provides、 公式 repo に独立 package なし)
- [x] PKGBUILD 本 repo commit SHA: `2b50d97` (= 初回 add commit、 以降 .SRCINFO / REVIEW.md / .deps.lock 追加は同 PKGBUILD 不変)

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `replace=` → `replaces=` | AUR 側の typo 修正 (makepkg の正しい変数名) |
| D-Bus service ファイル perms: 744 → 644 | .service activation ファイルに execute bit 不要 |
| `package()` 内 `${_pkgname}` → `${pkgname}` | AUR 側の未定義変数 bug 修正 (`_pkgname` は PKGBUILD 内で未定義、 空文字に展開されてインストールパスが崩れる) |
| `package()` の `_docfiles` / `_licensefiles` ループ廃止 → 直接 `install` 1 行 | AUR 側 boilerplate の簡略化 (機能等価、 同じ install 先 + 同じ perms) |

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可。
