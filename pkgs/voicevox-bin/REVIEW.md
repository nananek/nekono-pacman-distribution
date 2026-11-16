# voicevox-bin review

## 状態

**review 済み、 approve** (最新: 2026-11-16 / 0.26.3、初回: 2026-05-19)

AUR の `voicevox-bin` PKGBUILD (pkgver=0.25.2, pkgrel=1) を fork。 意図的
改変 2 件 (split package 解消 + GPU 動的切替の削除)。

## Source

- AUR: https://aur.archlinux.org/packages/voicevox-bin
  - Maintainer: t1ckbase
- Upstream:
  - https://github.com/VOICEVOX/voicevox
  - VOICEVOX Project、 LGPL-3.0-only
  - 上流 release tag `0.26.3` の commit SHA: `ffa4802e49df7d62f930c66373c442ef0deeda4d`
  (0.26.2 は `b7ad1169ab4a0a7a574312b2de63eb8829c00b1d`、 0.26.1 は
  `65ba65b43c30f1763f72ff8fe237b991ff73bb1d`)
  - 上流公式の prebuilt AppImage: GitHub Releases から download
  (= `VOICEVOX-CPU-X64.AppImage.7z.00{1,2}`。 0.26.3 の part は
  1,992,294,400 + 123,856,375 byte。 0.26.2 の 1,992,294,400 + 123,413,662 byte
  から part 002 が +442,713 byte)

## 検証結果

- [x] source URL = `github.com/VOICEVOX/voicevox/releases/download/0.26.3/...`
  - upstream 公式 GitHub Release、 typosquatting なし (release author は
    `github-actions[bot]` = upstream の CI からの publish で、 tag commit は
    `ffa4802e...`、 0.26.2 時点と同じ経路。 release は draft / prerelease では
    なく、 `releases/latest` も 0.26.3 を返す)
- [x] sha256sums が実測値と一致 (= `makepkg --verifysource` 通過、 rc=0)
  - part 001: `0e6fa43b69dff3972b536c6bb6b857fb680a24c80b91c861955349afa996569d`
  - part 002: `8b6d9698ff17289617f9214af0abb8e356cacec9e9cf9fd48005501ae93903dc`
  - 上記 2 値は GitHub Releases API が asset ごとに返す digest
    (`sha256:0e6fa43b...` / `sha256:8b6d9698...`) と完全一致した。 **fresh
    download した 0.26.3 の bytes に対する照合 (= verifysource) も通過**して
    いるため、 「API の値だけを写した」 状態ではない
- [!] **b2sums → sha256sums に切り替えた (本 bump のみの事情)**
  - 0.26.2 までは b2sums を実測して pin していたが、 **この run の harness では
    b2sum を実測できなかった**。 理由は 2 つ:
    (a) shell が無く `b2sum` を実行できない、
    (b) `fetch` tool も GitHub release asset を取れない
    (= `release-assets.githubusercontent.com` への cross-host redirect が
    allow-list 外)。 よって upstream 公開 manifest
    `linux-cpu-x64-appimage.7z.txt` も本文を読めていない
  - 代替として **GitHub API の asset digest (sha256) を pin** し、
    `makepkg --verifysource` の実 download 照合で裏付けた。 sha256sums は
    CLAUDE.md の review 観点 (「`sha256sums` (or `sha512sums`) で tarball が
    pin されているか」) に沿う形式
  - b2sum を実測できる host で次回 bump する場合は b2sums に戻してよい
    (= 戻す際は値の再計算が必須)
- [x] source の download 名を version 付き (`::` 形式) に変更
  - upstream の asset 名は version を含まないため、 makepkg が前回 bump の
    cache を「同じ file」 と見做して再利用し、 新旧取り違えが起きる。
    **実際に本 bump の検証で 0.26.2 の cache を掴んで checksum mismatch に
    なった** (= そのまま sum を書き換えていたら「古い bytes を新しい version の
    正」 として pin する事故になっていた)。 `voicevox-0.26.3-...` と version を
    名前に含めて構造的に潰した
- [ ] **AppImage 内部の listing は再取得できていない** (本 run の制約)
  - `7z` が shell allow 外で、 AppImage を展開して `voicevox.desktop` /
    `voicevox.png` / `resources/app.asar` の path を再確認できていない。
    代わりに upstream の `0.26.2...0.26.3` 差分 (3 commit / 3 file:
    `.github/workflows/build.yml` / `public/qAndA.md` /
    `public/updateInfos.json`) が **`build/` 配下の electron-builder config と
    `.desktop` template を一切触っていない** ことを確認し、 抽出 path が不変で
    あることを裏付けた。 実 path の最終確認は Phase 3 の `bin/build-all`
    (= prepare() が fail すれば即検出) で行う
