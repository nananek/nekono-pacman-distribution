# localsend-bin review

## 状態

**review 済み、approve** (最新: 2026-11-26 / 1.19.0-1)

AUR の `localsend-bin` PKGBUILD を fork。desktop file の解決だけ本 repo で改変
(v1.19.0 で arch 非対称になったため。詳細は「検証結果」と「更新履歴」)。

## Source

- AUR: https://aur.archlinux.org/packages/localsend-bin
  - maintainer: NourEddineX
- Upstream: https://github.com/localsend/localsend (= LocalSend organization、Apache-2.0、81k stars)

## 検証結果

- [x] `source_x86_64` URL = `github.com/localsend/localsend/releases/download/v1.19.0/LocalSend-1.19.0-linux-x86-64.deb`
  - upstream 公式 org `localsend/localsend` の release artifact、typosquatting / mirror spoof なし
- [x] `sha256sums` が upstream release と一致 (release asset digest と独立実測の両方)
  - x86_64 実測: `f2cc2d1f8a2542799576e4c1cc61ed1f35c6a7dba2f375870e8500eab6fbfe66`
  - aarch64 実測: `8a54f1dc0f70ef4d9fa3037a29e8713d5580c461675fae9a5ba59b39010513c3`
  - どちらも release API の `sha256:` digest と一致
- [x] Tag `v1.19.0` の git commit (`44f48f18a5503c43873570ff489d27ff1983e198`) は
      **GPG verified** (author `Tienisto` = Tien Do Nam = 主開発者、release は
      `github-actions[bot]` 発行)、tampering なし
- [x] `.deb` の payload を実物で確認 (ar で `data.tar.zst` を取り出して列挙)
  - `opt/localsend_app/`、`usr/share/icons/hicolor/{128x128,256x256}/apps/localsend_app.png`
    は 1.18.2 と同一
  - **desktop file 名だけ arch 非対称**: x86_64 は
    `org.localsend.localsend_app.desktop`、aarch64 は `localsend_app.desktop`
    (upstream の x86_64 deb workflow にだけ flutter_app_packager を rename する
    sed step が入ったため)。`build()` / `package()` は glob で実ファイルを 1 個解決し、
    0 個 / 2 個以上なら build を失敗させる形にした
  - 1.19.0 から desktop file に `StartupWMClass=org.localsend.localsend_app` が
    upstream 側で入った (両 arch とも)。本 repo の rename 後もこの行は維持され、
    x86_64 の KWin は app id 一致、aarch64 も WMClass で解決できる
- [x] `prepare()`: `tar -xf data.tar.zst` (1.18.0 以降の deb と同じ zst 構成)
- [x] `build()`: `.desktop` の `Exec` / `Icon` を `localsend_app` → `localsend`
      に sed 置換のみ (= 表示用 rename、外部処理なし)
- [x] `package()`:
  - `/usr/share/applications/localsend.desktop`
  - `/usr/share/icons/hicolor/{128x128,256x256}/apps/localsend.png`
  - `/opt/localsend/` (binary + bundled libapp.so / Flutter runtime)
  - `/usr/bin/localsend` (symlink to /opt/localsend/localsend)
  - すべて install -Dm / cp -a / mv / ln -s 標準コマンド、network fetch / eval なし
- [x] `depends`: `fuse2`, `xdg-user-dirs`, `libayatana-appindicator`,
      `libayatana-indicator`
  - 1.19.0 の `libtray_manager_plugin.so` の `DT_NEEDED` を readelf で確認:
    `libayatana-appindicator3.so.1` / `libayatana-indicator3.so.7` は 1.18.2 と
    同じで、依存の追加・削除なし
  - `localsend_app` 本体の要求 GLIBC 最大は `2.34` (Arch で問題なし)
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
| 2026-09-04 | 1.18.2-2 | (this PR) | — (pkgrel bump のみ) | runtime dependency fix。実行時エラー `libayatana-indicator3.so.7: cannot open shared object file` を確認。ELFの`DT_NEEDED`と`pacman -Qo`で、同ライブラリの提供元が既存依存`libayatana-appindicator`ではなく別packageの`libayatana-indicator`であることを確認し、直接依存へ追加。併せて`.deps.lock`を現在のArch repo versionへ更新。 |
| 2026-10-13 | 1.18.2-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libayatana-appindicator 0.6.0-2 → 0.6.1-1 |
| 2026-11-26 | 1.19.0 | (this PR) | `44f48f18a5503c43873570ff489d27ff1983e198` | safe-to-bump (Issue #843、保留 Issue #847 の判断もここで確定)。release は `github-actions[bot]` 発行、tag commit は `Tienisto` 署名で GPG verified、`.deb` 2 つの digest は release API と独立実測が一致。payload を実物で列挙し、`opt/localsend_app/` とアイコンは不変、**desktop file 名だけ x86_64 = `org.localsend.localsend_app.desktop` / aarch64 = `localsend_app.desktop`** と arch 非対称であることを確認。単一 `_pkgdesktop` では x86_64 で build が壊れるため、`build()` / `package()` が `usr/share/applications/*.desktop` を 1 個だけ解決する `_desktop_file()` を使う形に変更 (0 個 / 2 個以上は build 失敗)。AUR 1.19.0-2 は x86_64 名を固定していて aarch64 が壊れるので、この glob は意図的 diff。upstream が 1.19.0 で `StartupWMClass=org.localsend.localsend_app` を両 arch の desktop file に追加済み。`libtray_manager_plugin.so` の `DT_NEEDED` は 1.18.2 と同じで depends 変更なし。x86_64 は `makepkg -f --nodeps` でビルド成功、aarch64 payload でも glob 解決を確認。Closes #843, closes #847。 |

## 更新方針

upstream の新 release (v1.18.0 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立に再計算 (= `curl -fsSL <url> | sha256sum` で照合)
4. tag commit が GPG verified か gh api で確認
5. REVIEW.md に確認日 + 結論 update
