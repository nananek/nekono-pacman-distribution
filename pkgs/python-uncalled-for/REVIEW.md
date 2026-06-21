# python-uncalled-for review

## 状態

**review 済み、 approve** (最新: 2026-05-21 / 0.3.2)

AUR の `python-uncalled-for` PKGBUILD を fork。 純 fork
(= 機能変化なし、 改変は Maintainer 行 + nekono 説明コメント + prepare()/check()/
checkdepends 削除のみ)。

## 用途

`python-fastmcp` の transitive 依存。 fastmcp 3.2.4 がこの pkg の 0.2.0 を要求
する (= upstream 0.3.x まで進んでいるが当面 0.2.0 で固定)。 Arch 公式 repo に
不在で AUR にしか無いため [nekono] に fork 投入する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-uncalled-for
  - Maintainer: Mohamed Amine Zghal (medaminezghal) <medaminezghal at outlook dot com>
- Upstream: https://github.com/chrisguidry/uncalled-for/
  - chrisguidry (Chris Guidry)、 MIT
  - tag `0.2.0` (= prefix なし、 lightweight tag)、 commit `b39a1115e11db09533461df5341e1c78a583a80d`
- distfile: **PyPI sdist** (= `files.pythonhosted.org/packages/source/u/uncalled-for/uncalled_for-0.2.0.tar.gz`)
  - upstream maintainer が build した正規 distribution

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/u/uncalled-for/uncalled_for-0.2.0.tar.gz`
  - PyPI 公式 CDN、 PEP 503 path 構造 (= `<l>/<name>/<filename>`)、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `b4f8fdbcec328c5a113807d653e041c5094473dd4afa7c34599ace69ccb7e69f`
  - AUR PKGBUILD と一致
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` のみ、 shell injection / curl / wget なし
- [x] `depends`: python のみ — pyproject.toml も pure Python の最小依存
- [x] `makedepends`: python-hatchling / python-hatch-vcs / python-build / python-installer / python-wheel — 全て Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer は `# Upstream AUR Maintainer:` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つ。 上流情報は信頼の起点として残す |
| `prepare()` 関数を削除 | prepare() は check() で pytest-cov の `--cov` flag を pyproject.toml から sed する役目のみ。 check() を削除する以上 prepare() も無用 |
| `check()` 関数 + 4 個の `checkdepends` (= python-pytest / python-pytest-asyncio / python-pytest-randomly / python-pytest-timeout) を削除 | [nekono] は AUR helper 前提を取らない方針、 build host は `bin/build-all` が `--check` なし default 運用 |

## 結論

**approve** — build host で `bin/build-all python-uncalled-for` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release (= 0.3.x 等) が出たら、 **fastmcp の依存要件と整合**して
から bump する:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. fastmcp の `pyproject.toml` が新 version を許容するか確認 (= 不整合なら
   uncalled-for は据え置き、 fastmcp 側を先に bump する)
3. 本 dir の PKGBUILD + .SRCINFO を差し替え
4. sha256 を独立再計算
5. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-21 | 0.3.2 | (this PR) | `007051c7607cc1d77c04933bf67c8c166af2f65f` | safe-to-bump: depends/makedepends 変化なし、sha256 Issue #77 確認済み。fastmcp `>=0.2.0` 要件を満たす。Closes #77 |
| 2026-06-15 | 0.3.2-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
| 2026-05-19 | 0.2.0 | `cfcbb1e840b347748dc121d957fda2a7ffee46c5` | `b39a1115e11db09533461df5341e1c78a583a80d` | 初回 add、 純 fork (= prepare()/check() 削除のみ) |
| 2026-06-21 | 0.3.2-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
