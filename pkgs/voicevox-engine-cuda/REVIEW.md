# voicevox-engine-cuda review

## 状態

**review 済み、 approve (条件付き)** (2026-05-19)

AUR の `voicevox-engine` PKGBUILD (pkgver=0.24.1, pkgrel=1) を fork、 CUDA
専用 variant 化 + supply-chain 強化のためフル改修。 意図的改変多数 (後述)。

## Source

- AUR: https://aur.archlinux.org/packages/voicevox-engine
  - Maintainer: t1ckbase
- Upstream:
  - https://github.com/VOICEVOX/voicevox_engine
  - VOICEVOX Project (= Hiroshiba 個人運営の非営利プロジェクト)
  - LGPL-3.0-only + 同梱モデル / 音声に独自ライセンス (キャラクター利用規約)
  - 上流 release tag `0.24.1` の commit SHA: `fb9972dd5753f81c2a3a1af52b09bfc01f1608ab`
- 上流公式の prebuilt bundle: GitHub Releases から download
  (= `voicevox_engine-linux-nvidia-${pkgver}.7z.00{1,2}`)
- vendored PyPI distfiles (= source[] に sha256 pin 済み):
  - kanalizer 0.1.1 (PyPI、 manylinux_2_28_x86_64 wheel)
  - pyopenjtalk 0.4.1 (PyPI、 sdist)
  - pyworld 0.3.5 (PyPI、 sdist)
- open_jtalk dict 1.11 (r9y9/open_jtalk リリース、 UTF-8 版)

## 検証結果

### 公式 release tarball / dict

- [x] `0.24.1.tar.gz` (= engine ソース)
  - URL: `github.com/VOICEVOX/voicevox_engine/archive/refs/tags/0.24.1.tar.gz`
  - sha256: `0488d3b98b4e2876c3070c2287347df92cacf724f0e986c60cad7aabffb6cf7e` (AUR 値と一致)
  - tag commit `fb9972dd5753f81c2a3a1af52b09bfc01f1608ab` から生成された archive
- [x] `open_jtalk_dic_utf_8-1.11.tar.gz`
  - URL: `github.com/r9y9/open_jtalk/releases/download/v1.11.1/...`
  - sha256: `fe6ba0e43542cef98339abdffd903e062008ea170b04e7e2a35da805902f382a` (AUR 値と一致)
  - typosquat なし、 r9y9 は pyopenjtalk maintainer 本人

### NVIDIA bundle (prebuilt binary)

- [x] `voicevox_engine-linux-nvidia-0.24.1.7z.001` (1,992,294,400 byte)
  - URL: `github.com/VOICEVOX/voicevox_engine/releases/download/0.24.1/...`
  - sha256: `e5ff4bc3af152eda0e086820eeaf985f2cb4cb504fe66b839cff35e38b3628a7`
    (build host で実 DL → `sha256sum` で計算、 2026-05-19)
- [x] `voicevox_engine-linux-nvidia-0.24.1.7z.002` (335,022,269 byte)
  - sha256: `07cd32e093a30c210b931e04cb7a18b1585b0b720a9014c322c4d2783496a7df`
- [x] 上流公開 manifest `voicevox_engine-linux-nvidia-0.24.1.7z.txt` は part ファイル名
  リストのみで sha256 を含まない (= MD5 や sha256 を上流が公開していないため、
  hash の信頼の起点は build host で 1 度 DL + 計算したこの REVIEW.md commit となる)。
- [x] bundle 内部構造 (= `7z l` で確認):
  - top-level dir: `linux-nvidia/` (PKGBUILD の `_bundle_dir` で参照)
  - CUDA 11 stack 同梱: `libcudart.so.11.0`, `libcublas.so.11`,
    `libcublasLt.so.11`, `libcufft.so.10`, `libcurand.so.10`,
    `libcudnn.so.8` + `libcudnn_adv_infer.so.8` / `_cnn_infer.so.8` /
    `_ops_infer.so.8`
  - onnxruntime: `libonnxruntime.so` + `libonnxruntime_providers_cuda.so` +
    `libonnxruntime_providers_tensorrt.so` + `libonnxruntime_providers_shared.so`
  - VOICEVOX core: `libvoicevox_core.so`
  - **Python runtime も bundle に同梱**: `libpython3.11.so.1.0`
  - 他に basic runtime: `libgcc_s.so.1`, `libstdc++.so.6`, `libmvec.so.1`,
    `libgfortran-*-*.so.5.0.0`, `libquadmath-*-*.so.0.0.0`,
    `libscipy_openblas64_-*.so`

### PyPI vendored distfiles

- [x] kanalizer 0.1.1 wheel
  - sha256: `bb86f3fa2cfe4034fa68f3035e3235bb6ee314e1957aac9c0426d651e473709a` (PyPI digests と一致)
- [x] pyopenjtalk 0.4.1 sdist
  - sha256: `d5ada46f7fc2b52c1c79c273eb9668ff6ad7ab276a8db9d8be119ef93440f0dc`
