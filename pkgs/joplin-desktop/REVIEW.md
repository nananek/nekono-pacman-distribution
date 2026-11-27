# joplin-desktop review

## 状態

**review 済み / user 承認待ち** (2026-11-27 / 3.7.24、初回)

AUR の `joplin` (pkgbase=joplin, pkgver=3.7.24, pkgrel=1, AUR commit
`5974aab` = 2026-11-19) を fork。 意図的改変 5 件 (split 解消 / check() 削除 /
`options=(!strip !debug)` / arch を x86_64 のみに / `yarn install --immutable`)。

## 用途

デスクトップのノート / ToDo アプリ。 ayaka で AUR 版の代わりに `[nekono]`
から install する。 CLI (`joplin`) は配布しない。

## Source

- AUR: https://aur.archlinux.org/packages/joplin
  - Maintainer: CUI Hao <cuihao.leo@gmail.com> (2025-05-18 から。 AUR 初回
    submit は 2018-04-18 = `masterkorp`)
  - fork 元 AUR commit: `5974aab` (`upgpkg: 3.7.24`, 2026-11-19)
  - AUR 履歴は 2025-06 以降すべて CUI Hao の単独 commit で、 source URL の
    domain 変更 / orphan 化 / maintainer の不審な交代は無い。 electron の
    major は 37→38→39→41→42 と upstream の pin に追随している
    (`_electronVersion`)。
- Upstream:
  - https://github.com/laurent22/joplin
  - AGPL-3.0-or-later
  - release tag `v3.7.24` の commit SHA:
    `20301fdae662c52eae22ba6b2f008e83a00fd5bf`
  - desktop の electron pin: `packages/app-desktop/package.json` の
    devDependencies `electron: 42.11.18`。 Arch 公式 extra の `electron42`
    (42.11.18-1) と**完全一致**するため自家 electron は不要
    (= voicevox-bin と同じ判断)

## 検証結果

- [x] source URL = `github.com/laurent22/joplin/archive/v3.7.24.tar.gz`
  - upstream 公式 GitHub の tag tarball (codeload redirect)。 typosquat 無し。
    実 download して sha256 を実測 = `8d9de9ef14528259f182883a90f63f3ad4d7d44011e2c278496f958cb73cdf4d`
    (= AUR の pin と byte 一致、155,468,766 bytes)
- [x] `.desktop` / wrapper は **package() 内で生成** (source に置かない)
  - `.gitignore` が `pkgs/*/*.sh` と `pkgs/*/*.desktop` を除外するため、
    ローカルファイルとして commit できない (前例: docker-rootless-extras は
    URL source、 voicevox-bin は `run.sh` を生成)。 AUR の該当 2 ファイルを
    読んで内容を確認し、 package() に同内容を置いた
  - 実測: AUR の `joplin-desktop.sh` (`@electronversion@`→42) と生成物の
    `diff` は **完全一致**、 `.desktop` も **完全一致**
  - `.desktop` の `StartupWMClass=appimagekit-joplin` は upstream PR #16547 に
    合わせた値。 wrapper は stock electron wrapper と同じ flags 読み込みで
    `/usr/lib/joplin-desktop/app.asar` を exec するだけ
- [x] `prepare()` / `build()` / `package()` に curl / wget / eval /
  pipe-to-shell は無い
- [!] **build 内で行う動的取得 (AUR と同じ、意図的に受入)**:
  - `npx yarn install` が npm registry から依存を取得 (yarn.lock で固定。
    `--immutable` を追加し lockfile 外の解決を build 失敗にした)
  - `@electron/rebuild` / node-gyp は該当時のみ electron headers を取得
  - `.yarnrc.yml` に `approvedGitRepositories: https://github.com/electron/node-gyp.git`
    があり、 node-gyp だけは git から取得する。 upstream の設定で、 AUR も同じ
- [x] depends の根拠
  - `electron42`: build 時に `-c.electronDist=/usr/lib/electron42` で同梱
    electron を作らず system のものを使う。 **実行時依存**
  - `libvips`: sharp は `SHARP_IGNORE_GLOBAL_LIBVIPS=1` で prebuilt libvips を
    同梱するため実際には不要の可能性が高いが、 AUR に合わせて残す
    (外すと挙動が変わるリスクを取らない)
  - native dep は `sqlite3 5.1.6` / `sharp 0.34.5` / `keytar 7.9.0`。 いずれも
    N-API prebuilt で ABI 非依存のため、 system electron42 で electron-rebuild
    無しに動く (upstream の `yarn dist` は `yarn electronRebuild`
    (`--force-abi 146` = Electron 42 の ABI) を挟むが、 AUR は省いている。
    AUR で動いている実績と N-API であることから現状は不要と判断。
    **bump 時に native dep が N-API を外したら electronRebuild を追加する**)
- [x] `*.install` は無し。 post-install hook なし
- [x] `.git` は含めない。 secret 混入なし
- [x] ローカルで build()〜package() 相当を実測 (2026-11-27、 125 GB / 8 core)
  - `yarn install`: 1m15s (yarn cache 温あり。 冷は約 4 分)
  - `gulp before-dist`: 3s
  - `electron-builder --linux dir`: 39s → `dist/linux-unpacked/resources` 309 MB
    (app.asar 222 MB + app.asar.unpacked 40 MB)
  - 生成物レイアウトで electron 42.11.18 を xvfb 上に起動 → エラー無し、
    `database.sqlite` が生成された (= sqlite3 の native module がロードできた)
  - `app.asar.unpacked` の native `.node` は sqlite3 (`napi-v6`)、
    onnxruntime (`napi-v6`)、 keytar (`napi-v3`)。 3 つとも undefined symbol に
    `napi_` を含み **N-API** = electron の ABI に依存しない
    → electron-rebuild を省く AUR の判断は妥当 (実測で裏付け)
