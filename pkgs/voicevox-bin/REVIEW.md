# voicevox-bin review

## 状態

**review 済み、 approve** (2026-05-19)

AUR の `voicevox-bin` PKGBUILD (pkgver=0.25.2, pkgrel=1) を fork。 意図的
改変 2 件 (split package 解消 + GPU 動的切替の削除)。

## Source

- AUR: https://aur.archlinux.org/packages/voicevox-bin
  - Maintainer: t1ckbase
- Upstream:
  - https://github.com/VOICEVOX/voicevox
  - VOICEVOX Project、 LGPL-3.0-only
  - 上流 release tag `0.25.2` の commit SHA: `823a760f632d63d23522eb2cee6992b9b42119a5`
- 上流公式の prebuilt AppImage: GitHub Releases から download
  (= `VOICEVOX-CPU-X64.AppImage.7z.00{1,2}`)

## 検証結果

- [x] source URL = `github.com/VOICEVOX/voicevox/releases/download/0.25.2/...`
  - upstream 公式 GitHub Release、 typosquatting なし
- [x] b2sums が upstream tarball と一致 (= AUR と同値、 makepkg --verifysource 通過)
  - part 001: `e36263391b8af03da7e4e44995f5c65961fbbef7905d4cb107cb8b3fedf7bb0fbc381da261464f587126a123004b28530116170c77c2e578da182ad618ffcdc9`
  - part 002: `b0079c64ef268c207fcf870771e2436d440ad6a8465c6ae0a65d05343ae6e0e1500bd7c372e0aeef5229ab80b861776b07432ae79d8a5006824e496d28ffc41d`
  - 参考 sha256 (cross-record):
    - part 001: `32704fc4731a97f285511512cbd0034c1eb39fc21729708cbb2f073f063b9a77`
    - part 002: `ae48664607f1d87959cfda542f724a2a05b63e48eb66166c226eebf67f6b232c`
- [x] 上流公開 manifest `linux-cpu-x64-appimage.7z.txt` は part サイズ + MD5
  (= AD2BE311C55AC08E59A97557D107DAFA / A2C861AF488DB6E97BDBDA4B304C68BA) のみ
  公開。 sha 系は未公開。 hash の信頼の起点は AUR メンテナの b2sums 計算 + 当
  PR で build host (nekono-pacman0) でも再計算して一致確認した値。
- [x] `prepare()`:
  - 7z で multi-part AppImage を結合解凍
  - AppImage 自体は self-extracting ELF。 そこから `voicevox.desktop` /
    `voicevox.png` / `resources/app.asar` のみを 7z で再抽出。 AppImage 全体は
    package に含めない (= 軽量化、 system electron37 経由で起動するため)
  - `.desktop` の Exec を `voicevox` に固定 (= /usr/bin/voicevox 経由)
  - シェル injection・eval・curl・wget なし
- [x] `package()`: `install` のみ、 ネットワーク操作なし
- [x] depends: 7zip / bash / electron37 — 全て妥当
- [x] optdepends: voicevox-engine — voicevox-engine-cuda (`provides=voicevox-engine`)
  が満たす。 ayaka 単独 install 時は network 経由 (Tailscale 100.117.10.53:50021)
  で remote engine を指定する運用

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| split package (`pkgname=(voicevox-bin voicevox-appimage)`) を解消、 `voicevox-bin` のみに | [nekono] では editor 1 variant (= system electron37 経由) に絞る。 AppImage 全体を抱える voicevox-appimage は配布対象外 (ayaka では electron37 を共通で持っている方が他 Electron アプリと共有できて軽い) |
| AUR の `if [ -c /dev/nvidia0 ]; then source=(...NVIDIA AppImage...)` を完全削除 | **重大な supply-chain 欠陥**。 build host の GPU 有無で source を動的切替する非決定 PKGBUILD で、 NVIDIA 版 AppImage 分の sha256 は pin されていなかった (= GPU 環境で build すると無検証 binary が package に混入)。 ayaka は editor だけ要り engine は remote (= nekono-pacman0 上の voicevox-engine-cuda) なので、 editor の GPU 対応は不要 → CPU AppImage 固定で十分 |
| `optdepends` の wording を「local もしくは remote、 EngineSettings で URL 指定」に改善 | engine が必ずしも同 host にある前提を消し、 remote engine 運用 (= nekono-pacman0 → ayaka) を documentation 化 |

## 結論

**approve** — build host (nekono-pacman0) で `makepkg -s --sign --key 483D...` 可。
ayaka 側にも [nekono] 経由で `pacman -S voicevox-bin` で入れて editor を起動、
初回起動時に EngineSettings の API URL を `http://100.117.10.53:50021`
(nekono-pacman0 の Tailscale IP) に向ける。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.25.2 (build fix) | (本 PR の commit SHA を merge 時に追記) | — | prepare() を 2-step `7z x` に分割。 7zip 26.01 が split merge 後に nested .7z まで自動で降りない挙動差異への対応。 build host (nekono-pacman0) の実 build で発覚 |
| 2026-05-19 | 0.25.2 | (本 PR の commit SHA を merge 時に追記) | `823a760f632d63d23522eb2cee6992b9b42119a5` | 初回 add、 split package 解消 + GPU 動的切替の supply-chain 欠陥を除去 |
| 2026-06-06 | 0.25.2-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): bash 5.3.9-1 → 5.3.12-1 |
