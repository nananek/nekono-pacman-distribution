# python-fastmcp review

## 状態

**review 済み、 approve** (2026-05-20、 初回 add)

AUR の `python-fastmcp` PKGBUILD (pkgver=3.2.4, pkgrel=1) を fork。 純 fork
(= 機能変化なし、 改変は Maintainer 行 + nekono 説明コメントのみ)。
AUR 元 PKGBUILD は check() / checkdepends / optdepends が空で minimal な
ため、 [nekono] convention の削除作業は不要。

## 用途

`nekono-pipewire-mcp` (PR #8) / `nekono-voicevox-mcp` (PR #9) の **直接依存**。
fastmcp chain の頂点、 ayaka 上の MCP server runtime を支える Python 製
MCP framework。

ここまでに投入した 7 個の [nekono] fork (= jsonschema-path / httpx-sse /
sse-starlette / openapi-pydantic / uncalled-for / mcp / uv-dynamic-versioning)
全て resolve できることが本 PKGBUILD build の前提。 build host で publish
済みであることを確認した。

## Source

- AUR: https://aur.archlinux.org/packages/python-fastmcp
  - Maintainer: Rafael Dominiquini <rafaeldominiquini at gmail dot com>
- Upstream: https://github.com/jlowin/fastmcp
  - **NOTE**: repo が `jlowin/fastmcp` → `PrefectHQ/fastmcp` に migration 済み
    (= GitHub 上の redirect 有効、 API / archive DL は jlowin alias で問題なく動作)。
    AUR PKGBUILD は `jlowin/fastmcp` のままなので [nekono] も同じく踏襲、
    将来 jlowin alias が廃止されたら PrefectHQ に直接書き換える (= 更新方針参照)
  - 元 author: jlowin (Jeremiah Lowin)、 MIT
  - 現 owner: PrefectHQ (Prefect Technologies)
  - tag `v3.2.4`、 commit `7d7607473d7713d9937cbbbe0bfc635976c511d3` (= PrefectHQ/fastmcp 上)
- distfile: **PyPI sdist** (= `files.pythonhosted.org/packages/source/f/fastmcp/fastmcp-3.2.4.tar.gz`)

## 検証結果

- [x] `source` URL = PyPI 公式 CDN、 PEP 503 path 構造、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `083ecb75b44a4169e7fc0f632f94b781bdb0ff877c6b35b9877cbb566fd4d4d1`
  - AUR PKGBUILD と一致
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + LICENSE / README.md / AGENTS.md / SECURITY.md / CODE_OF_CONDUCT.md 配置のみ
  - shell injection / curl / wget / eval なし
  - 全ての doc / license は upstream tarball 同梱、 sha 検証済み tarball 内
- [x] `depends`: 20 entry、 全 resolve OK
  - Arch 公式 16 個 (= python / uvicorn / python-dotenv / python-pydantic-settings 等)
  - [nekono] 4 個 (= python-openapi-pydantic [PR #58] / python-mcp [PR #60] / python-jsonschema-path [PR #54] / python-uncalled-for [PR #57])
- [x] `makedepends`: 6 entry、 全 resolve OK
  - Arch 公式 5 個、 [nekono] 1 個 (= python-uv-dynamic-versioning [PR #59])
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR Maintainer は `# Upstream AUR Maintainer:` として comment 保持 | 当 PKGBUILD は Nekono が責任を持つ |
| `pkgrel` を 1 → 2 にバンプ、 `depends` に **10 個追加** (= jsonref / cyclopts / py-key-value-aio + opentelemetry-api / packaging / platformdirs / yaml / websockets / watchfiles / griffelib) | **fastmcp dep audit で AUR + 当初の [nekono] fork が漏らしていた 10 個を補完**。 ayaka で `ModuleNotFoundError: No module named 'jsonref'` が出た事案を契機に upstream `pyproject.toml` の dependencies と diff を取って洗い出した。 AUR 側にも修正 PR を別途送るのが筋だが [nekono] では先に正しい状態を反映 |

## 結論

**approve** — build host で `bin/build-all python-fastmcp` で build + sign + repo db 追加可。

これが merge + publish されれば PR #8 (= `nekono-pipewire-mcp`) と PR #9 (= `nekono-voicevox-mcp`)
が depends 解決可能になり、 fastmcp chain 投入の山場 (PR #54〜#60 + #59) が報われる。

## 更新方針

upstream で新 release (= v3.2.5 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. 本 dir の PKGBUILD の pkgver を更新、 sha256 を独立再計算 (= PyPI digests + curl)
3. depends に新 transitive が増えていれば確認 (= Arch 公式 / [nekono] / 新規 fork の判断、
   特に AUR-only な新 dep が出てきたらそちら も先に [nekono] に投入してから fastmcp を bump)
4. もし `jlowin/fastmcp` redirect が将来切れたら `url=` を `PrefectHQ/fastmcp` に書き換え、
   nvchecker.toml の `github = "jlowin/fastmcp"` も同じく PrefectHQ に変更
5. `.SRCINFO` + `.deps.lock` を再生成
6. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 3.2.4 | `a5d709ad36d834b5e5a42d30eae67c37cd0d821f` | `7d7607473d7713d9937cbbbe0bfc635976c511d3` | 初回 add、 純 fork (= Maintainer 行のみ改変)。 fastmcp chain 頂点 |
| 2026-05-20 | 3.2.4 (pkgrel +1 → 2) | `7196acd19e029061f28f4c2dfd6bab4fc7af121b` | (同上) | dep audit で漏れていた 10 個を depends に追加 (= jsonref / cyclopts / py-key-value-aio [nekono] + opentelemetry-api / packaging / platformdirs / yaml / websockets / watchfiles / griffelib Arch 公式)。 ayaka 上の ModuleNotFoundError 'jsonref' 事案が契機 |
