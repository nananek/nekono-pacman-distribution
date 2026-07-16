# aria-bin review

## 状態

**review 済み、approve** (2026-07-14)

AUR に `aria` / `aria-bin` は存在しないため、本 repo オリジナルの PKGBUILD として
新規作成 (= AUR からの fork ではない)。

## 背景

- upstream: https://github.com/poppingmoon/aria (poppingmoon 氏、AGPL-3.0-only、
  cross-platform Misskey client、Flutter 製)
- aria は https://github.com/shiosyakeyakini-info/miria の fork。miria は AUR に
  `miria-bin` として登録済み (maintainer: 4sterisk、友人)。ただし AUR の
  `aria` は無関係の別プロジェクト (実験的 C 系言語処理系、maintainer: shkhuz)
  であり、`aria-bin` は未登録。
- PKGBUILD の設計は `miria-bin` (= 「release artifact を bsdtar 相当で展開する
  だけ」という構造) を参考にしたが、aria の Linux release 形式が miria と異なる
  ため package() の実装は書き下ろし (詳細は「AUR (miria-bin) との相違点」参照)。

## Source

- Linux release tarball: `aria-v${pkgver}-linux-x64.tar.gz` / `...-linux-arm64.tar.gz`
  - URL: `https://github.com/poppingmoon/aria/releases/download/v${pkgver}/...`
  - upstream 公式 org `poppingmoon/aria` の release artifact、typosquat / mirror
    spoof なし
- `.desktop` / icon (`.png`) / appstream metainfo (`.xml`):
  - upstream repo 内 `flatpak/com.poppingmoon.aria.{desktop,png,metainfo.xml}` を
    同一 tag (`v${pkgver}`) の `raw.githubusercontent.com` 経由で個別に vendor
  - 自前生成ファイルではなく upstream 公式ファイルそのもの (flatpak 配布用に
    upstream が用意している成果物を linux tarball 版パッケージングにも転用)

## 検証結果

- [x] `source_x86_64` / `source_aarch64` URL = `github.com/poppingmoon/aria`
      本家 release、typosquatting 無し
- [x] `sha256sums` / `sha256sums_x86_64` / `sha256sums_aarch64` が実測値と一致
  - `aria-v1.5.8-linux-x64.tar.gz`: `42c1652bf5a3ccae63477079794d266d7e9e38525b8d0cda1aaa942ba5f1c005`
  - `aria-v1.5.8-linux-arm64.tar.gz`: `ed22aeb6e3aed0ad5b0dafa04cf5fb4e49d300d64be9a8e21c29aabc1a35fd04`
  - `com.poppingmoon.aria.desktop`: `b774b7ea50393e78016c3a8bfa0b26b480d3a32de8121c9694e97aaccb7e2342`
  - `com.poppingmoon.aria.png`: `878b0a27b7706036a2ddcb41bab05bf65ba9dd4718cabc25961f45daa5ab8ace`
  - `com.poppingmoon.aria.metainfo.xml`: `872b20c997e2ea35dbcdd106c97be92852fbd83a41b547b7618a0d55d11aa358`
  - `makepkg --verifysource` で全 source の sha256 検証成功済み
- [x] tag `v1.5.8` の git commit (`dbce179b61f597b43aca4e8b0f63a5c8f079adf7`) は
      **GPG verified** (author: poppingmoon 本人)。release は
      `github-actions[bot]` による CI 公開 (= verified commit からの自動 build
      pipeline、tampering の兆候なし)
- [x] `package()`: `install -dm755` / `cp -a` / `ln -s` / `install -Dm644` の
      標準コマンドのみ。network fetch / eval / curl / pip 等の動的取得なし
  - `/opt/aria/{aria,data,lib}`: upstream tarball の中身をそのまま配置
  - `/usr/bin/aria`: `/opt/aria/aria` への symlink
  - `/usr/share/applications/com.poppingmoon.aria.desktop`: `Exec=aria` /
    `Icon=com.poppingmoon.aria` が既にバイナリ名と一致しており sed 改変不要
  - `/usr/share/icons/hicolor/512x512/apps/com.poppingmoon.aria.png` (実測
    512x512 PNG)
  - `/usr/share/metainfo/com.poppingmoon.aria.metainfo.xml`
