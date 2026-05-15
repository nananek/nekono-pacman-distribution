# qview review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `qview` PKGBUILD (pkgver=7.1, pkgrel=2) を fork。改変なし、そのまま採用。

## Source

- AUR: https://aur.archlinux.org/packages/qview
- Upstream: https://github.com/jurplel/qView (= PKGBUILD maintainer `jurplel` 本人の repo)
- Homepage: https://interversehq.com/qview/

## 検証結果

- [x] `source` URL = `github.com/jurplel/qView/archive/refs/tags/7.1.tar.gz`
  - upstream 公式 repo (jurplel が author、2018 から運用)、typosquatting なし
- [x] `sha256sums` が upstream tarball と一致
  - 実測: `89189b508b60526af09a15bc7b467eecb7f3d074f5dd21d251afe23406b24e8a`
  - PKGBUILD 値: `89189b508b60526af09a15bc7b467eecb7f3d074f5dd21d251afe23406b24e8a`
  - 一致
- [x] Tag `7.1` の git commit (`0ec246b78c310d5c842836a91566db24037c75c2`) は
      **GPG verified** (= author `jurplel` / Benjamin O が署名)、tampering なし
- [x] `makedepends`: `qt6-tools` (qmake6 のため、妥当)
- [x] `depends`: `qt6-base`, `hicolor-icon-theme` (= Qt6 GUI + icon theme、標準)
- [x] `optdepends`: `qt6-imageformats`, `kimageformats`, `qt6-svg` (= 追加 image format support、optional)
- [x] `build()`: `qmake6 PREFIX=/usr && make` — 標準パターン、network fetch / eval / 怪しい構造なし
- [x] `package()`: `make INSTALL_ROOT="$pkgdir/" install` — 標準
- [x] license `GPL3` — upstream GPL-3.0 と一致

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key 483D...` 可。
upstream の release pipeline と PKGBUILD maintainer の identity が一致しており、
supply chain として強い。

## 更新方針

upstream の新 tag が出たら:
1. AUR で pkgver / pkgrel / sha256sums の値を確認
2. このディレクトリの PKGBUILD を差し替え
3. sha256 を独立に再計算 (= AUR の値を盲信せず `curl -fsSL <url> | sha256sum` で照合)
4. tag commit が GPG verified か gh api で確認
5. REVIEW.md に確認日 + 結論 update
