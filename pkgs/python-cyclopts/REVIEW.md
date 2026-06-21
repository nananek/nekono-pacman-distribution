# python-cyclopts review

## 状態

**review 済み、 approve** (最新: 2026-05-21 / 4.14.1)

AUR `python-cyclopts` 4.12.0-1 を fork。 **AUR の supply-chain 欠陥
(`sha256sums=('SKIP')`) を fix** して [nekono] に投入。

## 用途

`python-fastmcp` 3.2.4 の **直接依存** (= upstream pyproject.toml の
`cyclopts>=4.0.0`)。 fastmcp dep audit で漏れ発覚した 10 個のうちの 1 つ、
[nekono] 投入が必要だった分。

## Source

- Upstream: https://github.com/BrianPugh/cyclopts
  - tag `v4.12.0`、 license Apache-2.0
  - PyPI sdist (`cyclopts-4.12.0.tar.gz`) を vendor
- AUR: https://aur.archlinux.org/packages/python-cyclopts
  - AUR maintainer: Jesus Alvarez <jesusalv@rez.codes>

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/c/cyclopts/cyclopts-4.12.0.tar.gz`
  - PyPI 公式 CDN
- [x] `sha256sums` 独立検証
  - build host で curl + sha256sum で実測: `86bfb5b35cb078decc1cca6c1be41f9a0e6202dc43b4f6056d5cfc6d1f4a69d1`
  - **AUR は `SKIP` で hash 検証を skip** していた supply-chain 欠陥を [nekono]
    では実 hash で固定
- [x] `build()`: `python -m build --wheel --no-isolation` のみ
- [x] `package()`: `python -m installer --destdir` のみ
- [x] `depends`: python / python-attrs / python-docstring-parser / python-rich /
  python-rich-rst — Arch 公式 + [nekono] cross-dep 2 個 ([[python-docstring-parser]] / [[python-rich-rst]])
- [x] `makedepends`: python-build / python-installer / python-wheel / python-hatchling / python-hatch-vcs — Arch 公式
- [x] `secrets` 混入なし
- [x] `arch=('any')` — pure Python

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | Jesus Alvarez → Nekono、 Jesus を Contributor に降格 | [nekono] fork 表示 |
| `sha256sums` | `'SKIP'` → `'86bfb5b35cb078decc1cca6c1be41f9a0e6202dc43b4f6056d5cfc6d1f4a69d1'` | **supply-chain audit 必須**、 SKIP は [nekono] 規約違反 (= AUR maintainer に修正 PR 送るのが筋だが [nekono] では先に正しい状態を反映) |

それ以外は **0 行差分** (= 純 fork)。

## 結論

**approve** — build host で `bin/build-all python-cyclopts` で build + sign + repo db 追加可。 ただし **python-docstring-parser (= PR #70) + python-rich-rst (= PR #71) を先に publish してから** でないと depends が解決しない。

## 更新方針

upstream で新 release が出たら nvchecker (= `[python-cyclopts]` section) が検知 → Issue 経由で人間が手作業更新。 sha256 は **毎 bump 必ず独立計算** (= AUR が SKIP のままなら [nekono] でも入れ続ける必要)。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-21 | 4.14.1-1 | (本 commit 後に確定) | upstream tag `v4.14.1` | python-rich-rst 上限緩和 (<2.0.0→<3.0.0)。PKGBUILD 改変は pkgver + sha256 の 2 値のみ。safe-to-bump。 |
| 2026-06-03 | 4.16.1-1 | (this PR) | upstream tag `v4.16.1` | safe-to-bump。 機能追加 3 件 (App.synonym / __cyclopts_returncode__ / PEP 692) + バグ修正 2 件。 deps / build 無変化。 sha256 独立検証済み |
| 2026-06-10 | 4.17.0-1 | (this PR) | upstream tag `v4.17.0` | safe-to-bump。 機能追加 (async コマンド対応 / Parameter.show_default の文字列対応) + PyPI publish attestation 導入 (供給チェーン強化)。 deps / build 無変化。 sha256 独立検証済み |
| 2026-06-12 | 4.18.0-1 | (this PR) | upstream tag `v4.18.0` | safe-to-bump (Issue #199)。 slice 型 native サポート + Slice validator / NonEmptySlice 追加、 ネスト Annotated/Optional/NewType の Parameter メタデータ脱落修正。 deps / build 無変化。 sha256 独立検証済み (PyPI API + 実測一致) |
| 2026-06-15 | 4.18.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
| 2026-06-21 | 4.18.0-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
| 2026-05-20 | 4.12.0-1 | `0cced1785b136180286916e6e0225cc709cb9de7` | upstream tag `v4.12.0` | 初回 add、 AUR fork + sha256 SKIP fix、 fastmcp 直接依存 |
