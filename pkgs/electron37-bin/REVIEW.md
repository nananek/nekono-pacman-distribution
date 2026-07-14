# electron37-bin review

## 状態

**review 済み、approve** (2026-07-14)

AUR の `electron37-bin` PKGBUILD (pkgver=37.10.3, pkgrel=1) を fork。
`arch=('x86_64')` のみに絞る (= [nekono] は `repo/x86_64/` のみ配信、
aarch64/armv7h source/sha256sums block を削除) 以外は無改変。

## 追加の経緯

`voicevox-bin` (VOICEVOX editor) は `depends=(electron37)` で system electron37
経由起動する構成 (= AUR 本家と同じ方針)。 Arch 公式 `extra` repo はローリング
更新で electron37 を既に落としており (`extra` に残るのは electron39〜43 のみ、
2026-07-14 build host migration で発覚)、`voicevox-bin` の build が
`sudo pacman -S electron37` 解決失敗で fail していた。

VOICEVOX 0.25.2 (= 現行最新 release、2026-04-30 公開) の
`package.json` を確認したところ `"electron": "37.7.1"` に固定されている
(https://raw.githubusercontent.com/VOICEVOX/voicevox/0.25.2/package.json)。
upstream に electron39+ へ bump した新 release は存在しないため、
「新しい electron に voicevox-bin 側を追従させる」という解決は取れない
(= major version 越しの Electron は Node ABI / API 非互換のリスクがあり、
upstream が実際に build/test した version に合わせるのが安全)。

AUR には Arch 公式から落ちた electron37 を追う 2 系統が存在:
- `electron37` (= Chromium を git source から自前 build、 makedepends に
  clang/gn/ninja/rust 等、 実質 Chromium full build = build host での
  自動 build には非現実的な時間・リソースを要する)
- `electron37-bin` (= electron/electron 公式 GitHub Release の prebuilt
  zip を re-package するだけ、 makepkg 上は軽量)

[nekono] は `electron37-bin` を採用。 `provides=(electron37=37.10.3)` を
持つため、`voicevox-bin` 側の `depends=(electron37)` PKGBUILD は無改変で
そのまま解決できる。

## 検証結果

- [x] `source_x86_64` URL = `github.com/electron/electron/releases/download/v37.10.3/{chromedriver,electron}-v37.10.3-linux-x64.zip`
  — electron org 公式 GitHub repo の Release asset、typosquat / domain spoof
  無し。`gh api repos/electron/electron/releases/tags/v37.10.3` で release
  author `sudowoodo-release-bot[bot]` (electron 公式リリース bot) 確認済み。
- [x] `sha256sums_x86_64` は AUR 記載値をそのまま信頼せず、build host で
  両 zip を直接 `curl | sha256sum` で独立実測し **完全一致確認**:
  - chromedriver zip: `ced6d8721ce57a3fa10d2bc614e4d49ab031c46629ed5af03a253ce7def8b747`
  - electron zip: `c0b4edd6bd9858cda4cf7ab299e69a2d3ecd2e5fcca78507bc0851ba35614660`
- [x] `prepare()` / `package()`: zip を bsdtar 展開 → `usr/lib/electron37` に
  そのまま配置するだけ、curl/wget/eval/pipe-to-shell 等の追加無し。
- [⚠] `chmod u+s "${srcdir}/${_pkgname}/chrome-sandbox"`: Chromium
  sandbox helper に setuid root を付与する行。**Electron/Chromium の
  sandbox 機構が要求する標準的な挙動** (= 公式 Arch `extra/electron*` 系
  package も同じことをしている)。setuid 対象は electron 本体が展開時に
  含む `chrome-sandbox` バイナリのみで、他ファイルには影響しない。
  supply-chain 上の懸念ではなく Chromium sandbox の既知の要件。
- [x] `depends`: `alsa-lib` / `gtk3` / `nss` — Electron ランタイムの
  audio/GUI/crypto 依存として妥当。
- [x] `provides=(electron37=37.10.3)` / `conflicts=(electron37)` —
  Arch 公式 `extra/electron37` が存在した頃と同名 provide を維持する
  設計、`voicevox-bin` 等の `depends=(electron37)` から見て透過的に置換
  可能。
- [x] license: `MIT` + `LicenseRef-custom` (Chromium 由来の各種 license
  file 同梱、`install -Dm644 LICENSE*` で反映) — upstream 通り。

## 依存方針 (AUR との意図的 diff)

- `arch=('x86_64')` のみに削減 (AUR 本家は aarch64/armv7h も persist)。
  [nekono] は `repo/x86_64/` のみ配信するため aarch64/armv7h の
  source/sha256sums block は不要、削除。

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可。
`voicevox-bin` の `depends=(electron37)` を本 pkg の `provides` が満たす。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-07-14 | 37.10.3-1 | (this commit) | `v37.10.3` (electron/electron 公式 Release) | 初回 add。voicevox-bin の electron37 depends が Arch 公式 repo から落ちて resolve 不能になっていた問題への対応。sha256 独立実測で AUR 記載値と一致確認。 |
