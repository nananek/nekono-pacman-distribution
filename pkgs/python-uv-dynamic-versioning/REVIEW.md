# python-uv-dynamic-versioning review

## 状態

**review 済み、 approve** (2026-05-19、 初回 add)

AUR の `python-uv-dynamic-versioning` PKGBUILD (pkgver=0.14.0, pkgrel=1) を
fork。 純 fork (= 機能変化なし、 改変は Maintainer 行 + nekono 説明
コメント + sha512sums → sha256sums 統一のみ)。

## 用途

`python-mcp` (PR #6) の **makedepends** として必要。 mcp の pyproject.toml は
git tag から動的に version を解決する hatchling plugin (`uv-dynamic-versioning`)
を build 時に要求する。

Arch 公式 + AUR の他 mirror ではこれ単体の存在が AUR にしかなく、 fastmcp chain
の build を完結させるため [nekono] に fork 投入する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-uv-dynamic-versioning
  - Maintainer: Butui Hu <hot123tea123@gmail.com>
- Upstream: https://github.com/ninoseki/uv-dynamic-versioning
  - ninoseki、 MIT
  - tag `v0.14.0`、 commit `46cb92b53a9b822cee62572a0a14aa0be7cd045d`

## 検証結果

- [x] `source` URL = `github.com/ninoseki/uv-dynamic-versioning/archive/refs/tags/v0.14.0.tar.gz`
  - upstream 公式 GitHub Release archive、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `curl + sha256sum`): `b31e00f9a4ec58b751c470d4f6608fd645b35a6161b65675b6dd08ff874f4dd0`
  - AUR PKGBUILD は sha512 だったので独立に sha256 算出 (= [nekono] は sha256 統一)
  - makepkg --verifysource で検証通過
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + `install -Dm644 LICENSE` のみ、 shell injection / curl / wget なし
- [x] `depends`: python-tomlkit / python-pydantic / python-jinja / python-hatchling / python-dunamai — 全て Arch 公式 extra
- [x] `makedepends`: python-build / python-installer / python-wheel — 全て Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer は `# Upstream AUR Maintainer:` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つ |
| `sha512sums` → `sha256sums` に変更 | [nekono] は sha256 統一 (= bin/build-all は両対応だが convention) |

PKGBUILD 本体 (= pkgver / pkgrel / depends / makedepends / source URL / build / package) は upstream と完全同一。

## 結論

**approve** — build host で `bin/build-all python-uv-dynamic-versioning` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release (= 0.15.0 等) が出たら:
1. AUR PKGBUILD の pkgver / sha512sums を確認
2. 本 dir の PKGBUILD の pkgver を更新、 sha256 を独立再計算 (= AUR は sha512、 我々は sha256 統一)
3. `.SRCINFO` + `.deps.lock` を再生成
4. mcp 側の hatch-plugin 互換性 (= breaking change の有無) を REVIEW.md 「更新方針」 に短評
5. 本 REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.14.0 | (初回 PR commit SHA、 push 後に固定) | `46cb92b53a9b822cee62572a0a14aa0be7cd045d` | 初回 add、 純 fork (= sha512 → sha256 統一のみ) |