- [x] `depends`: 実 binary (`aria` および `lib/*.so`) を `ldd` (`LD_LIBRARY_PATH`
      に `lib/` を追加) で検証し、system 側で解決が必要なライブラリを特定
  - `gtk3` (GTK/pango/cairo/gdk-pixbuf/atk/X11/wayland 系を推移的に充足)
  - `libsecret` (`flutter_secure_storage` プラグインの keyring backend)
  - `gst-plugins-base-libs` (`libgstapp`/`libgstbase`/`libgstreamer` 系、
    gtk3 の依存には含まれず明示が必要)
  - `lib/libmdk.so.0` + `lib/libffmpeg.so.8` はバンドル同梱 (self-contained
    media backend) のため **system `mpv` は不要** (= miria-bin との相違点)
  - `lib/libdartjni.so` は `libjvm.so` を参照するが `aria` バイナリの依存閉包
    には含まれない (Android 向け JNI bridge の残骸、Linux desktop 実行では
    未使用と確認)
- [x] `options=(!strip !debug)`: `lib/libapp.so` は既に strip 済み、
      `aria` 本体 / `lib/libmdk.so.0` 等は not stripped 混在。CLAUDE.md
      「pitfall #6」に従い makepkg の自動 strip/debug package 生成を抑止
- [x] license `AGPL-3.0-only` — upstream `LICENSE` (AGPLv3 全文) および
      `flatpak/com.poppingmoon.aria.metainfo.xml` の `<project_license>` と一致
- [x] build host (このマシン) で `makepkg -f` を実行し実際に build 成功、
      生成された `.pkg.tar.zst` の中身 (`opt/aria/`, `usr/bin/aria` symlink,
      `usr/share/applications`, `usr/share/icons`, `usr/share/metainfo`) を
      目視確認

## 未検証事項 (制約による)

- **GUI 起動の目視確認は未実施**: この build host セッションは headless
  (X server / Wayland 無し、`xvfb-run` 等も未install) のため、実際に
  ウィンドウが開き Misskey ログイン画面が表示されるところまでは確認できて
  いない。確認したのは (1) 実行ファイルが即座に crash/セグフォルトしない
  こと、(2) 全共有ライブラリ依存がこのマシン上で解決すること、の2点まで。
  client 機 (ayaka 等、GUI 環境あり) での実地起動確認を推奨。

## AUR (miria-bin) との相違点 (依存方針)

| 項目 | miria-bin (AUR) | aria-bin (本 repo) |
|---|---|---|
| release 形式 | `.deb` (`dpkg` 相当で展開) | `.tar.gz` (self-contained bundle) |
| `.desktop` / icon 入手元 | `.deb` に同梱 | upstream `flatpak/` dir から個別 vendor |
| 動画/音声backend | system `mpv` に depends | bundle 同梱 `libmdk`/`libffmpeg` (mpv 不要) |
| 追加 depends | (`gtk3` `mpv` `libsecret`) | `gst-plugins-base-libs` (gstreamer audio backend、ldd 実測で判明) |

なぜ AUR に無いパッケージを追加するか: aria は miria の fork で機能追加が
活発、fork元の 4sterisk 氏は本 repo 保守者の友人。upstream の Linux
release pipeline (`github-actions` bot による GPG verified commit からの
自動 build) を信頼境界に入れる形。

## 結論

**approve** — build host で `makepkg -s --sign --key <Nekono key>` 可。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-07-14 | 1.5.8-1 | (this commit) | `dbce179b61f597b43aca4e8b0f63a5c8f079adf7` (GPG verified) | 初回追加、approve |
| 2026-07-17 | 1.5.8-2 | bot PR #406 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): gst-plugins-base-libs 1.28.5-1 → 1.28.5-2 |

## 更新方針

upstream の新 release (v1.5.9 等) が出たら:
1. GitHub Releases で新 tag の `linux-x64` / `linux-arm64` tarball の sha256 を
   独立に再計算 (`curl -fsSL <url> | sha256sum`)
2. `flatpak/com.poppingmoon.aria.{desktop,png,metainfo.xml}` が同 tag で変更
   されていないか確認、変更あれば sha256 を再計算
3. PKGBUILD の `pkgver` / `sha256sums*` を差し替え、`.SRCINFO` を再生成
   (`makepkg --printsrcinfo > .SRCINFO`)
4. tag の git commit が GPG verified か `gh api repos/poppingmoon/aria/commits/<sha>`
   で確認
5. `ldd` で新 release の依存ライブラリに変化がないか確認 (bundle 同梱の
   media backend が変わっていないか等)
6. 本 REVIEW.md の更新履歴に確認日 + 結論を追記
