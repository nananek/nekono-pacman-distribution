# python-openapi-pydantic review

## 状態

**review 済み、 approve** (2026-05-19、 初回 add)

AUR の `python-openapi-pydantic` PKGBUILD (pkgver=0.5.1+r37+g0766d59, pkgrel=3)
を fork。 純 fork (= 機能変化なし、 改変は Maintainer 行 + nekono 説明
コメント + check() 削除のみ)。

## 用途

`python-fastmcp` の transitive 依存 (= fastmcp → mcp → openapi-pydantic)。
Arch 公式 repo 不在で AUR にしか無いため [nekono] に fork 投入する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-openapi-pydantic
  - Maintainer: Vitalii Kuzhdin <vitaliikuzhdin@gmail.com>
- Upstream: https://github.com/mike-oakley/openapi-pydantic
  - Mike Oakley、 MIT
  - tag `v0.5.1` (= release commit `82fd769f2820f6ff2acf650f551a97a7b8fbe189`)
  - **AUR / 我々が pin している commit: `0766d599bbe9bccda12b6ede069647c7bef2299f`**
    - 0.5.1 release 後の +37 commits の snapshot
    - 2025-10-17 merge の PR #73 "Add support for Python 3.14 and drop ..." を含む
    - **理由**: 0.5.1 release は Python 3.14 サポートを持たない、 Arch system Python が
      3.14 なので、 release tag ではなくこの commit を pin (= AUR maintainer の判断踏襲)

## 検証結果

- [x] `source` URL = `github.com/mike-oakley/openapi-pydantic/archive/0766d599.../openapi-pydantic-0766d599...tar.gz`
  - upstream 公式 GitHub commit archive、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `8b4091f4a2f1613a05aa88a66ad72e0b1c3da7781be1a207d0d7edff4483ffcf`
  - AUR PKGBUILD と一致
- [x] **commit `0766d599` の出自確認**: GitHub API で確認、 mike-oakley/openapi-pydantic master 上の
  legitimate commit (= PR #73 merge、 author = Mike Oakley、 date = 2025-10-17)
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + README/LICENSE 配置のみ
  - LICENSE は site_packages 内の wheel dist-info から symlink (= upstream LICENSE.txt が
    sdist root に無く dist-info 内に配置される pattern)
- [x] `depends`: python>=3.9 / python-pydantic>=1.8 / python-pydantic-core — 全て Arch 公式 extra
  - upstream `pyproject.toml` runtime deps は `pydantic>=1.8` のみ、 pydantic-core は pydantic の transitive
- [x] `makedepends`: python-build / python-hatchling>=1.26 / python-installer / python-wheel — 全て Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer は `# Upstream AUR Maintainer:` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つ |
| `check()` 関数 + 2 個の `checkdepends` (= python-pytest / python-openapi-spec-validator) を削除 | [nekono] は AUR helper 前提を取らない、 build host は `bin/build-all` が `--check` なし default 運用。 特に `python-openapi-spec-validator` は AUR 限定 |

PKGBUILD 本体 (= `_commit_rel` / `_commit` / pkgver / pkgrel / depends / makedepends / source /
sha256sums / build / package) は upstream と完全同一。

## 結論

**approve** — build host で `bin/build-all python-openapi-pydantic` で build + sign + repo db 追加可。

## 更新方針

upstream の状況次第で 2 パターン:

### パターン A (= 推奨): 0.6.0 等の clean tag release が出たら
1. AUR PKGBUILD で snapshot pin が解除されているか確認
2. もし AUR が release tag に乗り換えていれば、 当 PKGBUILD も追従:
   - `_commit_rel` / `_commit` 変数を削除、 `pkgver=0.6.0` の clean 形式に
   - source URL を `archive/v0.6.0.tar.gz` 形式に変更
   - sha256 再計算
3. `.SRCINFO` + `.deps.lock` を再生成
4. REVIEW.md「更新履歴」 に 1 行追記

### パターン B: 0.5.1 続きの新 snapshot commit に乗り換える場合
- AUR が `_commit` を更新したら同じ手順、 ただし `pkgver=0.5.1+rNN+gXXXXXXX` の commit-pinned snapshot 形式を維持

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream commit (pinned) | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.5.1+r37+g0766d59 | `f7ab982b98becf14be0afba058610f0d86b5cc0d` | `0766d599bbe9bccda12b6ede069647c7bef2299f` (= 0.5.1 後 PR #73 merge、 Python 3.14 対応) | 初回 add、 純 fork (= check() 削除のみ)。 AUR の snapshot pin 戦略 (= release tag 0.5.1 に Python 3.14 対応が無いため commit pin) を踏襲 |
| 2026-06-15 | 0.5.1+r37+g0766d59-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
| 2026-06-21 | 0.5.1+r37+g0766d59-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
