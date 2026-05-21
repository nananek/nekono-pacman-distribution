# python-beartype review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR `python-beartype` 0.22.9-1 を **純 fork** (= 差分は Maintainer 行追加のみ)。

## 用途

`python-py-key-value-aio` (= 別 PR で並行投入) の **直接依存**。 py-key-value-aio
は `python-fastmcp` 3.2.4 の `py-key-value-aio[filetree,keyring,memory]>=0.4.4`
の transitive 依存。

ayaka 上の自家 MCP server を pacman 経由で完全 install できるようにする
取り組み、 fastmcp dep audit 漏れ fix の一環。

## Source

- Upstream: https://github.com/beartype/beartype
  - tag `v0.22.9`、 license MIT
  - GitHub archive (= `archive/v0.22.9.tar.gz`)
- AUR: https://aur.archlinux.org/packages/python-beartype
  - AUR maintainer 不在 (= orphan)、 Contributors: redponike / Carl Smedstad / Achmad Fathoni

## 検証結果

- [x] `source` URL = `github.com/beartype/beartype/archive/v0.22.9.tar.gz`
  - GitHub 公式 release archive
- [x] `sha256sums` 独立検証
  - AUR PKGBUILD の sha256 と一致 (= 純 fork)
- [x] `build()`: `python -m build --wheel --no-isolation`
- [x] `package()`: `installer` + LICENSE 配置
- [x] `check()`: pytest tests 走らせる構造を AUR から踏襲 (= makepkg default では skip
  されるので build pipeline には影響なし)
- [x] `depends`: python>=3.9 のみ (= 真の leaf)
- [x] `makedepends`: python-build / python-hatchling / python-installer / python-wheel — Arch 公式
- [x] `secrets` 混入なし
- [x] `arch=('any')` — pure Python

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | 追加 (= AUR は orphan で Contributor のみ)、 Contributors はそのまま | [nekono] fork 表示 |
| `checkdepends` に `python-poetry` 追加 + `pkgrel` 1 → 2 | (= AUR には無い差分) | AUR の checkdepends だけだと `beartype_test/a90_func/a50_external/test_poetry.py::test_poetry` が `subprocess.CalledProcessError: Command '('/usr/bin/python', '-m', 'poetry', 'install', ...)'` で fail する (= 2026-05-20 PR #69 直後の再 build で実測)。 AUR maintainer の手元では poetry が偶々入っていて test 通っていたものと推測。 [nekono] では check() を意味的に走らせるため明示。 全 checkdepends は Arch 公式で揃う |

それ以外は AUR と 0 行差分。

## 結論

**approve** — build host で `bin/build-all python-beartype` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-beartype]` section) が検知 → Issue 経由で人間が手作業更新。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 0.22.9-1 | `795f1f0` | upstream tag `v0.22.9` | 初回 add、 AUR 純 fork、 py-key-value-aio dep 経由で fastmcp chain 完成のため必要 (= 当初 SHA `659b27b`、 master rebase で `795f1f0` に再付番) |
| 2026-05-20 | 0.22.9-2 | `328e5a532467a7df920b773635091ae68235c394` | (同上) | `pkgrel` +1: checkdepends に `python-poetry` 追加 (= AUR PKGBUILD 漏れの test_poetry 通過用、 build host で実測 fail を確認、 全 dep Arch 公式で揃う) |
| 2026-05-21 | 0.22.9-3 | `d83ada9` | (同上) | `pkgrel` +1 (deps changed): python-click 8.3.3-1, python-poetry 2.4.1-1, python-pygments 2.20.0-1, python-pytest 1:9.0.3-1, python-rich-click 1.9.7-1, python-xarray 2026.04.0-1 |
