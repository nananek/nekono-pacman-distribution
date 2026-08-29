# github-copilot-cli review

## 状態

**review 済み、approve** (2026-08-11、AUR fork + 供給源 vendor 改変)

AUR の `github-copilot-cli` PKGBUILD (pkgver=1.0.77, pkgrel=1) を fork し、
**build 時の npm registry 動的取得を排除する改変** を加えた。詳細は
「依存方針 (AUR との意図的 diff)」参照。

## Source

- AUR: https://aur.archlinux.org/packages/github-copilot-cli
  - maintainer: Rafael Dominiquini `<rafaeldominiquini at gmail dot com>`
  - co-maintainer: edu4rdshl `<edu4rdshl at protonmail dot com>`
- Upstream: https://github.com/github/copilot-cli (= GitHub org、official)
- npm:
  - `@github/copilot` (= main package、実体は `npm-loader.js` の薄いラッパー)
  - `@github/copilot-linux-x64` (= platform package、native binary を同梱)
  - `detect-libc` (= main package の `dependencies`、`^2.1.2`)
- 配布経路: npm registry のみ (= AUR も npm tarball を使用、typosquat の
  懸念なし)。license は `LicenseRef-GitHub-Copilot` (= LICENSE.md 同梱)

## 検証結果

- [x] `source` URL が全て upstream official path
  - `https://registry.npmjs.org/@github/copilot/-/copilot-${pkgver}.tgz`
  - `https://registry.npmjs.org/@github/copilot-linux-x64/-/copilot-linux-x64-${pkgver}.tgz`
  - `https://registry.npmjs.org/detect-libc/-/detect-libc-${_detectlibc_ver}.tgz`
  - `https://raw.githubusercontent.com/github/copilot-cli/v${pkgver}/changelog.md`
  - いずれも official (npm registry / github 公式 raw)、 redirect ・
    typosquat なし
- [x] `b2sums` 独立検証 (2026-08-11、registry から実 download して照合)
  - `copilot-1.0.77.tgz`: `770cc8c8…` 一致
  - `copilot-linux-x64-1.0.77.tgz` (≈181MB): `fa5b78ed…` 一致
  - `detect-libc-2.1.2.tgz`: `b036f6d4…` 一致
  - `CHANGELOG-1.0.77.md`: `3699bdb9…` 一致
  - SKIP エントリなし (全 source pin 済み)
- [x] upstream tag 検証: `v1.0.77` = commit
  `aee1edd29ef0f2058425bf399bcc9e5002a2b8f2` (github.com/github/copilot-cli)
- [x] `@github/copilot@1.0.77` の依存
  - `dependencies: { "detect-libc": "^2.1.2" }` → vendor の 2.1.2 で充足
  - `optionalDependencies`: 8 platform package (`-linux-x64` / `-linux-arm64` /
    `-linuxmusl-*` / `-darwin-*` / `-win32-*`) が全て `1.0.77` で version sync。
    linux-x64 以外は install されない (optional dep は解決失敗で skip)
  - `scripts` 無し (postinstall 等の動的実行なし)
- [x] `@github/copilot-linux-x64@1.0.77` (platform package)
  - `bin: { "copilot-linux-x64": "copilot" }` → 実 binary
  - `scripts` 無し (install 時 network fetch なし)
  - 中身: `prebuilds/` (native .node)、`ripgrep/`、`clipboard/`、
    `foundry-local-sdk/`、`pvrecorder/`、`webview/`、`voice-*` 等を同梱
    (= 本体実装の大半はここ)
- [x] `package()` の offline npm install を **実地検証** (2026-08-11、
  fakeroot 無しの dry-run 相当: `npm install --offline` を同 3 tarball で
  実行し成功、`usr/bin/copilot` + `node_modules/{@github/copilot,
  @github/copilot-linux-x64, detect-libc}` が正しく配置されることを確認)
  - `--offline` + `--cache` で registry 参照ゼロ (= 想定どおり)
  - `copilot --version` = `GitHub Copilot CLI 1.0.77.`、completion 生成も動作
  - `usr/bin/copilot-linux-x64` (platform package 由来の bin entry) が生成
    される → PKGBUILD の `rm -f` は正当な対処
- [x] `build()`: 無し (= package() のみ、AUR と同じ)
- [x] `package()` に curl / wget / pipe-to-shell / eval / pip install なし
  (= npm install --offline + 自前 install のみ)
- [x] `depends`: `glibc` `gcc-libs` `nodejs` `glib2` `libsecret` (= AUR の
  まま、native 依存として妥当)
- [x] `makedepends`: `npm` `jq` (= AUR のまま。`jq` は package.json の
  npm 注入 field strip に使用、depends への重複なし)
- [x] `noextract`: 3 tarball を対象 (= npm が直接読み込むため展開不要、
  CHANGELOG のみ makepkg 展開)
- [x] package() の cleanup step が platform package の実レイアウトと一致
  (= `prebuilds/linux-x64`、`ripgrep/bin/linux-x64`、
  `clipboard/node_modules/@teddyzhu/{clipboard,clipboard-linux-x64-gnu}`、
  `foundry-local-sdk/…/prebuilds/linux-x64`、`pvrecorder/…/lib/linux/x86_64`
  を実 tarball で確認。`mxc-bin` は 1.0.77 に存在しないが `[ -d ]` guard
  付きなので無害)
