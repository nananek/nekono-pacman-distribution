# pika-backup review

## 状態

**review 済み、approve** (2026-05-16、0.8.2 に bump)

AUR の `pika-backup` PKGBUILD (pkgver=0.7.4) を fork したが、AUR
maintainer が **0.7.4 で stuck** している。最近の Arch rust が rust 2024
edition の `never type fallback` を hard error 化したため、0.7.4 (edition
2021 で書かれていて互換性なし) は build 不能。upstream **0.8.2** (= 2026-04-13
release、`edition = "2024"` 明示済み) に bump して問題を解消する。

**改変 2 箇所** (AUR からの divergence):
1. `pkgver=0.7.4` → `0.8.2`
2. `source=("git+...#tag=v$pkgver")` → `tag=$pkgver` (= upstream tag は
   `v` prefix なしの `0.8.2`、`v0.8.2` は存在しない)
3. `sha256sums` を `SKIP` に (= git+ source 用、Arch Wiki ガイドライン準拠)

## Source

- AUR (参考): https://aur.archlinux.org/packages/pika-backup (= 0.7.4 で stuck)
- Upstream: https://gitlab.gnome.org/World/pika-backup
  - GNOME World group の official sub-project (= 公認 third-party GNOME app)
  - 主開発者 Sophie Herold、Fina Wilke (Cargo.toml authors)
- Homepage: https://apps.gnome.org/PikaBackup

## 0.7.4 → 0.8.2 の主な変更点 (Cargo.toml workspace.package より)

- **edition = "2024"** (= rust 2024 native、deprecated dependency_on_unit_never_type_fallback 対応)
- **rust-version = "1.92"** (= Arch の rust 1.93+ で OK)
- **resolver = "3"** (cargo dependency resolver、新しい仕様)
- 既存 dep 要件 update:
  - gtk4 >= 4.12.5
  - libadwaita-1 >= 1.7
  - meson >= 0.57
- 開発者 `Fina Wilke` が co-maintainer に追加

## 検証結果

- [x] `source` URL = `git+https://gitlab.gnome.org/World/pika-backup.git#tag=0.8.2`
  - GitLab gnome 公式インスタンス、`World/pika-backup` group は GNOME 公認
- [x] Tag `0.8.2` の commit (`5325daffa11b6de4e5b6b0e5bce82bc5431b2e89`) は
      author Sophie Herold、2026-04-13 release、title "utils/oo7_workaround: Add proper workaround"
- [⚠] Tag に GPG signature **無し** (GNOME 一般運用)、GitLab TLS + git
      history で integrity 確保
- [x] `sha256sums=('SKIP')` — git+ source 用、Arch Wiki 準拠
- [x] `prepare()`: `cargo fetch --target ...` で Cargo.lock に基づく
      crates.io fetch (= 既存と同じ、build 中 network access 必要)
- [x] `build()`: `arch-meson + meson compile` 標準
- [x] `check()`: `desktop-file-validate` + `appstreamcli validate` のみ、
      network なし
- [x] `package()`: `meson install -C build --no-rebuild --destdir "$pkgdir"`
- [x] `depends`: `borg`, `fuse3`, `libadwaita`, `libsecret`, `python-pyfuse3`
      — pika-backup の GUI フロントとして妥当 (0.7.4 と同)
- [x] `makedepends`: `cargo`, `git`, `itstool`, `meson` — Rust + meson 標準
- [x] license `GPL-3.0-or-later` — upstream Cargo.toml と一致

## 結論

**approve** — build host で `bin/build-all pika-backup` 実行。0.7.4 build
失敗の root cause (= rust 2024 hard error) は 0.8.2 で解消されるはず。

## 更新方針

upstream の新 release (0.8.3 等) が出たら:
1. `pkgver` を bump
2. tag commit を GitLab API で confirm
3. Cargo.lock / Cargo.toml の `edition` / `rust-version` 更新を確認
4. `gtk4` / `libadwaita-1` の version 要件が新 release で上がってないか確認
5. REVIEW.md に確認日 + 結論 update

将来 AUR maintainer が 0.8.x に追随したら、AUR PKGBUILD と diff を取り直して
本 fork の divergence を縮める方向を検討。
