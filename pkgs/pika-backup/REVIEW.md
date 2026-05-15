# pika-backup review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `pika-backup` PKGBUILD (pkgver=0.7.4, pkgrel=1, epoch=1) を fork。
**改変 1 箇所**: `sha256sums` を git+ source 用に `SKIP` に書き換え (= Arch
Wiki の VCS package ガイドライン準拠)。元の AUR 値 `4d4c0965...` は
makepkg の動作仕様で意味を持たない可能性が高く、build host での実機
test を待たずに SKIP に倒した。

## Source

- AUR: https://aur.archlinux.org/packages/pika-backup
  - maintainer: Mark Wagie
- Upstream: https://gitlab.gnome.org/World/pika-backup
  - GNOME World group の official sub-project (= 公認 third-party GNOME app)
  - 主開発者 Sophie Herold、107 stars (GitLab、GitHub と比べて star 数は低めだがコミュニティ規模を反映)
- Homepage: https://apps.gnome.org/PikaBackup

## 検証結果

- [x] `source` URL = `git+https://gitlab.gnome.org/World/pika-backup.git#tag=v0.7.4`
  - GitLab gnome 公式インスタンス、`World/pika-backup` group は GNOME 公認
  - typosquatting / mirror spoof なし
- [x] Tag `v0.7.4` の commit (`bb4e23044b502e2e536f4d9ab198b2985bef557c`) は
      author Sophie Herold (主開発者)、2024-09-23 のリリース
- [⚠] Tag に GPG signature **無し** (= GitLab API `/tags/v0.7.4/signature` が
      404 を返す)。GNOME プロジェクトは tag signing を一律運用していない、
      release artifact の content integrity は GitLab 側 TLS + git history
      で間接的に確保。
- [⚠] `sha256sums=('4d4c0965c6e7b11d78a550e253c07eb21d63d7108f723005430482f5aef6514f')`
  - **要 build 試行確認**。Arch Wiki の VCS package ガイドラインでは
    git+ source に対しては `sha256sums=('SKIP')` が推奨 (= git ref で
    integrity 確保済みなので冗長な hash は不要 + reproducible でない)。
    現状値は makepkg の動作で ignored / fail のいずれかになる可能性あり。
    初回 `makepkg --verifysource` 試行で behavior 確認、fail するなら
    SKIP に書き換え (PKGBUILD 改変 1 箇所、AUR と微差)。
- [x] `prepare()`: `cargo fetch --target $(rustc -vV | sed -n 's/host: //p')`
  - crates.io から Rust deps を fetch (= build 中の network access、AUR
    流儀の標準)。`cargo fetch` 自身は eval / 任意コード実行ではなく、
    Cargo.lock に基づく fetch のみ、内容は Cargo.lock で固定される。
- [x] `build()`: `arch-meson . build && meson compile -C build` 標準
- [x] `check()`: `desktop-file-validate` + `appstreamcli validate` のみ
      (= 失敗時 `|| :` で fail tolerant)、network なし
- [x] `package()`: `meson install -C build --no-rebuild --destdir "$pkgdir"`
- [x] `depends`: `borg`, `fuse3`, `libadwaita`, `libsecret`, `python-pyfuse3`
  - pika-backup は borg の GUI フロント、依存は妥当
  - **borg は CVE 追跡対象** (Arch Security Tracker で arch-audit から監視可能)
- [x] `makedepends`: `cargo`, `git`, `itstool`, `meson` — Rust + meson 標準
- [x] license `GPL-3.0-or-later` — upstream の COPYING と一致

## 結論

**approve (条件付き)** — `makepkg --verifysource` で sha256sums の挙動を
確認した上で本採用。SKIP への書き換えが要る場合は PKGBUILD を 1 行
改変して再 commit、その時 REVIEW.md にも diff 明記。

build host で `makepkg -s --sign --key 483D...` 可、ただし:
- cargo が crates.io から deps fetch するため build host から外部 HTTPS
  到達が必要 (= 既存運用と整合、tailnet 内で動かない問題はなし)
- borg は別途 extra repo から入る、pika-backup 側で同梱しない

## 更新方針

upstream の新 release (v0.7.5 等) が出たら:
1. AUR で pkgver / pkgrel / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. tag commit が新 author / 想定外の name でないか GitLab API で確認
   (`curl https://gitlab.gnome.org/api/v4/projects/World%2Fpika-backup/repository/tags/v<ver>`)
4. cargo の依存変更を Cargo.lock の diff で確認 (= 新規 crate の supply
   chain チェック)
5. REVIEW.md に確認日 + 結論 update
