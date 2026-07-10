# python-fastmcp review

## 状態

**review 済み、 approve** (最新: 2026-06-06 / 3.4.2)

AUR の `python-fastmcp` PKGBUILD を fork (+ 3.3.x の package split 対応)。

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
| 2026-05-21 | 3.3.1 | (this PR) | `PrefectHQ/fastmcp v3.3.1` | needs-attention: 3.3.0 で fastmcp PyPI が空 meta-pkg に分離。source を fastmcp-slim sdist に切替 + meta sdist (LICENSE 取得用) を追加 source。url を PrefectHQ/fastmcp に更新。python-multipart を depends に追加 (= [server] extra 新規 dep)。OAuth proxy security fix あり。Closes #76 |
| 2026-05-20 | 3.2.4 (pkgrel +1 → 2) | `7196acd19e029061f28f4c2dfd6bab4fc7af121b` | (同上) | dep audit で漏れていた 10 個を depends に追加 (= jsonref / cyclopts / py-key-value-aio [nekono] + opentelemetry-api / packaging / platformdirs / yaml / websockets / watchfiles / griffelib Arch 公式)。 ayaka 上の ModuleNotFoundError 'jsonref' 事案が契機 |
| 2026-05-26 | 3.3.1-4 | `7a5fe13` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-starlette 1.0.0-1 → 1.1.0-1 |
| 2026-05-23 | 3.3.1-3 | `14fc8c8` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-opentelemetry-api 1.42.0-1 → 1.42.1-1 |
| 2026-05-29 | 3.3.1-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-platformdirs 4.9.6-1 → 4.10.0-1 |
| 2026-06-03 | 3.4.0-1 | (this PR) | `PrefectHQ/fastmcp v3.4.0` | needs-attention: JWT 処理を `joserfc` に移行 → `python-joserfc` を depends 追加 (Arch 公式 extra)。 license を `MIT` → `Apache-2.0` に修正 (upstream は 3.3 以降 Apache-2.0、 PKGBUILD 側で長らく不一致)。 proxy `initialize` forwarding が breaking (3.3 silent fail → 3.4 loud fail)。 OAuth トークン寿命の切離し、 Code Mode default sandbox 制限、 OTEL spans 拡充、 GHA workflow security fix 3 件。 sha256 (slim / meta) 独立検証済み。 Closes #153 |
| 2026-06-06 | 3.4.2-1 | (this PR) | `PrefectHQ/fastmcp v3.4.2` | safe-to-bump: 3.4.0 → 3.4.2 (中間 v3.4.1 含む)。 build system (hatchling + uv-dynamic-versioning) / depends / build()・package() 変更なし。 v3.4.1 で `mcp` extra に `starlette>=1.0.1` floor 追加 (CVE-2026-48710 対応、 `python-starlette` は既に depends 済みで PKGBUILD 変更不要)。 v3.4.2 で JWTVerifier が private/non-critical な JWS header (Clerk `cat` 等) を署名検証前に拒否する問題を修正。 release author `jlowin` (= 過去と同一)。 sha256 (slim / meta) PyPI JSON API で独立検証済み。 Closes #175 |
| 2026-06-12 | 3.4.2-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-cyclopts 4.18.0-1, python-py-key-value-aio 0.4.4-6 (build-all の cascade-warn 由来) |
| 2026-06-15 | 3.4.2-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-cyclopts 4.18.0-2, python-mcp 1.27.2-3, python-openapi-pydantic 0.5.1+r37+g0766d59-4, python-py-key-value-aio 0.4.4-7, python-uncalled-for 0.3.2-2, python-uv-dynamic-versioning 0.14.0-2 (build-all の cascade-warn 由来) |
| 2026-06-17 | 3.4.2-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-mcp 1.27.2-4 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-21 | 3.4.2-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-cyclopts 4.18.0-3, python-jsonref 1.1.0-2, python-mcp 1.27.2-5, python-openapi-pydantic 0.5.1+r37+g0766d59-5, python-py-key-value-aio 0.4.4-8, python-uncalled-for 0.3.2-3 (build-all の cascade-warn 由来) |
| 2026-06-23 | 3.4.2-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1, python-anyio 4.13.0-1 → 4.14.0-1, python-cryptography 48.0.0-1 → 49.0.0-1, python-griffelib 2.0.2-2 → 2.1.0-1, python-hatchling 1.29.0-1 → 1.30.1-1, python-pydantic-settings 2.14.1-1 → 2.14.2-1, python-starlette 1.1.0-1 → 1.3.1-1, python-watchfiles 1.1.1-3 → 1.2.0-1, uvicorn 0.38.0-2 → 0.49.0-1 |
| 2026-06-24 | 3.4.2-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-py-key-value-aio 0.4.4-10 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-25 | 3.4.2-8 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-opentelemetry-api 1.42.1-1 → 1.43.0-1 |
| 2026-06-27 | 3.4.2-9 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-anyio 4.14.0-1 → 4.14.1-1 |
| 2026-07-01 | 3.4.2-10 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-mcp 1.27.2-8 → 1.28.1-1 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-02 | 3.4.2-11 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-py-key-value-aio 0.4.4-11 → 0.4.4-12 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-06 | 3.4.2-12 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-cyclopts 4.20.0-1 → 4.20.0-2 + python-py-key-value-aio 0.4.4-12 → 0.4.4-13 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-07 | 3.4.3-1 | (this PR) | `PrefectHQ/fastmcp v3.4.3` (tag commit `1eedd1f`) | safe-to-bump: 3.4.2 → 3.4.3。 "a month of SSRF and OAuth hardening" のセキュリティ中心リリース (NAT64/6to4/Teredo/ISATAP IPv6 transition address の SSRF allow-list bypass を block、 Streamable HTTP に Host/Origin 検証追加で DNS rebinding 防止、 OAuth redirect URI の `javascript:`/`data:`/`file:` 等 unsafe scheme 拒否、 `fastmcp dev` の HTML/JSON embedding escaping 修正 = XSS)。 build system (hatchling + uv-dynamic-versioning) / depends / build()・package() 変更なし (meta sdist の dev-dep `ty` floor 変化のみで PKGBUILD 影響なし)。 sha256 (slim / meta) を PyPI から直接 download + `sha256sum` 独立検証 (byte 一致)。 併せて `.deps.lock` の python-typing_extensions 4.15.0-3 → 4.16.0-1 を更新 (closed bot PR #359 の dep change を fold-in)。 Closes #356 |
| 2026-07-09 | 3.4.3-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-py-key-value-aio 0.4.4-14 → 0.4.4-15 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-10 | 3.4.4-1 | (this PR) | `PrefectHQ/fastmcp v3.4.4` (tag commit `9138d40`, GitHub `verified`) | needs-attention → 精査後 approve: 3.4.3 → 3.4.4。 upstream の sdist 全文 diff で `pyproject.toml` (build-system/deps/optional-deps) byte 一致、差分は `server/http.py`・`mixins/transport.py`・`settings.py`・新規 `auth/providers/huggingface.py` の 4 file のみ。 build()/package()・depends/makedepends 変更なし、 sha256 (slim/meta) を `makepkg --verifysource` で PyPI から独立検証 (byte 一致)。 **⚠️ security 注記**: 3.4.3 で strict default だった `http_host_origin_protection` が 3.4.4 で default `False` に緩和 (upstream が既存デプロイ互換のため意図的に後退) → ただし本 repo の依存元 `nekono-pipewire-mcp`/`nekono-voicevox-mcp` は **どちらも claude-code から stdio 起動 (HTTP server transport 不使用)** のため Host/Origin guard は無関係で実害なし。 新規 `huggingface.py` は opt-in OAuth provider (明示設定時のみ import)。 Closes #376 |