- [ ] **8 GB build host での実測はまだ** (この検証は 125 GB のホスト)。
  `check()` を外した状態で足りるかは build host で `bin/build-all <pkg>` の
  初回に確認する
- [ ] `check()` (jest) は 48 GB 要求のため未実行。 CI に任せる
- [x] keytar の Linux 有効化 patch (2026-11-28): `packages/lib/shim.ts` の
  `platformSupportsKeyChain()` を sed で Linux 許可に変更する。 sandbox
  (`bin/joplin` + `--talk=org.freedesktop.secrets`) 内で keytar の
  `setPassword` / `getPassword` / `deletePassword` が通り、`mytest` を
  読み戻せることを実測した。 Joplin 本体の `KeychainService` 検出
  (`keychain.supported=1`) は build 後の実機で確認する。

## 依存方針 (AUR との意図的な diff)

1. **split package の解消**: AUR は pkgbase=`joplin` で CLI (`joplin`) と
   `joplin-desktop` を吐く。 [nekono] は desktop のみ。 pkgbase / dir /
   nvchecker section を `joplin-desktop` に統一 (= voicevox-bin の前例)。
   あわせて build() で `packages/app-cli` も削除する。 どの package も
   `@joplin/app-cli` / `joplin` を依存に持たないため (= app-desktop の
   @joplin 依存は default-plugins / editor / lib / onenote-converter /
   renderer / tools / utils のみ)、 CLI の tsc build と、 それが使う
   `rsync` (AUR makedepends) を丸ごと落とせる。
2. **`check()` の削除**: upstream app-desktop の jest テストは AUR コメント
   (2026-09-24, pfrenssen) で「48 GB のメモリ (swap 込み) が必要。 足りないと
   OOMKiller が落とす」と報告されている。 `bin/build-all` は
   `makepkg -sc` で check() を自動実行するため、 build host (8 GB) では
   daily-ops が回らない。 テストは upstream CI が担保する。
3. **`options=(!strip !debug)`**: asarUnpack される native `.node` を strip
   せず debug pkg も作らない (= voicevox-bin / voicevox-engine-cuda と同じ)。
4. **`arch=('x86_64')` のみ**: `[nekono]` は `repo/x86_64/` のみ配信。
   AUR の aarch64 分岐 (`dist/linux-arm64-unpacked`) は削除。
5. **electron-builder を dir target のみに**: AUR の `--dir=dist/` は
   boolean フラグに値が付いておらず target を絞れていない (= 実測で
   AppImage と deb も build され、 fpm が `libcrypt.so.1` 不足で落ちた)。
   `--linux dir -c.directories.output=dist` に変え、 dir target では
   生成されない AppImage の checksum hook (`afterAllArtifactBuild`) を
   package.json から 1 行外す。 効果: ビルド 39 秒、 fpm/ruby 不要、
   `libxcrypt-compat` も makedepends から外せる。
6. **`yarn install --immutable` は使わない**: build() 冒頭で workspace を
   rm するため lockfile からそれらの entry が落ち、 `--immutable` は YN0028
   で必ず失敗する (実測: 「The lockfile would have been modified by this
   install」)。 AUR と同じ素の `yarn install` に戻した。 解決は yarn.lock
   準拠で、 書き換わるのは削除した workspace 分だけ。
7. **makedepends から `rsync` / `libxcrypt-compat` を外した**: rsync は
   app-cli の build だけが、 libxcrypt-compat は fpm (deb 生成) だけが
   使っていた。 どちらも消えたので不要 (実測: 両方無しで dir build が通る)。
8. **Linux でも keytar/libsecret を有効化**: `packages/lib/shim.ts` の
   `platformSupportsKeyChain()` を `shim.isLinux()` も true を返すように
   build() で sed patch する。 upstream が Linux を無効化している理由は
   2019 年の sharp/libz 衝突 (コメントに残っている) で、 keytar 7.9.0 は
   N-API prebuilt、 system electron42 と pass-secret-service の組み合わせ
   では再現しない。 これにより `secure: true` の設定
   (`encryption.masterPassword` / `encryption.passwordCache` /
   `sync.9.password` 等) は `settings` テーブルではなく
   `org.freedesktop.secrets` に保存される。 sandbox wrapper
   (`~/.dotfiles/bin/joplin`) 側で `--talk=org.freedesktop.secrets` を
   許可する。 **注意**: Secret Service API には per-app ACL が無く、
   Joplin から default collection 全体に到達できる (bin/geary と同じ制約)。
   upstream の該当行が変わった場合は build を fail させる grep gate を入れて
   あるので、 bump 時に sed を追随させること。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-11-27 | 3.7.24 | (this PR, 検証時 `4a55c9f2a9f89c6cad287e31f3fa5fbeef75b438a2529b58ba54fc7ba3683a2a`) | `20301fdae662c52eae22ba6b2f008e83a00fd5bf` | 初回 review。 sha256 実測一致、 build〜起動まで実測、 AUR 改変 6 件で受入 (user 承認待ち) |
| 2026-11-28 | 3.7.24-2 | (this PR) | `20301fdae662c52eae22ba6b2f008e83a00fd5bf` | keytar/Linux 有効化 patch を追加 (E2EE password を Secret Service へ)。 keytar の set/get/delete は sandbox 内で実測済み。 Joplin 本体の keychain 検出は build 後確認 |
