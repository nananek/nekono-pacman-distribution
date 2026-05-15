# moonlight-qt review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `moonlight-qt` PKGBUILD (pkgver=6.1.0, pkgrel=4) を fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/moonlight-qt
  - maintainers: Konstantin Liberty / Cedric Girard / Michael Herzberg
- Upstream: https://github.com/moonlight-stream/moonlight-qt
  - moonlight-stream organization、GPL-3.0、17k stars、開発者 Cameron Gutman (cgutman)
- Homepage: https://moonlight-stream.org

## 検証結果

- [x] `source` URL = `github.com/moonlight-stream/moonlight-qt/releases/download/v6.1.0/MoonlightSrc-6.1.0.tar.gz`
  - upstream 公式 release artifact (= 単純な tag archive ではなく
    submodules を resolve した source tarball を release CI が生成)、
    typosquatting なし
- [x] `sha512sums` が upstream tarball と一致
  - 実測: `390fe3f686c86a52dd0ff4b67e8e8beb6edcb175ddf92bc5de11d92ffdaf0b6a8d76be781c483b685626c705e63f07e156506112923c848a4a798ba703254829`
  - PKGBUILD 値: 同上
  - 一致 (sha512 強度、bit collision 実用上不可能)
- [⚠] Tag `v6.1.0` の git commit (`f786e94c7b2f943e24e65d7d74deb539b827fc84`) は
      **GPG verified: false** — Moonlight project は commit signing を運用
      していない。release artifact の sha512 pin で supply chain は確保
      (= GitHub release CI が cgutman 名義の release から自動生成、TLS で守られる)
- [x] `prepare()` の sed patch:
  - `app/masterhook_internal.c` の 14 行目に `#include <string.h>` を挿入
  - `app/streaming/streamutils.cpp` の 2 行目に `#include <cstring>` を挿入
  - 共に **header include 1 行追加** のみ、code injection ではない。
    GCC 14+ strict mode (= `-Wimplicit-function-declaration` 等で `strlen`
    の暗黙宣言が error になる) への AUR maintainer local fix。
- [x] `build()`: `qmake6 PREFIX=... moonlight-qt.pro && make release` 標準パターン
- [x] `package()`: `make install` (= qmake で PREFIX 設定済み、install rules
      は upstream の .pro 定義通り)
- [x] `depends`: `qt6-base`, `qt6-declarative`, `qt6-svg`, `ffmpeg`, `sdl2_ttf`,
      `sdl2-compat` — GameStream client の Qt6 GUI + H.264/H.265 デコード
      + 入力デバイス系、妥当
- [x] `optdepends`: `libva-intel-driver`, `intel-media-driver` (HW accel、Intel GPU 用)
- [x] license `GPL-3.0-or-later` — upstream 一致

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key 483D...` 可。

submodule 込みの release tarball を upstream の GitHub release CI が生成
してるので、source tree の整合性は upstream 側で保証される (= AUR 側で
git submodule update する必要なし)。

## 更新方針

upstream の新 release (v6.1.1 等) が出たら:
1. AUR で pkgver / pkgrel / sha512sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha512 を独立再計算 (= `curl -fsSL <url> | sha512sum`)
4. prepare() の sed patch が新 source tree でも当たるか確認 (=
   GCC 警告対応のため、新 upstream version で fix 済みなら sed 不要)
5. REVIEW.md に確認日 + 結論 update