- [x] upstream 差分 (0.26.2 → 0.26.3、 3 commit / 3 file)
  - `.github/workflows/build.yml`: `VOICEVOX_ENGINE_VERSION` 0.26.1 → 0.26.2、
    macOS ARM64 runner を `macos-14` → `macos-15` (= Linux/x64 の build 条件と
    AppImage 名には影響なし)
  - `public/updateInfos.json`: 「0.26.3 / 一部環境で起動しない問題を修正」 の
    1 entry 追加 (= アプリ内更新履歴の表示データのみ)
  - `public/qAndA.md`: 文言 1 行修正
  - **`package.json` / `pnpm-lock.yaml` は無変更**。 `package.json` の
    `electron` は `42.11.18` のまま (0.26.3 tag の同 file を直接確認)、
    `scripts` の `preinstall` / `postinstall` にも変更なし (= install script
    の新設なし)
- [x] `prepare()`: 手順は 0.26.2 と同一 (split 2 volume を `cat` で物理結合 →
    `7z x` → glob で AppImage を拾い、 内部から 3 file のみ再抽出)。
    source の download 名を version 付きにしたのに合わせ、 結合対象を
    `voicevox-${pkgver}-cpu-x64.AppImage.7z.00{1,2}` として明示
    (glob ではなく変数で固定)。 シェル injection・eval・curl・wget なし
- [x] `package()`: `install` のみ、 ネットワーク操作なし (0.26.2 から無変更)
- [ ] **`resources/app.asar` の中身は本 run では未確認**
  - 0.26.2 時点では `voicevox-linux-0.26.2.tar.gz` の同 file で native module
    (`.node`) 0 個 / `postinstall`・`preinstall` 文字列なしを確認済み。
    0.26.3 では当該 tar.gz も release asset のため取得できず、 再確認できて
    いない。 ただし upstream の差分が `package.json` を触っていない
    (= 依存追加なし) かつ build config 無変更なので、 asar の構造・install
    script が変わる経路は考えにくい。 Phase 3 の build で pkg 内容を確認する
- [x] depends: `bash` / `electron42` (+ makedepends `7zip`) — **無変更**。
  upstream の electron pin は 42.11.18 のまま (上記 `package.json` 確認) で、
  Arch 公式 extra の `electron42` 42.11.18-1 と一致するため `.deps.lock` も
  変更なし
- [x] optdepends: voicevox-engine — 0.26.2 から無変更。 voicevox-engine-cuda
  (`provides=voicevox-engine`) が満たし、 ayaka 単独 install 時は network 経由
  (Tailscale 100.117.10.53:50021) で remote engine を指定する運用も不変

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| split package (`pkgname=(voicevox-bin voicevox-appimage)`) を解消、 `voicevox-bin` のみに | [nekono] では editor 1 variant (= system electron42 経由) に絞る。 AppImage 全体を抱える voicevox-appimage は配布対象外 (ayaka では electron42 を共通で持っている方が他 Electron アプリと共有できて軽い) |
| AUR の `if [ -c /dev/nvidia0 ]; then source=(...NVIDIA AppImage...)` を完全削除 | **重大な supply-chain 欠陥**。 build host の GPU 有無で source を動的切替する非決定 PKGBUILD で、 NVIDIA 版 AppImage 分の sha256 は pin されていなかった (= GPU 環境で build すると無検証 binary が package に混入)。 ayaka は editor だけ要り engine は remote (= nekono-pacman0 上の voicevox-engine-cuda) なので、 editor の GPU 対応は不要 → CPU AppImage 固定で十分 |
| `optdepends` の wording を「local もしくは remote、 EngineSettings で URL 指定」に改善 | engine が必ずしも同 host にある前提を消し、 remote engine 運用 (= nekono-pacman0 → ayaka) を documentation 化 |
| system electron を `electron37` から `electron42` へ変更 (0.26.1 で追加) | upstream が 0.26.0 で electron を 42.1.0 に上げ、 0.26.1 で 42.11.18 pin。 配布 AppImage 自身も Electron 42 系を積む。 Arch 公式 extra に pin 一致の `electron42` 42.11.18-1 があるため [nekono] 自家 build は不要。 AUR は 0.25.2 止まりのため AUR 側にはこの変更はまだ現れていない |
| source に version 付き download 名 (`::` 形式) を付与 (0.26.3 で追加) | upstream asset 名に version が無く、 makepkg の download cache が bump をまたいで再利用される。 実際に 0.26.3 検証で 0.26.2 の cache を掴んで checksum mismatch になったため、 名前で世代を分離した |
| checksum を b2sums → sha256sums に変更 (0.26.3 のみ) | 本 run の harness が b2sum を実測できない (shell なし / release asset の cross-host redirect が allow-list 外)。 GitHub API の asset digest + `makepkg --verifysource` の実 download 照合で値を確定した。 詳細は「検証結果」 参照 |

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可。
ayaka 側にも [nekono] 経由で `pacman -S voicevox-bin` で入れて editor を起動、
初回起動時に EngineSettings の API URL を `http://100.117.10.53:50021`
(nekono-pacman0 の Tailscale IP) に向ける。 0.26.1 以降 runtime 依存は
`electron42` (Arch 公式 extra、 0.26.3 でも pin 42.11.18 は不変)。

