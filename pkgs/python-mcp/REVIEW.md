# python-mcp review

## 状態

**review 済み、 approve (条件付き)** (2026-05-20、 初回 add)

AUR の `python-mcp` PKGBUILD (pkgver=1.27.1, pkgrel=2) を fork。 改変は
Maintainer 行 / check() 削除 + depends に upstream pyproject hard dep の
`python-opentelemetry-api` を追加 (= AUR では省略されていた)。

## 用途

`python-fastmcp` の直接依存。 fastmcp → mcp の transitive chain の核心、
Model Context Protocol SDK 本体。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-mcp
  - Maintainer: Mohamed Amine Zghal (medaminezghal) <medaminezghal at outlook dot com>
- Upstream: https://github.com/modelcontextprotocol/python-sdk
  - Anthropic / Model Context Protocol、 MIT
  - tag `v1.27.1` (= lightweight tag)、 commit `77431ebe7dda9ed0c61451b22d3e7f8d981bc092`
- patch file: `1834.patch` (= AUR repo 同梱、 upstream PR #1834 "Support Python 3.14" by Marcelo Trylesinski)
  - 当 dir 配下に local file として保持、 sha256 pin

## 検証結果

- [x] `source` URL = `git+github.com/modelcontextprotocol/python-sdk.git#tag=v1.27.1` + local `1834.patch`
  - upstream 公式 modelcontextprotocol org、 typosquat なし
- [x] `sha256sums` 独立検証
  - `mcp` (git clone tag verification): `65ef8600938fbf3b1e5b96712c56f58cd5029783143e88e3aa4f4badb74d00d9`
  - `1834.patch` (local file): `b0447596b4e75375c4a37759c088a867b2d7d3e64cb05b78d71762a90bceba0c`
  - AUR PKGBUILD と一致、 build host `makepkg --verifysource` 通過
- [x] `1834.patch` 内容確認: upstream の本物 PR #1834 (= author Marcelo Trylesinski、 starlette pyproject の `--cov` 等を Python 3.14 対応に修正)、 typosquat / 改竄なし
- [x] `prepare()`: `git clean -fdx` + `patch -Np1 < 1834.patch` + `sed -i 's/timeout=20/timeout=None/' tests/client/test_config.py`
  - test の timeout を伸ばす sed (= check で意味あるが check() 削除後は dead code、 但し AUR 踏襲で残す)
  - 1834.patch 適用、 ネットワーク取得なし
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + LICENSE symlink (= upstream LICENSE は wheel dist-info 内に配置されるため symlink で /usr/share/licenses/ に向ける)
- [x] `depends`: 全 16 entry いずれも Arch 公式 extra + [nekono] (= python-httpx-sse / python-sse-starlette は PR #55/#56 で投入済み)
  - upstream pyproject.toml hard dep `opentelemetry-api>=1.28.0` を AUR は省略 → 我々で追加 (= REVIEW.md§AUR との意図的差分 参照)
- [x] `makedepends`: 6 entry すべて Arch 公式 + [nekono] (= python-uv-dynamic-versioning は PR #59 で投入済み)
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer は `# Upstream AUR Maintainer:` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つ |
| `check()` 関数 + 13 個の `checkdepends` を削除 | [nekono] は AUR helper 前提を取らない、 bin/build-all は --check なし default 運用。 checkdepends には `python-pytest-examples` / `python-inline-snapshot` / `python-dirty-equals` 等の AUR 限定 pkg や `uv` (= 多くは [nekono] に投入したくない深い chain) を含む |
| `optdepends` を削除 (= `python-rich: rich` / `python-typer: cli` / `python-dotenv: cli` / `python-websockets: ws`) | MCP server (= 我々の用途) として起動するには不要な CLI / 補助機能。 必要なら user が個別 install |
| `depends` に **`python-opentelemetry-api` を追加** | upstream `pyproject.toml` で hard dep 宣言 (= `"opentelemetry-api>=1.28.0"`)、 AUR は省略していたが upstream 準拠で組み込む。 Arch 公式 extra にあるので resolve OK (1.42.0-1) |

PKGBUILD 本体 (= pkgver / pkgrel / prepare / build / package / source / sha256sums) は AUR と完全同一。

## 結論

**approve (条件付き)** —

build host で `bin/build-all python-mcp` で build + sign + repo db 追加可。 条件:
- PR #55 (python-httpx-sse) / PR #56 (python-sse-starlette) / PR #59 (python-uv-dynamic-versioning)
  が事前に publish 済みであること (= 本 PR open 時点では全て confirmed publish 済み)
- patch file (`1834.patch`) は upstream PR #1834 が今後 merge されたら不要になる
  (= 次の upstream release で吸収) — その時点で当 PKGBUILD からも削除する

## 更新方針

upstream で新 release (= 1.27.2 or 1.28.x 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256 / patch file の必要性を確認
2. もし upstream 新 release に PR #1834 が merge 済みなら `1834.patch` 不要 (= source[] と sha256sums から削除、 prepare() の `patch` 行も削除)
3. mcp の pyproject.toml が新 hard dep を追加していれば depends 反映 (= 今回 `opentelemetry-api` の hard dep を AUR が見落としていた pattern が再発しないか確認)
4. `.SRCINFO` + `.deps.lock` 再生成
5. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 1.27.1-2 | `c3370e73e2fe9588b3a6a4f0df8ecd8437ac9924` | `77431ebe7dda9ed0c61451b22d3e7f8d981bc092` | 初回 add、 純 fork ベース (= check() / optdepends 削除) + upstream 準拠で `python-opentelemetry-api` を depends 追加 (= AUR 省略 fix)。 1834.patch 同梱 (= Python 3.14 対応、 upstream 未 merge) |
| 2026-05-22 | 1.27.1-3 | `6e9889f` | `77431ebe7dda9ed0c61451b22d3e7f8d981bc092` | `pkgrel` +1 (deps changed): python-opentelemetry-api 1.42.0-1 → 1.42.1-1。 PKGBUILD 本体変更なし |
