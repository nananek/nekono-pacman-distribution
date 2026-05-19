# python-sse-starlette review

## 状態

**review 済み、 approve** (2026-05-19、 初回 add)

AUR の `python-sse-starlette` PKGBUILD (pkgver=3.3.4, pkgrel=1) を fork。 純 fork
(= 機能変化なし、 改変は Maintainer 行 + nekono 説明コメント + check() 削除のみ)。

## 用途

`python-fastmcp` 依存 chain (= `python-fastmcp` → `python-mcp` → `python-sse-starlette`)
で必要な leaf 依存。 Arch 公式 repo に不在で AUR にしか無いため [nekono] に
fork 投入する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-sse-starlette
  - Maintainer: Dawei Yang <yangdawei.home+archlinux.org@gmail.com>
  - Contributors: envolution / Carl Smedstad <carsme@archlinux.org> / AngrySoft - Sebastian Zwierzchowski
- Upstream: https://github.com/sysid/sse-starlette
  - sysid、 BSD-3-Clause
  - tag `v3.3.4` (= annotated tag object `5682373aa2f29c1e72863031e0d5dacabeff3b59`)、 dereferenced commit `c938db3f6ea262f5f087c75d8631c3aab9cbf0ad`

## 検証結果

- [x] `source` URL = `github.com/sysid/sse-starlette/archive/v3.3.4.tar.gz`
  - upstream 公式 GitHub Release archive、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `0c8200c30dcda56f2bc7e3d116815257073a6ad6674bc1745a11ca1004211345`
  - AUR PKGBUILD と一致
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + `install -Dm644 LICENSE` のみ、 shell injection / curl / wget なし
- [x] `depends`: python / uvicorn / python-anyio / python-starlette — 全て Arch 公式 extra
  - `uvicorn` は Arch 命名例外 (= `python-` prefix なし、 extra/uvicorn)
- [x] `makedepends`: python-build / python-installer / python-setuptools / python-wheel — 全て Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer / Contributors は `# Upstream AUR ...` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つため。 上流情報は信頼の起点として残す |
| `check()` 関数 + 8 個の `checkdepends` を削除 | checkdepends に `python-portend` 等の AUR 限定 pkg や `python-asgi-lifespan` 等の少量 install pkg を引きたくない、 [nekono] は AUR helper 前提を取らない方針。 build host は `bin/build-all` が `--check` なし default 運用 |
| `uvicorn` を hard depends に置く (= upstream pyproject.toml では optional-dependencies 分類) | AUR PKGBUILD 踏襲。 sse-starlette は実質 uvicorn 前提 (= ASGI server なし では起動しない)、 optional とは名ばかり。 ayaka 上の MCP 利用では fastmcp 経由で必ず uvicorn が呼ばれる |

## 結論

**approve** — build host で `bin/build-all python-sse-starlette` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release (= 3.3.5 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. 本 dir の PKGBUILD + .SRCINFO を差し替え
3. sha256 を独立再計算
4. depends / makedepends が変化していれば `.deps.lock` を更新
5. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 3.3.4 | `31116c43b87e9f3524fa9295e206397cde846878` | `c938db3f6ea262f5f087c75d8631c3aab9cbf0ad` | 初回 add、 純 fork (= check() 削除 + Maintainer 書換) |
