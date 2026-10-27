# voicevox-bin review

## 状態

**review 済み、 approve** (最新: 2026-10-27 / 0.26.1、初回: 2026-05-19)

AUR の `voicevox-bin` PKGBUILD (pkgver=0.25.2, pkgrel=1) を fork。 意図的
改変 2 件 (split package 解消 + GPU 動的切替の削除)。

## Source

- AUR: https://aur.archlinux.org/packages/voicevox-bin
  - Maintainer: t1ckbase
- Upstream:
  - https://github.com/VOICEVOX/voicevox
  - VOICEVOX Project、 LGPL-3.0-only
  - 上流 release tag `0.26.1` の commit SHA: `65ba65b43c30f1763f72ff8fe237b991ff73bb1d`
- 上流公式の prebuilt AppImage: GitHub Releases から download
  (= `VOICEVOX-CPU-X64.AppImage.7z.00{1,2}`。 0.26.0 で 7z の分割サイズが
  1.9 GB 程度に上がったため、 0.26.1 の part は 1,992,294,400 + 62,607,471 byte)

## 検証結果

- [x] source URL = `github.com/VOICEVOX/voicevox/releases/download/0.26.1/...`
  - upstream 公式 GitHub Release、 typosquatting なし (release author は
    `github-actions[bot]` = upstream の CI からの publish で、 tag commit は
    `65ba65b4...`、 0.25.2 時と同じ経路)
- [x] b2sums が実測値と一致 (= `makepkg --verifysource` 通過)
  - part 001: `275680104b2ee50d73e95128b2036edc77de9dfec93bcfafc1dfed9e37e9a1f9734ce1b04dee5bc763c1e62bf95e5ff60b2a1167c74343bb5344620b3de79640`
  - part 002: `c1fe30aa99b8102b4db5feccfc0030709c239904b96d729c1f28236f2d3df666661e195eb5b4f85840f74213ddc90061b716580be7e3609e4f4144ba7537fb1a`
  - sha256 (cross-record):
    - part 001: `73e31a124778b62f1cf26a9bccf6bc2b193fadfc5b349aac63d577ea8958b4f5`
    - part 002: `365fb0f48c196fe4891afdd8493c85447b845a8090e8afbeff9a5a7ae8d5a3bc`
- [x] upstream 公開 manifest `linux-cpu-x64-appimage.7z.txt` と sha256 で一致
  - **0.26.1 から upstream はこの manifest で MD5 ではなく sha256 を公開する
    ようになった** (0.25.2 時点は part サイズ + MD5 のみ)。 上記 cross-record の
    sha256 2 値は manifest の記載と完全一致し、 同 asset の GitHub digest
    (`sha256:558666b9...`) とも一致。 つまり 0.25.2 時のような「AUR メンテナの
    手計算 b2sum を信頼の起点にする」必要はなくなり、 今回の b2sums は
    build host での実測 + upstream 公開 sha256 / GitHub asset digest との照合で
    裏付けられている
  - AUR は 0.25.2 止まり (`5508c98`) で 0.26.1 未追随のため AUR 値との
    cross-check は不可 (sha の裏付けは上記 upstream 公開値で足りる)
- [x] `prepare()`:
  - 7z で multi-part AppImage を結合解凍 (0.26.1 の展開後 AppImage 名は
    `VOICEVOX-CPU.0.26.1-x64.AppImage`。 0.25.2 の `VOICEVOX-CPU-X64.AppImage`
    から命名が変わったため、 glob `VOICEVOX*.AppImage` で拾う既存実装がそのまま効く)
  - AppImage 自体は self-extracting ELF。 そこから `voicevox.desktop` /
    `voicevox.png` / `resources/app.asar` のみを 7z で再抽出。 AppImage 全体は
    package に含めない (= 軽量化、 system electron42 経由で起動するため)。
    3 file とも 0.26.1 でも同 path に存在することを `7z l` で確認
  - `.desktop` の Exec を `voicevox` に固定 (= /usr/bin/voicevox 経由)。
    0.26.1 の `.desktop` も `Exec=AppRun %U` で sed 対象行は不変
  - シェル injection・eval・curl・wget なし
- [x] `package()`: `install` のみ、 ネットワーク操作なし
- [x] `resources/app.asar` の中身を確認: 81 file (`dist/` 配下の build 済み JS /
  CSS / 画像と `package.json` のみ)。 **native module (`.node`) は 0 個**、
  install / postinstall script も無し。 内部 `package.json` は
  `voicevox-cpu` 0.26.1、 `scripts` は空、 依存は全て build 時に bundle 済み
- [x] depends: `bash` / `electron42` (+ makedepends `7zip`) — 妥当。
  **0.26.0 で upstream の electron が 37.7.1 → 42.1.0、 0.26.1 で 42.11.18 に
  上がった** (emSign 証明書問題への対応)。 配布 AppImage 自身も
  `Chrome/148.0.7778.280` (= Electron 42 系) を積んでいることを `strings` で確認。
  Arch 公式 extra の `electron42` は pin と完全一致する 42.11.18-1 で、
  `/usr/bin/electron42` を持つことを package を落として `tar tf` で確認した。
  旧 `electron37` (Electron 37 = Node 22) のままでは upstream の
  `engines: node >=24.11.1` と噛み合わないため electron42 に切り替えた。
  これに伴い [nekono] の `electron37-bin` は依存元が無くなる (= 撤廃は
  user 承認事項なので本 bump では触らない)
