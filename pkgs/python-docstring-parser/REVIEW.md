# python-docstring-parser review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR `python-docstring-parser` 0.18.0-1 を fork。 **AUR の depends typo bug を fix**
(= AUR は `depends=('python-pytest')` が後段で `depends=('python')` を上書きしてしまい、
runtime に pytest を要求していた、 upstream pyproject.toml では python のみが runtime
dep なのでこれは明らかな AUR 側のミス)。

## 用途

`python-cyclopts` (= 別 PR で並行投入) の **直接依存**。 cyclopts は
`python-fastmcp` 3.2.4 の `cyclopts>=4.0.0` の直接依存。

ayaka 上の自家 MCP server を pacman 経由で完全 install できるようにする取り組み、
fastmcp dep audit 漏れ fix の一環。

## Source

- Upstream: https://github.com/rr-/docstring_parser
  - tag `0.18.0` (= prefix 無し)、 license MIT
  - PyPI sdist (`docstring_parser-0.18.0.tar.gz`) を vendor
- AUR: https://aur.archlinux.org/packages/python-docstring-parser
  - AUR maintainer: Agil Mammadov <mammadovagil@tutamail.com>

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/d/docstring_parser/docstring_parser-0.18.0.tar.gz`
  - PyPI 公式 CDN
- [x] `sha256sums` 独立検証
  - AUR PKGBUILD の sha256 と一致
- [x] `build()`: `python -m build --wheel --no-isolation` のみ
- [x] `package()`: LICENSE 配置 + `installer` のみ
- [x] `depends`: python のみ (= AUR bug fix 後の正しい値、 upstream pyproject.toml 一致)
- [x] `makedepends`: python-build / python-installer / python-hatchling — Arch 公式
- [x] `checkdepends`: python-pytest (= AUR が runtime に誤配置していたものを test-only に正す)
- [x] `secrets` 混入なし
- [x] `arch=(any)` — pure Python

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | Agil Mammadov → Nekono、 Agil を Contributor に降格 | [nekono] fork 表示 |
| `depends` | AUR の **重複行 bug を fix** (= `depends=('python')` の後に `depends=('python-pytest')` で上書きされていた)、 [nekono] では `depends=('python')` のみに correct + pytest を `checkdepends` に移動 | upstream `pyproject.toml` の runtime deps は **python のみ**。 pytest は test 用 dep であって runtime ではない。 AUR 修正 PR を別途送るのが筋だが [nekono] では先に正しい状態を反映 |

## 結論

**approve** — build host で `bin/build-all python-docstring-parser` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-docstring-parser]` section) が検知 → Issue 経由で人間が手作業更新。 AUR が depends typo を修正したら REVIEW.md「AUR との意図的差分」 から該当行を削除可。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 0.18.0-1 | `28fbf82` | PyPI sdist `docstring_parser-0.18.0.tar.gz` | 初回 add、 AUR fork + depends typo bug fix、 cyclopts dep 経由で fastmcp chain 完成のため必要 (= 当初 SHA `f54af1e`、 master rebase で `28fbf82` に再付番) |
| 2026-06-15 | 0.18.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
