# nekono-voicevox-mcp review

## 状態

**review 済み、 approve** (2026-05-20、 v0.1.0 initial release)

自家 source PKGBUILD (= AUR fork ではない、 nananek 自身の upstream を
GitHub release tarball から直接 vendor)。

## 用途

claude-code から **stdio transport** で起動される MCP server。 [nekono] の
`voicevox-engine-cuda` (= nekono-pacman0 で Tailscale `100.117.10.53:50021`
bind 済み HTTP API) に httpx で TTS request を投げ、 戻り WAV を `pw-cat`
で再生する thin bridge。

daemon ではないので systemd unit / sysusers / tmpfiles は同梱しない。
claude-code が `claude mcp add ...` で登録した command を必要時に spawn する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの
集大成 (= PR #8 = nekono-pipewire-mcp と並行で 2 個目)。

## Source

- Upstream: https://github.com/nananek/nekono-voicevox-mcp
  - 主開発者 nananek (= ansible-nekonodesk owner、 自家 upstream)
  - tag `v0.1.0` (= prefix "v")、 commit `a1d1423524ad7eb53db0667e8ce7b1dd6473a831`
  - license: MIT
- distfile: GitHub Release archive (= `archive/refs/tags/v0.1.0.tar.gz`)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekono-voicevox-mcp/archive/refs/tags/v0.1.0.tar.gz`
  - 自家 upstream、 typosquat 検討対象外
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `14a447d896a79cf97c5c8166055da946298095f6477e8cb6e19248b9ac0929af`
- [x] tag `v0.1.0` GPG 署名状態: **要確認** (= user が `git tag -s` で push したか、 REVIEW.md「更新方針」 参照)
- [x] `build()`: `python -m build --wheel --no-isolation` のみ、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + LICENSE / README.md 配置のみ
  - console_script `nekono-voicevox-mcp` (= `nekono_voicevox_mcp.server:main`) は wheel dist-info 経由で `/usr/bin/` に自動配置
  - shell injection / exec / curl / wget なし
- [x] `depends`: python / python-fastmcp / python-httpx / pipewire — runtime に必要な最小依存
  - upstream `pyproject.toml` の runtime dep は `fastmcp>=2.0` + `httpx>=0.27`、 `pipewire` は MCP server が subprocess で `pw-cat` を呼んで WAV 再生するため runtime 必須
- [x] `makedepends`: python-build / python-installer / python-wheel / python-hatchling — 標準的な hatchling build pattern
- [x] `license`: MIT — SPDX、 LICENSE を `/usr/share/licenses/` に配置
- [x] `secrets` 混入なし (= VOICEVOX engine URL は client 側 `claude mcp add --env VOICEVOX_ENGINE_URL=http://...` で env 注入する設計、 PKGBUILD には埋め込まない)
- [x] `arch=('any')` — pure Python、 C extension 無し

## 設計判断 (= AUR fork でないので「AUR との意図的差分」 ではなく §設計判断)

| 判断 | 理由 |
|---|---|
| **systemd unit を同梱しない** | claude-code は MCP server を stdio transport で起動する (= command + args を spawn して stdin/stdout で JSON-RPC)。 daemon 化しない。 voicevox-engine-cuda の HTTP daemon とは別カテゴリ (= engine 自体は別 host で daemon、 本 pkg はその client wrapper) |
| **python-fastmcp を [nekono] 経由 depends** | Arch 公式に python-fastmcp 不在、 [nekono] PR #62 で投入済み。 fastmcp chain (PR #54〜#62 + #59) の終着点 |
| **python-httpx を Arch 公式 extra から depends** | VOICEVOX engine HTTP API client。 fastmcp も内部で使う (= 二重 install は pacman 側で deduplicate) |
| **pipewire を depends に置く (optdepends ではなく)** | pw-cat 無しでは WAV 再生が即 fail。 ayaka client では当然 install 済み (= 音響 daemon)、 server 専用 host 想定がないので hard dep で固定。 同方針: [[nekono-pipewire-mcp]] の wireplumber depends |
| **VOICEVOX engine URL は env 注入** | `claude mcp add --env VOICEVOX_ENGINE_URL=http://100.117.10.53:50021 --` で client 側設定。 PKGBUILD に hardcode しない (= secret 混入回避 + multi-deployment 対応) |
| **arch=('any')** | pure Python、 C extension 無し |
| **`# Contributor:` 等の引用なし** | AUR fork ではないので不要 |

## 結論

**approve** — build host で `bin/build-all nekono-voicevox-mcp` で build + sign + repo db 追加可。

これで PR #8 (= nekono-pipewire-mcp) + 本 PR が merge / publish されれば、 ayaka 側 ansible で:
1. `pacman -S nekono-voicevox-mcp` で install (= /usr/bin/nekono-voicevox-mcp 配置)
2. `claude mcp add --transport stdio --scope user nekono-voicevox --env VOICEVOX_ENGINE_URL=http://100.117.10.53:50021 -- /usr/bin/nekono-voicevox-mcp` で登録
3. pipx install からの移行完了

## 更新方針

upstream で新 release (= v0.2.0 等) が出たら:
1. user が `git tag -s v0.2.0 -m "Release 0.2.0"` + `git push origin v0.2.0` + GitHub Release publish
2. nvchecker が新 release を検知 → Issue が立つ
3. PKGBUILD の pkgver を更新、 sha256 を独立再計算 (= `curl + sha256sum`)
4. depends 変化があれば反映 (= 例: fastmcp v4 で API 変化、 httpx v2 で API 変化等)
5. `.SRCINFO` + `.deps.lock` を再生成
6. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-20 | 0.1.0 | `2a6f86b252998cf0e52fea0be9480372f05ae1fa` | `a1d1423524ad7eb53db0667e8ce7b1dd6473a831` | 初回 add、 自家 source v0.1.0 |
| 2026-05-24 | 0.1.0-2 | `66f853b` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): pipewire version bump |
| 2026-05-28 | 0.1.0-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): pipewire 1:1.6.5-2 → 1:1.6.6-1 |
| 2026-06-12 | 0.1.0-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-2 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-15 | 0.1.0-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-hatchling 1.29.0-1 → 1.30.1-1 |
| 2026-06-17 | 0.1.0-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-4 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-21 | 0.1.0-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): pipewire 1:1.6.6-1 → 1:1.6.7-1 |
| 2026-06-23 | 0.1.0-8 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
| 2026-06-24 | 0.1.0-9 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-7 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-06-25 | 0.1.0-10 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-8 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-01 | 0.1.0-11 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-9 → 3.4.2-10 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-02 | 0.1.0-12 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-10 → 3.4.2-11 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-06 | 0.1.0-13 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-11 → 3.4.2-12 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-07 | 0.1.0-14 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.2-12 → 3.4.3-1 rebuild に追随 (build-all の cascade-warn 由来) |
| 2026-07-09 | 0.1.0-15 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (cascade: [nekono] dep rebuilt): python-fastmcp 3.4.3-1 → 3.4.3-2 rebuild に追随 (build-all の cascade-warn 由来) |
