# python-jsonref review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR `python-jsonref` 1.1.0-1 を **純 fork** (= 差分は Maintainer 行のみ)。

## 用途

`python-fastmcp` 3.2.4 の **直接依存** (= upstream pyproject.toml の
`dependencies = [..., "jsonref>=1.1.0", ...]`)。 fastmcp dep audit (= 22 dep を
Arch + [nekono] と diff) で漏れが発覚した 10 個のうちの 1 つ、 [nekono] 投入が
必要だった分。

ayaka 上の自家 MCP server (= [[nekono-pipewire-mcp]] / [[nekono-voicevox-mcp]])
を pacman 経由で完全 install できるようにする取り組みの一環。

## Source

- Upstream: https://github.com/gazpachoking/jsonref
  - upstream maintainer: Chase Sterling
  - tag `v1.1.0` (= prefix "v")、 commit (= sdist 由来、 GitHub tag と同一性は
    PyPI digest と一致で担保)
  - license: MIT
- AUR: https://aur.archlinux.org/packages/python-jsonref
  - AUR maintainer: piernov <piernov@piernov.org>
  - AUR PKGBUILD と diff は Maintainer 行のみ
- 当 PKGBUILD は AUR と同様 **PyPI sdist** (`jsonref-1.1.0.tar.gz`) を vendor

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/aa/0d/.../jsonref-1.1.0.tar.gz`
  - PyPI 公式 CDN、 PEP 503 形式 path
- [x] `sha256sums` 独立検証
  - build host で curl + sha256sum で実測: `32fe8e1d85af0fdefbebce950af85590b22b60f9e95443176adbde4e1ecea552`
  - claude-review 初回 verdict で AUR の md5sums は [nekono] 規約違反と指摘されたため、 hash 種別を sha256 に **bump 即時**
- [x] `build()`: `python -m build --wheel --no-isolation`
  - build-backend = `pdm.pep517.api` → makedepends `python-pdm-pep517` (= [nekono]
    別 PR で同時投入) が解決する
- [x] `package()`: `python -m installer --destdir` のみ
- [x] `depends`: python のみ
  - upstream `pyproject.toml` の runtime deps と一致
- [x] `makedepends`: python-pdm-pep517 (= [nekono]) + python-build / python-installer (= Arch)
- [x] `secrets` 混入なし
- [x] `arch=('x86_64')` — AUR 通り (= pure Python だが AUR maintainer の判断踏襲)

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | piernov → Nekono、 piernov を Contributor に降格 | [nekono] fork 表示 |
| `md5sums` → `sha256sums` | AUR の md5 を捨て、 [nekono] 規約 (CLAUDE.md 「sha256sums (or sha512sums) で tarball が pin されているか」) に従って sha256 を独立計算して固定 | MD5 は暗号学的に破綻 (= 衝突攻撃成立)、 supply-chain 監査上 keep するのは不可。 AUR にも修正 PR 送るのが筋だが [nekono] では先に正しい状態を反映 |

`arch=('x86_64')` / LICENSE 非設置 等は AUR 踏襲 (= LICENSE 設置は別 PR で別途)。

## 結論

**approve** — build host で `bin/build-all python-jsonref` で build + sign + repo db 追加可。 ただし **python-pdm-pep517 PR (= PR #67) を先に publish してから** でないと makedep が解決しない。

## 更新方針

upstream で新 release が出たら:
1. nvchecker が検知 → Issue 経由で人間が手作業更新
2. PKGBUILD pkgver + md5sums (or 後日 sha256sums) を更新
3. depends 変化があれば反映
4. `.SRCINFO` + `.deps.lock` を再生成
5. REVIEW.md 更新履歴に追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 1.1.0-1 | `b5896503a6339ab5fc42f2b1ecc119f90261ade4` | PyPI sdist `jsonref-1.1.0.tar.gz` | 初回 add、 AUR 純 fork、 fastmcp 直接依存 |