- [x] pyworld 0.3.5 sdist
  - sha256: `1b93e53cddb67a0e4faa34d6cf919ac6c662feb1c8c0ed901d71b595ab396aa3`

### 静的解析

- [x] `prepare()`: 無し
- [x] `build()`: `7z x -y ... .7z.001` で bundle 抽出 → `python -m build --wheel
  --no-isolation` で engine wheel 生成。 ネットワーク取得なし
- [x] `package()`:
  - `python -m installer` で wheel 展開、 ファイル move のみ
  - `python -m pip install --no-index --no-deps --target ...` で **オフライン**
    vendor 展開 (sha 検証済み tarball / wheel が source[] に揃っているため
    pip は外部 fetch 不能)
  - shell injection / exec / curl / wget なし
- [x] `voicevox-engine.install`: post_install で標準出力にガイド表示のみ。
  systemd-sysusers / systemd-tmpfiles は libalpm の標準 hook が自動実行
  (= PKGBUILD と install file 内で重複 call しない)
- [x] `voicevox-engine.service`: 標準的な systemd Type=simple、 hardening
  あり (NoNewPrivileges, ProtectSystem=strict, ProtectHome, PrivateTmp,
  RestrictAddressFamilies, LockPersonality)。 GPU 用に PrivateDevices=false
  + DeviceAllow=/dev/nvidia* を最小許可。 W+X mapping (CUDA JIT) のため
  MemoryDenyWriteExecute=false

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| pkgname: `voicevox-engine` → `voicevox-engine-cuda` | CUDA 専用 variant であることを名前で表明。 `provides=voicevox-engine` / `conflicts=voicevox-engine` で互換性は維持 (editor の optdepends を満たす) |
| source の prebuilt bundle: `voicevox_engine-linux-cpu-x64-${pkgver}.7z.001` → `voicevox_engine-linux-nvidia-${pkgver}.7z.00{1,2}` (multi-part) | nekono-pacman0 は RTX 3050 6GB を抱えているので CUDA 推論を活用 |
| `_bundle_dir` 変数を導入し PKGBUILD で 1 箇所参照に統一 | 7z 内部 dir prefix (`linux-nvidia`) の hardcode を変数化、 将来の variant 切替時の事故防止 |
| 解凍 tool: `bsdtar -xf` → `7z x -y` | NVIDIA bundle は multi-volume 7z で libarchive (bsdtar) では結合解凍不可 |
| build() の `pip install kanalizer pyopenjtalk` (= PyPI 動的取得、 sha 検証なし) を削除 | **supply-chain 上の重大欠陥**。 build 時に PyPI 上書きが起きると検知できない。 source[] に `kanalizer-*.whl`, `pyopenjtalk-*.tar.gz`, `pyworld-*.tar.gz` を vendor (sha256 pin) + `pip install --no-index --no-deps --target` でオフライン install に置換 |
| `_syspymods` から `pyworld` を除外 | Arch 公式 repo に `python-pyworld` 無し (AUR のみ)。 [nekono] は AUR helper 前提を取らない方針なので、 pyworld も source[] に vendor 投入 |
| depends から `cuda` と `cudnn` を **削除** | bundle に CUDA 11 + cuDNN 8 stack が完全同梱されており、 Arch の cuda (= 13.x) / cudnn (= 9.x) は ABI mismatch。 むしろ install すると LD path 競合の恐れ。 `nvidia-utils` (= libcuda.so.1) のみで足りる |
| systemd unit / sysusers / tmpfiles を同梱 (新規) | engine を daemon 化するための周辺ファイル。 ansible-nekonodesk が `/etc/systemd/system/voicevox-engine.service.d/override.conf` で bind 先 IP (= tailscale0) を上書きする 2 段構成 |
| systemd unit に `Environment=LD_LIBRARY_PATH=/usr/lib/VOICEVOX/vv-engine` | bundle 同梱の CUDA 11 stack を ld.so に見せる (system に CUDA 12/13 が入っていても優先される) |
| makedepends に `7zip` 追加 | multi-volume 7z 解凍に必要 |

## 既知の divergence / リスク

1. **upstream pyproject.toml の `requires-python = "==3.11.9"` を強制していない**。
   AUR が system Python (= Arch の python 3.13/3.14) で問題なく動作している実績に
   依拠 (= 主要処理は bundle 内の libonnxruntime/libvoicevox_core を C-only で叩く
   ため、 system Python の minor version 差で破綻しにくい構造)。 もし
   `--use_gpu` 起動時に Python ABI 関連の crash が起きたら、 entrypoint を bundle
   同梱の `libpython3.11.so.1.0` 経由で起動する形に切替検討。

2. **upstream pyproject.toml は pyopenjtalk を VOICEVOX/pyopenjtalk @
   74703b034dd90a1f199f49bb70bf3b66b1728a86 (git fork pin) で要求している** が、
   当 PKGBUILD は AUR と同じく PyPI 配布版 (r9y9/pyopenjtalk 0.4.1) を使う。
   AUR の動作実績を踏襲。 issue 発生時は VOICEVOX fork commit を vendoring に
   切替えること。