残課題 (build 時に確認する): AppImage 内部 path と `resources/app.asar` の
構造は本 run で再取得できていない。 Phase 3 の `bin/build-all` で
prepare() が通ること + pkg 内容 (`/usr/lib/VOICEVOX/resources/app.asar` /
`/usr/bin/voicevox` / desktop entry) を確認する。

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
| 2026-11-06 | 0.26.2 | (this PR) | `b7ad1169ab4a0a7a574312b2de63eb8829c00b1d` | Issue #778 (`Closes #778`)。 0.26.1 → 0.26.2。 part 001/002 の sha256 を実測し (part 001 `e1e64ff4...` / part 002 `834de350...`)、 **upstream 公開 manifest `linux-cpu-x64-appimage.7z.txt` と GitHub Release asset digest の両方に一致**。 b2sums は実測値を pin (`874678d5...` / `6d83fcf2...`)、 `makepkg --verifysource` 通過。 part 構成は 1,992,294,400 + 123,413,662 byte (part 002 が 62.6 MB → 123.4 MB に増加 = AppImage 本体の増分)。 upstream 差分は 4 file のみ (`build.yml` の engine/resource version 0.26.0→0.26.1、 `package.json`/`pnpm-lock.yaml` の playwright 1.60.0→1.61.0、 `updateInfos.json` の新キャラ「ブイブイ・コバコ」) で、 build/electron pin (`42.11.18`) / install script / native module に変更なし。 app.asar は `voicevox-linux-0.26.2.tar.gz` の同 file で構造確認 (`.node` 0 個 / `postinstall`・`preinstall` 無し)。 **depends は無変更** (Arch extra の electron42 も 42.11.18-1 のまま → `.deps.lock` 変更なし)。 7z が本 run の harness で実行できず AppImage 内部 listing は再取得できなかったため、 electron-builder config 無変更 + linux tar.gz の asar 構造 + Phase 3 の実 build で裏付ける方針 |
| 2026-11-09 | 0.26.2-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): 7zip 26.05-1 → 26.06-1 |
| 2026-11-16 | 0.26.3 | (this PR) | `ffa4802e49df7d62f930c66373c442ef0deeda4d` | Issue #819 (`Closes #819`)。 0.26.2 → 0.26.3。 **checksum を b2sums から sha256sums に切り替えた** (本 run の harness は shell を持たず `b2sum` を実行できず、 `fetch` も GitHub release asset の cross-host redirect で取れないため、 upstream 公開 manifest も読めない)。 pin 値は GitHub Releases API の asset digest (`sha256:0e6fa43b...` / `sha256:8b6d9698...`) と一致し、 **fresh download に対する `makepkg --verifysource` の照合も通過** (rc=0)。 part 構成は 1,992,294,400 + 123,856,375 byte。 **source の download 名を version 付き (`::`) に変更** — upstream asset 名に version が無いため makepkg が前回 cache を再利用し、 実際に本 bump の検証で 0.26.2 の cache を掴んで checksum mismatch になった (放置すると古い bytes を新 version の正として pin する事故)。 upstream 差分は 3 commit / 3 file (`build.yml` の engine version 0.26.1→0.26.2 + macOS runner macos-14→15、 `updateInfos.json` の 0.26.3 entry、 `qAndA.md` の文言) で、 **`package.json` / `pnpm-lock.yaml` は無変更 → electron pin 42.11.18 不変 → depends 無変更 → `.deps.lock` 変更なし**、 install script の新設もなし。 AppImage 内部 listing と app.asar 構造は本 run では再取得できていない (7z 不可 / tar.gz も release asset) ため、 build config 無変更を根拠に path 不変と判断し、 実確認は Phase 3 の `bin/build-all` に委ねる |
