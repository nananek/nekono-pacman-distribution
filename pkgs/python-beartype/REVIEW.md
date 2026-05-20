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

それ以外は **0 行差分** (= 純 fork)。 checkdepends に python-rich-click 等
AUR-only / 不要 pkg があるが makepkg default では check() skip なので無害、
本 PR では弄らない。

## 結論

**approve** — build host で `bin/build-all python-beartype` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-beartype]` section) が検知 → Issue 経由で人間が手作業更新。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 0.22.9-1 | (初回 PR commit SHA、 push 後に固定) | upstream tag `v0.22.9` | 初回 add、 AUR 純 fork、 py-key-value-aio dep 経由で fastmcp chain 完成のため必要 |