3. **prebuilt binary を信頼している**。 `libonnxruntime*.so` / `libvoicevox_core.so` /
   `libcuda*.so` / `libpython3.11.so` 等は upstream の GitHub Releases 経由でし
   か入手できず、 source build していない。 これは [nekono] のもう 1 つの
   prebuilt 系 pkg (vesktop-bin / docker-rootless-extras 等) と同じ trust model
   (= upstream 配布 binary を信頼)。 hash pin により download 経路の改竄は検知可能。

4. **hash の信頼の起点**: upstream は MD5 のみ部分公開 (= AppImage 系では `.7z.txt`
   に MD5、 engine 用 `.7z.txt` は part 名 list のみで MD5 すら無し)。 当 commit
   が記録する sha256 は build host (nekono-pacman0) で 2026-05-19 に 1 度 DL +
   `sha256sum` で計算した値。 次の bump 時には新 release を同じ手順で DL → sha
   再計算 → REVIEW.md に追記、 trust chain を継続する。

## 結論

**approve (条件付き)** —

- build host (nekono-pacman0) で `makepkg -s --sign --key 483D...` で build 可。
  `--use_gpu` 起動の smoke test を build 後に 1 度実施し、 LD_LIBRARY_PATH 経路で
  bundle 同梱 CUDA 11 stack が resolve されることを確認すること。
- もし Arch の `cuda` / `cudnn` が host に install されていると、 LD 解決順次第で
  bundle と system のどちらが先に拾われるか不定 → 念のため host から system
  cuda は外す or LD_LIBRARY_PATH を service unit で明示済みである現状を維持。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.24.1-3 (runtime fix + pkgrel bump) | (本 PR の commit SHA を merge 時に追記) | — | system Python の C extension (soxr_ext / 他) が `ImportError: CXXABI_1.3.15 not found in /usr/lib/VOICEVOX/vv-engine/libstdc++.so.6` で fail していた問題を修正。 bundle 内の `libstdc++.so.6` / `libgcc_s.so.1` / `libmvec.so.1` が LD_LIBRARY_PATH=/usr/lib/VOICEVOX/vv-engine 経由で system 版より先に resolve され、 古い ABI が system の新 C ext に互換しない問題。 package() で該当 3 lib を `rm -f` → system 版 (gcc-libs / glibc) にフォールバックさせる。 bundle 同梱の CUDA stack (libcudart 等) は system libstdc++ で問題なく動く (= libstdc++ は ABI backward compatible)。 pkgrel を 2 → 3 に bump |
| 2026-05-19 | 0.24.1-2 (pkgrel bump) | `736073d` (PR #52 merge) | — | pkgrel を 1 → 2 に bump。 PR #43 / #46 / #48 / #51 の build fix を pkgrel 据え置きで重ねていたが、 build host (nekono-pacman0) で 0.24.1-1 が既に build + install (= restart loop 観測) されてしまっていたため、 client の pacman -Syu で新 build を拾わせるには pkgrel +1 が必要。 fix 内容自体は #51 を踏襲 |
| 2026-05-19 | 0.24.1 (build fix) | `57f4d24` (PR #51 merge) | — | runtime ImportError 修正: `import uvicorn` で fail し systemd restart loop していた問題を解消。 Arch には extra/uvicorn (= `python-` prefix なしの命名例外) で存在するので depends に直接追加。 starlette / setuptools も `_syspymods` に追加 (= python-starlette, python-setuptools)。 当初 vendor で対応しかけたが Arch 公式 pkg 経由の方が clean (= pacman 自動更新、 サイズ減、 trust chain 短) なので方針転換 |
| 2026-05-19 | 0.24.1 (build fix) | `<前回の merge commit SHA>` (PR #48 merge) | — | direct_url.json (PEP 610) の $srcdir 絶対 path leak を pip install 後の `rm -f` で除去 + `options=(!strip !debug)` で prebuilt .so の debug pkg / gdb-add-index 警告を抑止 (= PR #48) |
| 2026-05-19 | 0.24.1 (build fix) | `b4108ea` (PR #46 merge) | — | bundle に同梱されている Python 3.11 build の kanalizer / pyopenjtalk / pyworld + dist-info を pip install --target 前に `rm -rf` で除去 (= system Python 3.14 と ABI 不一致、 かつ pip --target が既存 dir を warn+skip する挙動への対応)。 retroactive 記録 |
| 2026-05-19 | 0.24.1 (build fix) | (本 PR の commit SHA を merge 時に追記) | — | `--no-build-isolation` 追加: pyopenjtalk sdist の PEP 517 build isolation が `--no-index` と衝突し setuptools>=64 を fetch 不能となる問題を修正。 build host (nekono-pacman0) の実 build で発覚 |
| 2026-05-19 | 0.24.1 | (本 PR の commit SHA を merge 時に追記) | `fb9972dd5753f81c2a3a1af52b09bfc01f1608ab` | 初回 add、 supply-chain 強化フル改修 |
