# localsend-bin review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `localsend-bin` PKGBUILD (pkgver=1.17.0, pkgrel=1) を fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/localsend-bin
  - maintainer: NourEddineX
- Upstream: https://github.com/localsend/localsend (= LocalSend organization、Apache-2.0、81k stars)

## 検証結果

- [x] `source_x86_64` URL = `github.com/localsend/localsend/releases/download/v1.17.0/LocalSend-1.17.0-linux-x86-64.deb`
  - upstream 公式 org `localsend/localsend` の release artifact、typosquatting / mirror spoof なし
- [x] `sha256sums_x86_64` が upstream release と一致
  - 実測: `b0244b2c3eacb2a81d61b2662534d6036ab37ace10d6782da36b630c222fa04c`
  - PKGBUILD 値: `b0244b2c3eacb2a81d61b2662534d6036ab37ace10d6782da36b630c222fa04c`
  - 一致
- [x] Tag `v1.17.0` の git commit (`7f21d1f9082a43803e05c37f021912e012145aa5`) は
      **GPG verified** (= Tien Do Nam = 主開発者が署名)、tampering なし
- [x] `prepare()`: `tar -xf data.tar.xz` (= debian package 内 control 構造の
      展開、ar / dpkg-deb 経由ではないが結果同等、安全)
- [x] `build()`: `.desktop` の `Exec` / `Icon` を `localsend_app` → `localsend`
      に sed 置換のみ (= 表示用 rename、外部処理なし)
- [x] `package()`:
  - `/usr/share/applications/localsend.desktop`
  - `/usr/share/icons/hicolor/{128x128,256x256}/apps/localsend.png`
  - `/opt/localsend/` (binary + bundled libapp.so / Flutter runtime)
  - `/usr/bin/localsend` (symlink to /opt/localsend/localsend)
  - すべて install -Dm / cp -a / mv / ln -s 標準コマンド、network fetch / eval なし
- [x] `depends`: `fuse2`, `xdg-user-dirs`, `libayatana-appindicator`
  - Flutter binary が要求する system tray + xdg dir 系、妥当
- [x] license `Apache-2.0` — upstream 一致

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key 483D...` 可。

`.deb` を source にして Flutter SDK build chain を回避する判断は plan
通り。upstream の release CI に乗った prebuilt binary を信頼境界に入れる
形 (= LocalSend org の release pipeline 信頼)。改変 step は cosmetic
rename のみで、実行 binary は upstream そのまま。

## 更新方針

upstream の新 release (v1.18.0 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立に再計算 (= `curl -fsSL <url> | sha256sum` で照合)
4. tag commit が GPG verified か gh api で確認
5. REVIEW.md に確認日 + 結論 update
