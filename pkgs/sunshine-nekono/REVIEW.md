# sunshine-nekono review

## 状態

**review 済み、approve** (2026-09-15 / fork tag v2026.906.222525-nekono.1)

`moonlight-qt-nekono` と同じく **自分の fork を自分でビルドする** package。信頼の起点が
Nekono 自身の commit になる点は同じなので、そちらの REVIEW.md の「信頼モデルの差」も
併せて読むこと。

配信先クライアントは `moonlight-qt-nekono`。upstream の `sunshine-bin` は 1 ディスプレイ
のみなので、複数ディスプレイ配信にはこの package とペアで使う。

## Source

- Fork: https://github.com/nananek/Sunshine (branch `multidisplay`)
  - Upstream: https://github.com/LizardByte/Sunshine v2026.906.222525 ベース
  - GPL-3.0-only。ホストの複数ディスプレイを同時にキャプチャ・配信できるようにした改造版
- PKGBUILD の下敷き: AUR `sunshine` (ソースビルド版)
  - **大幅に改変**。AUR 版は `sha256sums=('SKIP')` + `prepare()` 内 `git submodule update`
    + CPM/npm の build 時取得という、本 repo の規約に真っ向から反する作りだったため、
    ネットワーク取得を全て排除する形に作り直した

## Tarball の素性

| 対象 | commit |
|---|---|
| Sunshine (superproject) | `c5afdde2` |
| moonlight-common-c | `6cb7e7aa1cb29dd70629685056cf8b57399292fa` (Nekono fork) |

submodule 16 個を含めた作業ツリーから `.git` を除去して生成。mtime は pin した commit の
author date (`git log -1 --format=%ct`) に固定しており、同じ commit からバイト単位で
再生成できる (CLAUDE.md「よく踏む落とし穴」#7 参照)。

### build-deps の間引き (本 repo で前例の無い操作)

`third-party/build-deps` は素の状態で **1.4G** あり、内訳は `third-party/FFmpeg/` 配下の
AMF SDK (1.2G) / FFmpeg source (114M) / SVT-AV1 (38M) / Vulkan-Headers (41M)。

本 package は **prebuilt FFmpeg を使う**ため、FFmpeg を自前ビルドするための素材は不要。
`Vulkan-Headers` のみ残して他を削除し、tarball を 1.6G → 176M (圧縮後 96M) に縮小した。

残す必要があるもの (削ると configure が落ちる):
- `build-deps/package-lock.cmake` — `CMakeLists.txt:74` が include
- `build-deps/third-party/FFmpeg/Vulkan-Headers/include` — `cmake/compile_definitions/linux.cmake:131`
  が参照 (system の Vulkan headers では古い場合があるため upstream が同梱している)

間引きの妥当性は **ビルドが通ることで検証**した。

## 検証結果

- [x] `source` は 6 本、**全て sha256 実値で pin。SKIP 不使用**

  | source | 提供元 | sha256 |
  |---|---|---|
  | fork source tarball | nananek/Sunshine release | `5f471841...` |
  | npm cache | nananek/Sunshine release | `dde939d6...` |
  | boost 1.89.0 | boostorg/boost 公式 release | `67acec02...` |
  | FFmpeg prebuilt | LizardByte/build-deps 公式 release | `496d2bbb...` |
  | nlohmann/json 3.11.3 | nlohmann/json 公式 release | `d6c65aca...` |
  | sunshine.install | 本 repo の `sunshine-bin` から流用 | `d7410467...` |

  - boost の hash は Sunshine 自身の `cmake/dependencies/Boost_Sunshine.cmake:61` が
    `URL_HASH SHA256=` で宣言している値と一致することを確認
  - FFmpeg / json は独立に download して `sha256sum` で再計算

- [x] **`build()` / `prepare()` でネットワーク取得なし**
  - boost: `FETCHCONTENT_SOURCE_DIR_BOOST` で展開済み dir を指定し FetchContent の download を抑止
  - FFmpeg: `FFMPEG_PREPARED_BINARIES` を定義すると `cmake/dependencies/ffmpeg.cmake` は
    冒頭の `if(NOT DEFINED FFMPEG_PREPARED_BINARIES)` で丸ごと skip される。
    これにより **download だけでなく `git describe --tags` / `git fetch --tags` も回避**
    される (`.git` を持たない tarball で問題になる箇所)
  - json: `FETCHCONTENT_SOURCE_DIR_JSON` で同様に差し替え
  - nv_codec_headers: prebuilt FFmpeg に同梱されているため CPM 経路に入らない
  - npm: vendor した cache から `npm ci --offline`。**registry に到達不能な
    `http://127.0.0.1:9` を明示**しており、将来 `--offline` が効かなくなった場合に
    黙ってネットワークへ出るのではなく失敗して気づけるようにしてある
  - curl / wget / pipe-to-shell / eval / `pip install` いずれも無し

- [x] `package()` は `DESTDIR="$pkgdir" cmake --install build` のみ
- [x] version は `git describe` が使えないため `BRANCH` / `BUILD_VERSION` / `COMMIT` を
  環境変数で注入 (AUR 版と同じ方式)。ビルド済みバイナリに `2026.906.222525` が入ることを確認
- [x] `SUNSHINE_ENABLE_CUDA=OFF` — `sunshine-bin` と同じ構成
- [x] `install=sunshine.install` は `sunshine-bin` と同一ファイル (sha256 一致)。
  `setcap cap_sys_admin,cap_sys_nice+p` を行う。KMS/DRM キャプチャに必要な特権で、
  `sunshine-bin` の review で既に承認済みの内容

## ビルド検証

- `makepkg -sf` 成功 (2026-09-15)
- 成果物の確認:
  - `usr/bin/sunshine` に multidisplay のコードが入っている (strings で確認)
  - `usr/share/sunshine/web/` に `apps.html` / `config.html` / `assets/` 32 ファイルが生成
    されている = **npm オフラインビルドが実際に機能している**
  - バイナリに `2026.906.222525` が埋まっている

## 既知の制限

- CUDA OFF のため NVENC は system ffmpeg 経由の経路に依存する。fork 側で
  「CUDA 非搭載ビルドで DMA-BUF を要求してしまい全フレームが変換失敗する」バグを
  修正済み (`9b2f957d` 相当)
- upstream の security fix は自動で入らない。`multidisplay` branch を定期的に
  upstream に追従させる運用が必要
- ペン / タッチのディスプレイ指定、追加ウィンドウでのキーボード・ゲームパッド入力は未実装

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-09-17 | 2026.906.222525.nekono1-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): nodejs 26.8.2-1 → 26.9.0-1 |
| 2026-09-15 | 2026.906.222525.nekono1-1 | (this PR) | `c5afdde2` | 初回 review、approve (詳細は本ファイル上記 section) |
