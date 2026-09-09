# sunshine-bin review

## 状態

**review 済み、approve** (最新: 2026-09-09 / upstream 2026.906.222525)

AUR の `sunshine-bin` PKGBUILD を fork。改変なし (faithful fork)。

## Source

- AUR: https://aur.archlinux.org/packages/sunshine-bin
  - maintainers: chung <me@chungn.com> / Jay Chu <tothesong@gmail.com>
- Upstream: https://github.com/LizardByte/Sunshine
  - LizardByte organization、GPL-3.0-only、Moonlight 用セルフホスト配信ホスト。
    配信先クライアントは本 repo の `moonlight-qt`。
- Homepage: https://app.lizardbyte.dev

## 検証結果

- [x] `source` URL = `github.com/LizardByte/Sunshine/releases/download/v2026.906.222525/sunshine-2026.906.222525-1-x86_64.pkg.tar.zst`
  - upstream 公式 release CI が生成する **公式 Arch package (`.pkg.tar.zst`)** を
    そのまま再梱包する形。typosquatting なし (LizardByte org の official release path)
- [x] `b2sums` が upstream artifact と一致 (**独立再計算で検証**)
  - 実測: `curl -fsSL <url> | b2sum` =
    `b73ae29bf3e7763c5187b0ebfe2061a87764a79d699df9737ba7170a0a5954e961601afbccb7ccfb0140d561f60c3a7895049ec421c4f9f239267a46111e5b43`
  - PKGBUILD 値: 同上
  - 一致 (BLAKE2b-512、collision 実用上不可能)。SKIP 不使用
- [x] `build()` 無し。`package()` は展開済み upstream pkg から
  `install -Dm755 usr/bin/sunshine` + `cp -r usr/lib usr/share` のみ。
  **build 時のネットワーク取得・curl/wget/exec・shell injection 無し**
- [x] `prepare()` は AUR 側で全行コメントアウト (旧 boost/icu ABI 用 patchelf、
  現 release では不要)。有効な処理無し
- [x] `install=sunshine.install` scriptlet:
  - `do_setcap`: `setcap cap_sys_admin,cap_sys_nice+p /usr/bin/sunshine`
    - `cap_sys_admin` = **KMS/DRM フレームキャプチャに必須** (Wayland/Sway で
      画面全体を掴む正規経路)、`cap_sys_nice` = エンコード thread の realtime 優先度。
      Sunshine 公式ドキュメント記載の権限。過剰付与ではない
  - `do_udev_reload`: `udevadm control --reload-rules` + `udevadm trigger`
    (DEVNAME=/dev/uinput,/dev/uhid,hidraw subsystem,input subsystem) +
    `modprobe uinput uhid` — 入力注入デバイス (仮想 kbd/mouse/gamepad/joypad) を
    有効化。2026-09-09 の libvirtualhid 移行で hidraw/input subsystem trigger が
    追加されたが injection 無し、妥当
- [x] `depends`: avahi(mDNS 探索) / libdrm・libva・vulkan-icd-loader(capture+encode) /
  libevdev・udev(入力) / miniupnpc(UPnP) / opus(音声) / qt6-base・qt6-svg・gtk3
  (2026-09-09 の Qt system tray 移行で新規、`ldd` で `libQt6*` リンク確認済み) /
  gcc-libs(libstdc++/libgcc_s、prebuilt バイナリの runtime dep) /
  hicolor-icon-theme(desktop icon fallback) 等、配信ホストとして妥当。
  全て Arch 公式 repo で解決 (`udev` のみ systemd の virtual provide → .deps.lock は
  `# MISSING` 扱い)
- [x] `optdepends`: `cuda`(NVIDIA NVENC) / `libva-mesa-driver`(AMD) /
  `xorg-server-xvfb`(headless test)。 いずれも任意
- [x] `options=('!strip' '!debug')`: 2026-09-09 に AUR 側で新規追加。本 pkg は
  upstream prebuilt バイナリ (既に strip 済み) の再梱包のため makepkg の
  strip/debug pkg 生成を抑制する目的、妥当
- [x] `conflicts`/`provides` = `sunshine` (source 版と排他、正しい)
- [x] license `GPL-3.0-only` — upstream 一致

## 結論

**approve** — そのまま build host で `makepkg -s --sign` 可。

公式 release CI が生成した Arch package をそのまま b2sums pin で取り込むだけの薄い
-bin package (localsend-bin / voicevox-bin と同型)。source 版 (`sunshine`) は
`sha256sums=SKIP` + build 時に npm / git submodule を取得するため [nekono] 方針
(pre-push gate ③ / supply-chain 方針) に抵触するので **-bin を採用**。

