# python-py-key-value-aio review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR `python-py-key-value-aio` 0.4.0-1 を fork し、 **pkgver 0.4.4 にバンプ** + **extras 反映**。

## 用途

`python-fastmcp` 3.2.4 の **直接依存** (= upstream pyproject.toml の
`py-key-value-aio[filetree,keyring,memory]>=0.4.4`)。 fastmcp の `server.py`
が `key_value.aio.stores.memory` をトップレベル import するため必須。

## Source

- Upstream: https://github.com/chrisguidry/py-key-value-aio
  - tag `v0.4.4`、 license Apache-2.0
  - PyPI sdist (`py_key_value_aio-0.4.4.tar.gz`) を vendor
- AUR: https://aur.archlinux.org/packages/python-py-key-value-aio
  - AUR maintainer: LY <ly-niko@qq.com>

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/p/py-key-value-aio/py_key_value_aio-0.4.4.tar.gz`
- [x] `sha256sums` 独立検証: `e3012e6243ed7cc09bb05457bd4d03b1ba5c2b1ca8700096b3927db79ffbbe55`
- [x] `prepare()`: sed で `uv_build>=0.8.2,<0.9.0` → `<2.0.0` (= Arch 0.11.x との衝突回避)
- [x] `build()`: `python -m build --wheel --no-isolation`
- [x] `package()`: `installer` + README 配置
- [x] `depends`: python>=3.10 / python-beartype ([nekono]) / python-typing_extensions / python-cachetools (= memory extra)
- [x] `optdepends`: python-keyring / python-aiofile (AUR-only) / python-redis (= lazy backends)
- [x] `secrets` 混入なし
- [x] `arch=('any')`

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | LY → Nekono、 LY を Contributor に降格 | [nekono] fork 表示 |
| `pkgver` | 0.4.0 → 0.4.4 | fastmcp 3.2.4 が `>=0.4.4` を要求 |
| `sha256sums` | 0.4.4 で再計算 | 上記 bump に伴い |
| `depends`: python-cachetools 追加 | (= memory extra) | fastmcp server.py が `key_value.aio.stores.memory` をトップレベル import するため hard dep |
| `optdepends`: python-keyring / python-aiofile / python-redis 追加 | (= keyring / filetree / redis extras) | fastmcp の oauth_proxy 等で lazy import、 当 repo の MCP server では使わない |
| `prepare()` の sed | `<1.0.0` → `<2.0.0` に拡大 | Arch python-uv-build 0.11.x、 念のため広げる |

## 結論

**approve** — build host で `bin/build-all python-py-key-value-aio` で build + sign + repo db 追加可。 ただし **python-beartype (= PR #69) を先に publish してから** でないと depends が解決しない。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-py-key-value-aio]` section) が検知 → Issue 経由で人間が手作業更新。 extras の dep 配分は fastmcp 側の import pattern を再確認して維持。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 0.4.4-1 | `e808468` | upstream tag `v0.4.4` | 初回 add、 AUR fork + 0.4.4 bump + extras 反映 (= 当初 SHA `eb0fae2`、 master rebase で `e808468` に再付番) |
| 2026-05-23 | 0.4.4-2 | `f6d9184` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-cachetools 7.1.3-1 → 7.1.4-1, python-uv-build 0.11.15-1 → 0.11.16-1 |
| 2026-05-30 | 0.4.4-3 | bot PR #137 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.16-1 → 0.11.17-1 |
| 2026-06-02 | 0.4.4-4 | bot PR #143 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.17-1 → 0.11.18-1 |
| 2026-06-05 | 0.4.4-5 | bot PR #169 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.18-1 → 0.11.19-1 |
| 2026-06-12 | 0.4.4-6 | bot PR #201 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.19-1 → 0.11.21-1 |
| 2026-06-15 | 0.4.4-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-beartype 0.22.9-5 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-21 | 0.4.4-8 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.21-1 → 0.11.22-1 |
| 2026-06-23 | 0.4.4-9 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1, python-uv-build 0.11.22-1 → 0.11.23-1 |
| 2026-06-24 | 0.4.4-10 | bot PR #276 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.23-1 → 0.11.24-1 |
| 2026-06-28 | 0.4.4-11 | bot PR #303 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.24-1 → 0.11.25-1 |
| 2026-07-02 | 0.4.4-12 | bot PR #324 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.25-1 → 0.11.26-1 |
| 2026-07-06 | 0.4.4-13 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-beartype 0.22.9-7 → 0.22.9-8 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-07 | 0.4.4-14 | bot PR #360 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-typing_extensions 4.15.0-3 → 4.16.0-1 |
| 2026-07-09 | 0.4.4-15 | bot PR #367 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.26-1 → 0.11.27-1 |
| 2026-07-10 | 0.4.4-16 | bot PR #374 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.27-1 → 0.11.28-1 |
| 2026-07-17 | 0.4.4-17 | bot PR #408 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.28-1 → 0.11.29-1 |
| 2026-07-22 | 0.4.4-18 | bot PR #426 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-uv-build 0.11.29-1 → 0.11.30-1 |
