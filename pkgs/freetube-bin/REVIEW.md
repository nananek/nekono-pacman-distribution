# freetube-bin review

## 状態

**review 済み、approve** (2026-08-30)

AUR の `freetube-bin` PKGBUILD (pkgver=0.25.3, pkgrel=1) を fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/freetube-bin
  - maintainer: plague-doctor / co-maintainer: bacteriostat
- Upstream: https://github.com/FreeTubeApp/FreeTube (= FreeTubeApp organization、AGPL-3.0、YouTube privacy client)

## 依存方針

AUR には source build 版 `freetube` (electron42 + `pnpm install` で build 時に
npm registry から動的取得が発生) も存在するが、CLAUDE.md の `pip install`
禁止と同じ理由で動的取得ビルドは避けたいため、prebuilt `.deb` 版の
`freetube-bin` のみを採用する (= 「基本は片方だけ採用」の判断)。

## 検証結果

- [x] `source_x86_64`/`source_aarch64` URL = `github.com/FreeTubeApp/FreeTube/releases/download/v0.25.3-beta/freetube_0.25.3_beta_{amd64,arm64}.deb`
  - upstream 公式 org `FreeTubeApp/FreeTube` の release artifact、typosquatting / mirror spoof なし
- [x] `sha256sums_x86_64` / `sha256sums_aarch64` が upstream release と一致
  - 実測 (amd64): `3a75dc4ef65edf184222c3389855be4699d04cce321e9695bb2989a50db2ba5d`
  - 実測 (aarch64): `59a39a54055f414c2ea9666c373c3a4970266d8d366d52a839278d20f5ee8bc8`
  - PKGBUILD 値と一致
- [x] Tag `v0.25.3-beta` の commit (`ff510f8f6c4a0da2602f90336c555f75c75bfa47`) は
      GitHub API 上 `verification.verified = true` (reason: valid)、tampering なし
- [x] `prepare()`: `bsdtar -x -f data.tar.xz -C data` (= debian package の
      data.tar.xz 展開のみ、外部 fetch / eval 無し)
- [x] `package()`: `cp -a` で `/opt/FreeTube/` 一式 (Electron bundle) と
      `/usr/share/applications/freetube.desktop` を配置、`ln -s` で
      `/usr/bin/freetube` symlink。標準コマンドのみ、network fetch 無し
  - `.deb` 中身を展開して直接確認 (`ar x` → `data.tar.xz` 一覧):
    `opt/FreeTube/freetube` (実行 binary)、bundled Chromium 系 `.so`
    (libEGL/libGLESv2/libvulkan/libffmpeg 等、Electron 標準同梱)、
    `opt/FreeTube/resources/app.asar`、`usr/share/applications/freetube.desktop`
    のみで、不審なファイルは無い
- [x] `depends`: `ttf-liberation` (Arch 公式 extra repo, 2.1.5-2) のみ
- [x] `options=('!strip' 'staticlibs')`: bundled prebuilt binary/so 群を
      strip しない指定、Electron bin として妥当
- [x] license `AGPL3` — upstream 一致 (repo は `GPL-3.0` ではなく AGPL-3.0-only 系)

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key <Nekono key>` 可。

`.deb` を source にして Electron build chain (electron42 + pnpm) を回避する
判断。upstream の GitHub Actions release pipeline (verified commit) 由来の
prebuilt binary を信頼境界に入れる形。改変 step なし、実行 binary は
upstream そのまま。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-08-30 | 0.25.3-1 | (this PR) | `ff510f8f6c4a0da2602f90336c555f75c75bfa47` | 新規導入、AUR PKGBUILD 無改変で採用 |

## 更新方針

upstream の新 release (v0.25.4-beta 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立に再計算 (= `curl -fsSL <url> | sha256sum` で照合)
4. tag commit の `verification.verified` を `gh api` で確認
5. REVIEW.md に確認日 + 結論 update