- [x] `options=(!strip emptydirs staticlibs zipman)` — prebuilt binary 同梱
  のため `!strip` は必須 (= CLAUDE.md「prebuilt .so 同梱 pkg の options」)
- [x] secrets 混入なし
- [x] arch: `x86_64` のみ (= aarch64 は `@github/copilot-linux-arm64` を
  vendor しないと不可。コメントに明記済み)

## 依存方針 (AUR との意図的 diff)

AUR 版は `npm install` を main tarball にのみ実行し、build 時に npm の
dependency resolver が `detect-libc` + 実 binary を積んだ
`@github/copilot-linux-x64` (~170MB) を **registry から動的取得** する
(= build 時に un-pinned download、supply-chain 上 [nekono] 不可)。

改変内容:

1. **3 tarball を source[] に vendor + b2sums pin**
   (`@github/copilot` / `detect-libc` / `@github/copilot-linux-x64`)
2. **`npm install --offline`** で完結 (registry 参照ゼロ)。3 tarball を
   明示的に top-level install target にすることで npm が sibling として
   hoist し、optionalDependencies 解決が vendor 済み tarball だけで成立
3. **`rm -f "${pkgdir}/usr/bin/copilot-linux-x64"`**: platform package 自身の
   bin entry (raw native executable) を除去し、public コマンドは
   `copilot` (npm-loader.js wrapper) のみにする
4. **cleanup step の対象を `_platdir` (= copilot-linux-x64) に変更**:
   AUR 版は `_moddir` (= 薄い wrapper package) に対して prebuilds cleanup
   していたが、実体は platform package 側にあるため。`[ -d ]` guard 付きで
   将来のディレクトリ構成変化に耐える
5. **`jq` による `_`-prefix field strip を 3 module に拡大**
   (AUR 版は main のみ。detect-libc / platform package にも npm が
   `_`-field を注入しうるため)
6. `mxc-bin` cleanup の CARCH 分岐削除: `arch=(x86_64)` 固定なので不要

## 結論

**approve** — build host で `makepkg -s --sign` 可 (実地 dry-run 済み:
offline install / bin entry / completion 生成まで確認)。

## 更新方針

upstream の新 release が出たら:

1. `nvchecker -c nvchecker.toml` の `[github-copilot-cli]` が新 version を
   返すか確認 (= npm latest dist-tag)
2. `@github/copilot@<ver>` の `dependencies` / `optionalDependencies` を
   registry API で確認 (detect-libc の vendor 要否 / platform package の
   version sync)
3. 3 tarball を registry から download して b2sums を**独立再計算**
   (CHANGELOG も同様)。platform tarball のレイアウト変化
   (prebuilds/ripgrep/clipboard/...) を確認し cleanup step を更新
4. `pkgver` / b2sums を差し替え、`.SRCINFO` を
   `makepkg --printsrcinfo` で再生成、REVIEW.md 更新履歴に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-08-30 | 1.0.81-1 | (this PR) | `v1.0.81` (GitHub Release, 2026-08-27) | safe-to-bump: Issue #577 調査済み。3 source (`@github/copilot` / `@github/copilot-linux-x64` / CHANGELOG) を npm registry から独立再取得、b2sums 完全一致確認 (`detect-libc` は 2.1.2 のまま version 据え置き、既存 pin と再検証一致)。release note は plugins dashboard 一般公開、MCP 2026-07-28 対応、OTel trace context 付き hooks 等の機能追加のみ、security/breaking なし。platform tarball のレイアウト確認: `prebuilds/linux-x64`、`foundry-local-sdk` prebuilds、`ripgrep/bin/linux-x64`、`pvrecorder-node` は健在。`mxc-bin` / `clipboard/node_modules/@teddyzhu` は tarball から消失 (upstream 依存整理と推測、package() 側は `if [ -d ... ]` guard 済みで no-op化するだけ、build 影響なし)。Closes #577。
| 2026-08-15 | 1.0.80-1 | (this PR) | — (deps.lock sync) | .deps.lock 更新: gcc-libs 16.1.1+r595+g171d15ac6959-1 → 16.2.1+r23+gd564253eb6c8-1, glibc 2.44+r5+g7cba77790f32-1 → 2.44+r24+g16be1518495f-1 (pkgrel-2 PR #524 は pkgver bump で superseded のため close) |
| 2026-08-14 | 1.0.80-1 | (this PR) | `ef627e1b` (v1.0.80) | safe-to-bump: 機械的 pkgver bump、npm package 1.0.77→1.0.80 diff 済み (新規は SKILL.md / doc / .d.ts のみ、実行コード変更なし)、build/depends/makedepends 変更なし、maintainer 1 名削除 (publish は OIDC 経由のため risk 無し)、b2sums 独立検証済み。Closes #541 |
| 2026-08-11 | 1.0.77-1 | (初版) | `aee1edd2` (v1.0.77) | 新規追加。AUR fork + npm 依存の offline vendor 化。4 source の b2sums を registry download で独立検証、offline install / bin / completion を実地確認。approve |
