# python-pdm-pep517 review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR `python-pdm-pep517` 1:1.1.4-5 を **純 fork** (= 差分は Maintainer 行のみ)。

## 用途

`python-jsonref` (= 別 PR で並行投入) の **build-time backend**。 jsonref の
`pyproject.toml` は `build-backend = "pdm.pep517.api"` を指定するため、
makepkg が `python -m build --no-isolation` する時にこれが import 解決
できる必要がある。 jsonref 経由で `python-fastmcp` chain の最終 dep。

ayaka 上の自家 MCP server (= [[nekono-pipewire-mcp]] / [[nekono-voicevox-mcp]])
を pacman 経由で完全 install できるようにする取り組みの「fastmcp dep 漏れ
fix」 (= upstream pyproject.toml audit で 10 個漏れ発覚) の一環。

## Source

- Upstream: https://github.com/pdm-project/pdm-pep517
  - tag は無し、 PyPI sdist (`pdm-pep517-1.1.4.tar.gz`) を直接 vendor
  - license: MIT
- AUR: https://aur.archlinux.org/packages/python-pdm-pep517
  - AUR maintainer: David Runge <dvzrv@archlinux.org>
  - AUR PKGBUILD と diff は Maintainer 行のみ (= 純 fork)
- 同梱 patch: `python-pdm-pep517-1.1.2-devendor.patch` (= AUR と同一、
  vendor された pip 等を strip して system Python pkg を使うようにする)

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/p/pdm-pep517/pdm-pep517-1.1.4.tar.gz`
  - PyPI 公式 CDN
- [x] `sha512sums` 独立検証
  - AUR PKGBUILD の sha512 と一致 (= 純 fork なので確認のみ)
- [x] devendor patch (= 同梱) の sha512 も AUR と一致
- [x] `prepare()`: patch 適用 + `_vendor` dir 削除のみ
- [x] `build()`: `python -m build --wheel --skip-dependency-check --no-isolation`
- [x] `package()`: `installer` + LICENSE / README 配置
- [x] `check()`: AUR で disabled (= 「This is disastrous on AUR」 コメント)、 [nekono] でも引き続き disabled
- [x] `depends`: python / python-cerberus / python-license-expression / python-packaging / python-tomli / python-tomli-w — 全 Arch 公式 extra
- [x] `makedepends`: python-build / python-installer — Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 項目 | 差分 | 理由 |
|---|---|---|
| `# Maintainer:` | David Runge → Nekono、 David を Contributor に降格 | [nekono] fork 表示 |

それ以外は **0 行差分** (= 純 fork)。 epoch (1:) / devendor flag / patch 同梱 / check() disabled も全て AUR 踏襲。

## 結論

**approve** — build host で `bin/build-all python-pdm-pep517` で build + sign + repo db 追加可。

## 更新方針

- AUR で bump があれば nvchecker (= `[python-pdm-pep517]` section) が検知 → Issue 経由で人間が手作業更新
- AUR maintainer (David Runge) の commit 履歴を参考に diff を取る
- devendor patch も pkgver bump 時に再点検 (= 1.1.4 以降で path 変化があれば patch 修正)

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag / commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 1:1.1.4-5 | `159506c629eddd7a485a2e5d9eaab5f308503d0f` | PyPI sdist `pdm-pep517-1.1.4.tar.gz` | 初回 add、 AUR 純 fork、 python-jsonref makedep 経由で fastmcp chain 完成のため必要 |
