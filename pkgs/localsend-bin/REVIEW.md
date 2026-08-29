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

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-06-21 | 1.17.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libayatana-appindicator 0.5.94-1 → 0.6.0-1 |
| 2026-08-11 | 1.18.0 (build fix) | — (upstream bump) | `82471a523411906d52f410e71503093c563ba44a` | pkgver bump、depends 変更なし。v1.18.0 の `.deb` が `data.tar.zst` と `opt/localsend_app` を使用するため `prepare()` / `package()` の対象を更新。実行ファイルの配置方針は不変、checksum 検証機能追加と Rust コアリライトは upstream CI 内で完結 |
| 2026-08-30 | 1.18.2 | (this PR) | `v1.18.2` (release author `github-actions[bot]`、CI 発行) | safe-to-bump (Issue #562)。v1.18.0→1.18.2 累積 diff (42 commits/110 files) は主に 1.17.0 以前との後方互換性回復、drag&drop 改善、web share ページのカスタマイズ機能追加。**security fix 含む**: peer が送る HTTP redirect を追従しないよう変更 (SSRF/redirect 悪用対策)。`.deb` の内部構造 (`data.tar.zst`、`opt/localsend_app/localsend_app`、`usr/share/applications/localsend_app.desktop`、`usr/share/icons/hicolor/{128x128,256x256}/apps/localsend_app.png`) は 1.18.0 と完全一致、`prepare()`/`build()`/`package()` 変更不要。depends 変更なし。sha256sums_x86_64/aarch64 を独立実測で更新。Closes #562。 |

## 更新方針

upstream の新 release (v1.18.0 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立に再計算 (= `curl -fsSL <url> | sha256sum` で照合)
4. tag commit が GPG verified か gh api で確認
5. REVIEW.md に確認日 + 結論 update