## AUR との意図的 diff

無し (faithful fork)。将来 bump 時に AUR との diff を取りやすくするため .SRCINFO を同梱。

## 依存方針

- `cuda` は optdepends のまま (= 巨大。NVENC は runtime に `nvidia-utils` の
  `libnvidia-encode`/`libcuda` を dlopen する経路で足りることが多い)。CUDA colorspace
  変換が要る場合のみ後入れ。
- `udev` は `# MISSING` (systemd が provide)。dep-version-pr の監視対象外。
- `qt6-base`/`qt6-svg`/`gtk3`/`gcc-libs`/`hicolor-icon-theme` は 2026-09-09
  (v2026.906.222525) の Qt system tray 移行で AUR 側が新規追加、`libnotify` は
  削除。faithful fork のため AUR の depends 変更をそのまま追従する方針。

## 更新方針

upstream の新 release (v2026.5xx.xxxxxx 等) が出たら:
1. AUR で pkgver / b2sums を確認 (nvchecker は tag の `v` prefix を除去して検知)
2. 本 dir の PKGBUILD / .SRCINFO を差し替え
3. b2sums を独立再計算 (`curl -fsSL <url> | b2sum`)
4. install scriptlet (setcap の cap 集合) が変わっていないか確認
5. REVIEW.md 更新履歴に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | findings |
|---|---|---|---|
| 2026-07-10 | 2026.516.143833-1 | (this PR) | 新規追加。faithful fork、b2sums 独立検証一致、approve |
| 2026-07-15 | 2026.516.143833-2 | (this PR) | `pkgrel` +1 (deps changed): libpipewire 1:1.6.7-1 → 1:1.6.8-1 |
| 2026-08-04 | 2026.516.143833-3 | bot PR #477 | `pkgrel` +1 (deps changed): vulkan-icd-loader 1.4.350.1-1 → 1.4.357.0-1 |
| 2026-08-20 | 2026.516.143833-4 | (this PR) | `pkgrel` +1 (deps changed): libevdev 1.13.6-1 → 1.13.7-1 |
| 2026-08-30 | 2026.516.143833-5 | (this PR) | `pkgrel` +1 (deps changed): libayatana-appindicator 0.6.0-1 → 0.6.0-2、openssl 3.6.3-1 → 3.6.4-1 |
| 2026-09-06 | 2026.516.143833-6 | (this PR) | `pkgrel` +1 (deps changed): curl 8.21.0-1 → 8.22.0-1 |
| 2026-09-09 | 2026.906.222525-1 | (this PR) | **needs-attention → 手当済み、優先 publish** (Issue #613)。5件の High/Medium security advisory (GHSA-6w33-pjh7-p77c / GHSA-6jvv-jqr7-m6m3 / GHSA-36ff-frg7-492f / GHSA-26q2-58j6-qmvv / GHSA-c428-87f8-rrv5) を修正するセキュリティリリース。upstream release notes 「The system tray migrated to Qt on every platform. Downstream packages must provide the applicable Qt Widgets and SVG dependencies.」に伴い AUR 側 PKGBUILD の depends が変化: `gcc-libs` `gtk3` `hicolor-icon-theme` `qt6-base` `qt6-svg` を追加、`libnotify` を削除。独立検証として展開済み `usr/bin/sunshine` を `ldd` し、`libQt6Widgets.so.6`/`libQt6Gui.so.6`/`libQt6Core.so.6`/`libQt6DBus.so.6`/`libstdc++.so.6`/`libgcc_s.so.1` が新規リンクされ `libnotify` へのリンクが無いことを確認、AUR の depends 変更は妥当と判断。`options=('!strip' '!debug')` も AUR 側で新規追加 (本 pkg は upstream prebuilt バイナリの再梱包のため既存 strip 済み)。b2sums を独立実測 `b73ae29b…` へ更新 (upstream artifact と一致)。`sunshine.install` は AUR 側で `udevadm trigger --subsystem-match=hidraw` / `--subsystem-match=input` の 2 行追加 (libvirtualhid 移行に伴う joypad udev trigger 拡充、injection 無し) を追従。`package()` の `cp -r usr/lib usr/share` は無改変で新規同梱の systemd user unit / udev rules / modules-load.d / icons を機械的に取り込む。tag `v2026.906.222525` target commit `cb72dffa3233c5815cd5ba88f09f049dd679ba75`、release author LizardByte 公式 org。`.deps.lock` に `gcc-libs`/`gtk3`/`hicolor-icon-theme`/`qt6-base`/`qt6-svg` を追加、`libnotify` を削除 (全て Arch 公式 core/extra で解決確認済み)。pkgrel reset 6→1。Closes #613。 |