- [x] optdepends: voicevox-engine — voicevox-engine-cuda (`provides=voicevox-engine`)
  が満たす。 ayaka 単独 install 時は network 経由 (Tailscale 100.117.10.53:50021)
  で remote engine を指定する運用。 0.26.0 で「エンジンをエディタから分離して
  アップデートできる」機能が入ったが、 本 pkg は app.asar のみを install し
  engine を同梱しない方針は不変 (runtime の engine 取得はユーザ操作)

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| split package (`pkgname=(voicevox-bin voicevox-appimage)`) を解消、 `voicevox-bin` のみに | [nekono] では editor 1 variant (= system electron42 経由) に絞る。 AppImage 全体を抱える voicevox-appimage は配布対象外 (ayaka では electron42 を共通で持っている方が他 Electron アプリと共有できて軽い) |
| AUR の `if [ -c /dev/nvidia0 ]; then source=(...NVIDIA AppImage...)` を完全削除 | **重大な supply-chain 欠陥**。 build host の GPU 有無で source を動的切替する非決定 PKGBUILD で、 NVIDIA 版 AppImage 分の sha256 は pin されていなかった (= GPU 環境で build すると無検証 binary が package に混入)。 ayaka は editor だけ要り engine は remote (= nekono-pacman0 上の voicevox-engine-cuda) なので、 editor の GPU 対応は不要 → CPU AppImage 固定で十分 |
| `optdepends` の wording を「local もしくは remote、 EngineSettings で URL 指定」に改善 | engine が必ずしも同 host にある前提を消し、 remote engine 運用 (= nekono-pacman0 → ayaka) を documentation 化 |
| system electron を `electron37` から `electron42` へ変更 (0.26.1 で追加) | upstream が 0.26.0 で electron を 42.1.0 に上げ、 0.26.1 で 42.11.18 pin。 配布 AppImage 自身も Electron 42 系を積む。 Arch 公式 extra に pin 一致の `electron42` 42.11.18-1 があるため [nekono] 自家 build は不要。 AUR は 0.25.2 止まりのため AUR 側にはこの変更はまだ現れていない |

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可。
ayaka 側にも [nekono] 経由で `pacman -S voicevox-bin` で入れて editor を起動、
初回起動時に EngineSettings の API URL を `http://100.117.10.53:50021`
(nekono-pacman0 の Tailscale IP) に向ける。 0.26.1 からは runtime 依存が
`electron42` (Arch 公式 extra) になる。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.25.2 (build fix) | (本 PR の commit SHA を merge 時に追記) | — | prepare() を 2-step `7z x` に分割。 7zip 26.01 が split merge 後に nested .7z まで自動で降りない挙動差異への対応。 build host (nekono-pacman0) の実 build で発覚 |
| 2026-07-14 | 0.25.2 (dep fix, PKGBUILD 無改変) | (this commit) | — | Arch 公式 extra が electron37 を rolling で落としており build 不能になっていた (nekono-pacman0 退役 + build host migration で発覚)。 upstream (VOICEVOX package.json) は electron 37.7.1 に pin 済みで新 major への追従不可のため、[nekono] 自家 `electron37-bin` (electron 公式 GitHub Release prebuilt、provides=electron37) を新規 add して解決。voicevox-bin 自体の PKGBUILD/sha256/pkgver は無変更 |
| 2026-05-19 | 0.25.2 | (本 PR の commit SHA を merge 時に追記) | `823a760f632d63d23522eb2cee6992b9b42119a5` | 初回 add、 split package 解消 + GPU 動的切替の supply-chain 欠陥を除去 |
| 2026-06-06 | 0.25.2-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): bash 5.3.9-1 → 5.3.12-1 |
| 2026-06-12 | 0.25.2-3 | bot PR #202 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): bash 5.3.12-1 → 5.3.15-1 |
| 2026-06-27 | 0.25.2-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): 7zip 26.01-1 → 26.02-1 |
| 2026-09-06 | 0.25.2-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): 7zip 26.02-1 → 26.03-1 |
| 2026-09-22 | 0.25.2-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): bash 5.3.15-1 → 5.3.20-1 |
| 2026-10-07 | 0.25.2-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): 7zip 26.03-1 → 26.04-1 |
| 2026-10-22 | 0.25.2-8 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): 7zip 26.04-1 → 26.05-1 |
| 2026-10-27 | 0.26.1 | (this PR) | `65ba65b43c30f1763f72ff8fe237b991ff73bb1d` | Issue #756 (`Closes #756`)。 0.25.2 → 0.26.1。 b2sum を実測し、 **0.26.1 から upstream が `linux-cpu-x64-appimage.7z.txt` で公開するようになった sha256** (part 001 `73e31a12...` / part 002 `365fb0f4...`) と GitHub Release asset digest の両方に一致 (AUR は 0.25.2 止まりで cross-check 不可)。 part 構成が 1 GB × 2 → 1.9 GB + 62 MB に変化 (0.26.0 の「7z 分割サイズを 1.9GB 程度に上げる」変更)。 展開後 AppImage 名は `VOICEVOX-CPU.0.26.1-x64.AppImage` に変わったが prepare() の glob で吸収、 抽出する 3 file (`voicevox.desktop` / 256x256 icon / `resources/app.asar`) の path は不変。 app.asar は 81 file で native module 0 / install script なし。 **depends を electron37 → electron42 に変更** (upstream が 0.26.0 で electron 42.1.0、 0.26.1 で 42.11.18 pin。 AppImage 自身も Chromium 148 = Electron 42 を積む。 Arch 公式 extra の electron42 42.11.18-1 が pin と一致し、 `/usr/bin/electron42` を持つことを package を落として確認)。 これで [nekono] `electron37-bin` は被依存が無くなる (撤廃は user 承認事項のため未処理)。 `.deps.lock` は `electron37=37.5.1-1` → `electron42=42.11.18-1`。 pkgrel は 1 にリセット、 pkgdesc も `system electron42` に更新 |
