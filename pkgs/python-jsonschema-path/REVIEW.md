# python-jsonschema-path review

## 状態

**review 済み、 approve** (2026-05-19、 初回 add)

PyPI sdist (`jsonschema_path-0.4.6.tar.gz`) を直接 vendor する自家 PKGBUILD。
AUR にも該当 pkg が無い (= 2026-05-19 時点で AUR search hit 0)。

## 用途

`python-fastmcp` 依存 chain (= `python-fastmcp` → `python-mcp` →
`python-openapi-pydantic` → `python-jsonschema-path`) の **葉ノード** として
必要。 fastmcp/mcp 系の transitive deps の中で唯一 Arch 公式 + AUR の
両方に存在しない pkg なので、 PyPI から直接 vendor して [nekono] に投入する。

ayaka 上の自家 MCP server (= `nekono-pipewire-mcp` / `nekono-voicevox-mcp`) を
pacman 経由 install できるようにする取り組みの一環。

## Source

- PyPI: https://pypi.org/project/jsonschema-path/
  - upstream maintainer: Artur Maciag <maciag.artur@gmail.com>
- upstream repo: https://github.com/p1c2u/jsonschema-path
  - tag `0.4.6` (= **prefix なし**)、 commit `a3de101627714f15d8d0699950f35d1f811a90a4`
  - license: Apache-2.0
- 当 PKGBUILD は **GitHub archive ではなく PyPI sdist** を採用 (= PyPI 配信は
  upstream 自身が build した正規 distribution であり、 GitHub archive より
  reproducibility が高い)

## 検証結果

- [x] `source` URL = `files.pythonhosted.org/packages/source/j/jsonschema_path/jsonschema_path-0.4.6.tar.gz`
  - PyPI 公式 CDN、 path 構造は PEP 503 のコンテンツアドレス (= path
    `<l>/<name>/<filename>`、 `<l>` は `name[:1]`)
  - typosquat / domain spoof リスクなし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `c89eb635f4d497c9ac328eeff359c489755838806a7d033510a692e9576f5c4b`
  - PyPI digests と一致
- [x] `build()`: `python -m build --wheel --no-isolation` のみ
  - build-backend: `poetry-core` (= makedepends `python-poetry-core` で provide)
  - ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` のみ
  - wheel 内 dist-info 経由でファイル配置、 shell injection / exec / curl / wget なし
  - LICENSE は sdist に含まれる前提で `install -Dm644` (= 不在時 skip)
- [x] `depends`: python / python-yaml / python-pathable / python-referencing
  - すべて Arch 公式 extra に存在 (= `.deps.lock` 参照)
  - upstream `pyproject.toml` の runtime deps と一致 (= PyYAML>=5.1 / pathable<0.6.0,>=0.5.0 / referencing<0.38.0)
  - `requests` は upstream で extras 扱い (= `[requests]`)、 我々の用途 (= mcp/fastmcp chain) では不要
- [x] `makedepends`: python-build / python-installer / python-wheel / python-poetry-core — 標準的な poetry-core build pattern
- [x] `secrets` 混入なし

## 設計判断

| 判断 | 理由 |
|---|---|
| **AUR fork ではなく PyPI sdist 直接 vendor** | AUR search hit 0 (= 2026-05-19、 `python-jsonschema-path` / `jsonschema-path` 両 query)。 fork 元が無いので自家 PKGBUILD 化が唯一の経路 |
| **GitHub archive ではなく PyPI sdist** | PyPI sdist は upstream maintainer が build した正規 distribution。 GitHub archive は git 状態の snapshot で sdist と微妙に差分が出ることがある (= MANIFEST.in 等で除外される file) |
| **arch=('any')** | pure Python、 C extension 無し |
| **`requests` は depends に含めない** | upstream `pyproject.toml` で extras 扱い (= optional)、 fastmcp/mcp chain では requests 経路を使わない |
| **pkgname に `python-` prefix** | Arch convention、 `python-` prefix なし命名例外 (= uvicorn / pipewire 等) には該当しない |

## 結論

**approve** — build host で `bin/build-all python-jsonschema-path` で build + sign + repo db 追加可。

本 pkg は **fastmcp 依存 chain の前提** として投入するもので、 直接 user に install
されることは想定していない (= 後段の `python-openapi-pydantic` (PR #?) → `python-mcp`
→ `python-fastmcp` を経て `nekono-{pipewire,voicevox}-mcp` の install で transitive
に install される pkg)。

## 更新方針

upstream で新 release (= 0.4.7 等) が出たら:
1. PyPI から sdist URL + sha256 を取得 (= `curl -s https://pypi.org/pypi/jsonschema-path/json | jq ...`)
2. PKGBUILD の `pkgver` + `sha256sums` を更新
3. `pyproject.toml` の runtime deps が変わっていれば `depends` を反映 (= 特に upper bound 緩和等)
4. `.SRCINFO` + `.deps.lock` を再生成
5. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.4.6 | (本 PR の commit SHA を merge 時に追記) | `a3de101627714f15d8d0699950f35d1f811a90a4` | 初回 add、 PyPI sdist 直接 vendor (= AUR 不在のため自家 PKGBUILD 化) |
