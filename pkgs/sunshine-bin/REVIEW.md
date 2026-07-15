# sunshine-bin review

## 状態

**review 済み、approve** (2026-07-10)

AUR の `sunshine-bin` PKGBUILD (pkgver=2026.516.143833, pkgrel=1) を fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/sunshine-bin
  - maintainers: chung <me@chungn.com> / Jay Chu <tothesong@gmail.com>
- Upstream: https://github.com/LizardByte/Sunshine
  - LizardByte organization、GPL-3.0-only、Moonlight 用セルフホスト配信ホスト。
    配信先クライアントは本 repo の `moonlight-qt`。
- Homepage: https://app.lizardbyte.dev

## 検証結果

- [x] `source` URL = `github.com/LizardByte/Sunshine/releases/download/v2026.516.143833/sunshine-2026.516.143833-1-x86_64.pkg.tar.zst`
  - upstream 公式 release CI が生成する **公式 Arch package (`.pkg.tar.zst`)** を
    そのまま再梱包する形。typosquatting なし (LizardByte org の official release path)
- [x] `b2sums` が upstream artifact と一致 (**独立再計算で検証**)
  - 実測: `curl -fsSL <url> | b2sum` =
    `86bc061bfdab533987a1d32faa02ca275364915cc7eebb54e0ab4ddee19c1b9de980e5c7b4a0bfbd0ef5b0a6bf509f25f9525bf214881ce4dd45da6ec6c0f778`
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
    (DEVNAME=/dev/uinput,/dev/uhid) + `modprobe uinput uhid` — 入力注入デバイス
    (仮想 kbd/mouse/gamepad) を有効化。妥当
- [x] `depends`: avahi(mDNS 探索) / libdrm・libva・vulkan-icd-loader(capture+encode) /
  libevdev・udev(入力) / miniupnpc(UPnP) / opus(音声) 等、配信ホストとして妥当。
  全て Arch 公式 repo で解決 (`udev` のみ systemd の virtual provide → .deps.lock は
  `# MISSING` 扱い)
- [x] `optdepends`: `cuda`(NVIDIA NVENC) / `libva-mesa-driver`(AMD) /
  `xorg-server-xvfb`(headless test)。 いずれも任意
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
