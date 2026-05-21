# python-rich-rst review

## 状態

**review 済み、 approve** (最新: 2026-05-21 / 2.0.1)

AUR `python-rich-rst` 1.3.2-1 を **純 fork** して投入後、upstream 2.0.1 に bump。

## 用途

`python-cyclopts` (= 別 PR で並行投入) の **直接依存**。 cyclopts は
`python-fastmcp` 3.2.4 の `cyclopts>=4.0.0` の直接依存。

## Source

- Upstream: https://github.com/wasi-master/rich-rst
  - tag `v1.3.2`、 license MIT
  - GitHub Release archive (= `releases/download/v1.3.2/rich_rst-1.3.2.tar.gz`)
- AUR: https://aur.archlinux.org/packages/python-rich-rst
  - AUR maintainers: Jesus Alvarez / Rafael Baboni Dominiquini

## 検証結果

- [x] `source` URL = `github.com/wasi-master/rich-rst/releases/download/v1.3.2/rich_rst-1.3.2.tar.gz`
  - GitHub 公式 release archive
- [x] `sha256sums` 独立検証
  - AUR PKGBUILD の sha256 と一致 (= 純 fork)
- [x] `build()`: `python -m build --wheel --no-isolation` のみ
- [x] `package()`: `installer` のみ
- [x] `depends`: python-pygments / python-rich — 両方 Arch 公式 extra (2.0.1 で python-docutils 削除、python-pygments 追加)
- [x] `makedepends`: python-build / python-installer / python-setuptools / python-wheel — Arch 公式 (2.0.1 で python-poetry-core 削除、build-backend が poetry-core → setuptools に変更)
- [x] `secrets` 混入なし
- [x] `arch=('any')` — pure Python

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | Jesus Alvarez (`Maintainer`) + Rafael Baboni Dominiquini (`Maintainer`) → Nekono、 両者を Contributor に降格 | [nekono] fork 表示 |

それ以外は **0 行差分** (= 純 fork)。

## 依存変更 (1.3.2 → 2.0.1)

upstream 2.0.1 で build-backend が `poetry-core` → `setuptools` に変更された。
これに伴い:
- `makedepends`: `python-poetry-core` を削除
- `depends`: `python-docutils` を削除、`python-pygments` を追加 (upstream pyproject.toml の requires より)

## 結論

**approve** — build host で `bin/build-all python-rich-rst` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-rich-rst]` section) が検知 → Issue 経由で人間が手作業更新。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-21 | 2.0.1-1 | (本 PR commit SHA) | upstream tag `v2.0.1` | upstream bump、build-backend poetry-core→setuptools、depends docutils→pygments、PR #97 close (pkgrel bump PR を pkgver bump で置換) |
| 2026-05-20 | 1.3.2-1 | `d72d1a0` | upstream tag `v1.3.2` | 初回 add、 AUR 純 fork、 cyclopts dep 経由で fastmcp chain 完成のため必要 (= 当初 SHA `005cfc9`、 master rebase で `d72d1a0` に再付番) |
