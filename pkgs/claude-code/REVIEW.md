# claude-code review

## 状態

**review 済み、approve** (最新: 2026-06-06 / 2.1.167)

AUR の `claude-code` PKGBUILD を fork。改変なし。各 release の review 履歴は
本ファイル末尾の「更新履歴」 section 参照。

(AUR には `claude-code` と `claude-code-bin` 両方ある。内容ほぼ同一で
`-bin` は古い version `2.1.126`。両者とも Anthropic 公式 prebuilt binary
を取るだけ、命名の歴史が違うだけ。**無印 `claude-code` の新しい方を採用**。)

## Source

- AUR: https://aur.archlinux.org/packages/claude-code
  - maintainers: Christopher Cooper / Jérôme Poulin / Fabio Fontana
  - automation: https://github.com/fabifont/claude-code-aur
- Upstream binary: `https://downloads.claude.ai/claude-code-releases/<ver>/linux-<arch>/claude`
  - Anthropic 公式 CDN、self-contained Bun executable (JS + resources embed)
- Upstream legal: `https://code.claude.com/docs/en/legal-and-compliance.md`
  - Anthropic 公式 ドメイン、TLS 有効、HTTP 200
- GitHub: https://github.com/anthropics/claude-code (= Anthropic Org、
  124k stars、license field は GitHub 上 null = repo に LICENSE 明示なし、
  PKGBUILD は `LicenseRef-claude-code` で扱う)

## 検証結果

- [x] `source_x86_64` URL = `downloads.claude.ai/claude-code-releases/2.1.143/linux-x64/claude`
  - Anthropic 公式 CDN、典型的な mirror spoof / DNS hijack に脆弱だが TLS で守られる
- [x] `source_aarch64` URL = `downloads.claude.ai/claude-code-releases/2.1.143/linux-arm64/claude`
  - 同上
- [x] `sha256sums_x86_64` が upstream binary と一致
  - 実測 (2.1.143): `f75fdc3ff9d9cd494b86192f9e349b5c5c6d3970ed4d5cd5c7b330c5a2b1dcc4`
  - PKGBUILD 値: 一致 (Issue #23 の事前調査 + PR #32 の claude-review.yml の独立再計算で確認)
- [x] `sha256sums_aarch64` が upstream binary と一致
  - 実測 (2.1.143): `32e8edc4a5c3c41d18607c75d1b8e7bec643330c03e266be46ac3b41a446c4eb`
  - PKGBUILD 値: 一致 (同上)
- [x] `source=("cc-legal::...legal-and-compliance.md")` の sha256 は `SKIP`
  - 法文 markdown は string、binary としては実行されないので SKIP は許容
  - 内容 churn が頻繁な doc に sha pin は意味薄い
- [x] `package()`:
  - `/opt/claude-code/bin/claude` ← upstream binary (= self-contained Bun)
  - `/usr/bin/claude` wrapper:
    ```sh
    #!/bin/sh
    export DISABLE_UPDATES=1
    exec /opt/claude-code/bin/claude "$@"
    ```
    → `DISABLE_UPDATES=1` を強制して self-update を抑止、pacman 管理に統一
  - `/usr/share/licenses/claude-code/LICENSE` ← legal doc
  - install -Dm 標準、curl / wget / eval なし
- [x] `options=('!strip')`: 必須 (= Bun 静的 binary に embedded JS を含む、
      stripping すると壊れる)
- [x] `depends`: `bash` のみ (= wrapper 用)
- [x] `optdepends`: `git / github-cli / glab / ripgrep / tmux / bubblewrap /
      socat` — claude が呼ぶ optional ツール、妥当
- [x] license `LicenseRef-claude-code` — Anthropic 独自 license (= SPDX 標準
      に無いため LicenseRef-)、`cc-legal` を licenses/ に同梱

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key 483D...` 可。

binary は Anthropic 公式 CDN + sha256 pin で守られる。Anthropic の GPG
signing は CLI binary には付かないため (= npm tarball / CDN download とも
署名なし)、tarball の sha256 を毎 release 手動更新する運用が必須。

## 更新方針

upstream の新 release (2.1.143 等) が出たら:
1. AUR で pkgver / sha256sums の値を確認 (`https://aur.archlinux.org/packages/claude-code`)
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立再計算 (= `curl -fsSL https://downloads.claude.ai/claude-code-releases/<ver>/linux-x64/claude | sha256sum`)
4. cc-legal 自体は churn 多いので照合不要、SKIP のまま
5. REVIEW.md の「検証結果」 section を新 sha256 で上書き + 「更新履歴」に
   1 行追記

## 更新履歴

- **2026-06-10 / 2.1.170** — approve。Issue #187 (+ #185) 調査済み (release author `ashwin-ant` = 過去 release と同一、source URL `downloads.claude.ai` 不変)。
  `2.1.168 → 2.1.169 → 2.1.170` の連続リリース (スキップなし)。 v2.1.169 で **セキュリティ修正** (信頼されていない project settings が
  trust 確認なしに OTEL クライアント証明書パスを設定できた脆弱性) + `--safe-mode` フラグ / `/cd` コマンド / `disableBundledSkills` 設定の追加、
  v2.1.170 で Claude Fable 5 (Mythos) 導入 + VS Code 統合ターミナル等から起動時に transcript が保存されず `--resume` に出ないバグ修正。
  build script / depends / package() / wrapper script の変更なし。 breaking change / 削除なし。 install hook / `curl|sh` 等の新規追加なし。
  sha256 は raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `849e007277a0442ab27570d3e3d6d43787507946590e8dd1947e5a39b7081f9e`
  - aarch64: `1bb9d032440a75532f7dd4cafbc687f220aaf16c63eba17e192dfbec2f04bd25`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #187 #185。

- **2026-06-07 / 2.1.168** — approve。Issue #182 調査済み (release author `ashwin-ant` = 過去 release と同一、source URL `downloads.claude.ai` 不変)。
  release notes は "Bug fixes and reliability improvements" の汎用テキストのみ (= 2.1.167 と同文)、 breaking change / 削除 / security fix の言及なし。
  binary サイズが 2.1.167 比で微増 (linux-x64.tar.gz: 74,641,738 → 74,643,179 bytes) のため実コード変更を含む genuine release。
  build script / depends / package() / wrapper script の変更なし。 install hook / `curl|sh` 等の新規追加なし。
  sha256 は build host (nekono-pacman0) で raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `e2f7cb50442bdee21bf2686ef3725a6af187a204e46c4af5c12d0f6d76326485`
  - aarch64: `40d50e7c45742aaa3707fa3628d7f765c55ed503108b6f100513e38d32477aa0`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #182。

- **2026-06-06 / 2.1.167** — approve。Issue #179 調査済み (release author `ashwin-ant` = 過去 release と同一、source URL `downloads.claude.ai` 不変)。
  `2.1.163 → 2.1.165 → 2.1.166 → 2.1.167` の連続リリース (`v2.1.164` は GitHub タグ無し)。
  バグ修正・信頼性向上が中心。 v2.1.166 で **セキュリティ強化** (`SendMessage` 経由でクロスセッション中継された
  メッセージがユーザー権限を失う変更、 受信者は中継 permission を auto モードでもブロック) +
  `fallbackModel` 設定 (最大 3 つ)・deny ルールの glob 対応・`MAX_THINKING_TOKENS=0` で thinking 無効化を追加。
  build script / depends / package() / wrapper script の変更なし。 breaking change / 削除なし。
  sha256 は raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `d6d2995bfca3f8539d9e9aa513ff43c3daa0d556d6d1af07c6df681e050e522c`
  - aarch64: `b8f383df1dca557dc8fb817e4e76335639f94a0a8c7b803ca2f5aef12d373f09`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #179。

- **2026-06-05 / 2.1.163** — approve。Issue #168 調査済み (release author `ashwin-ant` = 過去 release と同一)。
  バグ修正・UX 改善リリース、 breaking changes / security fix なし。 主な変更:
  managed settings に `requiredMinimumVersion` / `requiredMaximumVersion` の version 強制ポリシー追加、
  `/plugin list --enabled` / `--disabled` フィルタ、 `/btw` に "c to copy" ショートカット、
  Stop / SubagentStop hook の `hookSpecificOutput.additionalContext` 拡張、
  Skills の `\$` リテラルエスケープ、 stdio MCP の `--resume` 時 `CLAUDE_CODE_SESSION_ID` 受け渡し、
  `$TMPDIR` オーバーライドのリグレッション修正 (2.1.154)、 Windows read-only / OneDrive ディレクトリの
  Bash "EEXIST" 修正、 org managed permission rules / `$HOME` deny rule / hook `if:` 条件マッチング修正、
  Bedrock/Vertex/Foundry + `CI=true` の `claude -p` 修正、 バックグラウンドシェルハング改善。
  sha256 は raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `5dddcb2c091da60cf9b1bef782e6c78a7fada2f2cd3db4f131c9ebc2478fd447`
  - aarch64: `ca0010a80e3c4749e59c6e8429ec4a4e2ecbaafac36d3535636e04369bbb87c0`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #168。

- **2026-06-04 / 2.1.162** — approve。Issue #161 調査済み (release author `ashwin-ant` = 過去 release と同一)。
  バグ修正・UX 改善リリース、security fix なし。主な変更: `claude agents --json` の `waitingFor`
  フィールド追加、 read-only config dir での起動ハング修正、 WebFetch preapproved ドメインの
  permission rule 修正、 Windows の permission rule バックスラッシュ正規化、 MCP per-server timeout
  下限 1s への切り上げ修正、 LSP workspaceSymbol 修正、 `claude agents` UI 修正多数。
  sha256 は raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `947a49b0de8688f6a74a6e753c24771ff3ddd17b2a6dae85f36304ec514e61d1`
  - aarch64: `eca2a603dfebc3426a8469cbe797f9df95245738bc1c20ec842fc8f80af4010d`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #161。

- **2026-06-03 / 2.1.161** — approve。 Issue #152 調査済み (release author `ashwin-ant` = 過去 release と同一)。
  主な変更: Opus 4.8 サポート、 Dynamic Workflows、 plugin auto-loading、
  **セキュリティ修正複数** (PowerShell `cd` バイパス / managed-settings バイパス / MCP secrets 漏洩 修正)。
  v2.1.147 / v2.1.152 で `/simplify` → `/code-review` リネーム (breaking)。
  sha256 は raw binary を直接 `curl | sha256sum` 実測:
  - x86_64: `1f6a22f387a3bce496b6d869389a35dffb5a69c97d9831833f3bd6dc0e6c6c28`
  - aarch64: `7dfa0a79a2fc9f332057cdc0302f808cba63df7b75e2ccb5a7c1ab62639804e3`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。 Closes #152。

- **2026-05-21 / 2.1.146** — approve。upstream tag commit: Issue #92 調査済み (release author `ashwin-ant`)。
  バグ修正リリース (`/simplify` → `/code-review` リネーム、MCP paginating 修正、GNOME Terminal paste 修正 等)。
  セキュリティ修正なし。sha256 は Issue #92 の供給値 (upstream `SHASUMS256.txt` cross-check 済み):
  - x86_64: `825d5301380f1f5f466c5268de25a062927be658938fc1d630cfa02c521b8185`
  - aarch64: `af25334c7a2632a531b34e3f4c0d69763b997149d31d5f0d748e44813758806f`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。Closes #92。

- **2026-05-21 / 2.1.145** — approve。upstream tag commit: Issue #72 調査済み。
  **セキュリティ修正** (パーミッションプロンプトバイパス脆弱性を修正)。
  release author は `ashwin-ant` (= 2.1.144 と同一)。sha256 は Issue #72 の
  供給値 (= upstream `SHASUMS256.txt` と cross-check 済み):
  - x86_64: `b3ffbc12689bfe81389d6577787fcea4cab81bd3b6bba9b719e73770b62d720e`
  - aarch64: `75ad61d690d79440c82b5841444e1b42caae55736af37c97dd0e068ef20ce390`
  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ。Closes #72。

- **2026-05-19 / 2.1.144** — approve。 upstream tag commit:
  `69d707009ec5a9362ea3552b0580d0f658428f0a`。 主にバグ修正リリース (= breaking
  changes / security fix なし)。 release author は `ashwin-ant` (= 2.1.143 と同一)。
  Issue #38 (upstream-version-issue.yml) で `downloads.claude.ai` の binary を
  GitHub Release の `SHASUMS256.txt` (= 同 author 公開) と cross-check 完全一致確認:
  - x86_64: `147480774472e5720fd5e83617b3e9299344e7213efa84c326b25bd5a0f20b4e`
  - aarch64: `c8ccccbfce12d684588bd3af366394132f614dcf3c86beb2066f86bde2704513`

  PKGBUILD 改変は `pkgver` + 2 sha256 の 3 値のみ、 `package()` / wrapper script
  / depends は無変更。 `/extra-usage` → `/usage-credits` のリネームと
  `--bg` バックグラウンドセッション関連の改善が主。

- **2026-05-17 / 2.1.143** — approve。upstream diff は CHANGELOG.md 追記 1
  commit のみ、build script / depends / install script 変化なし。Issue #23
  (upstream-version-issue.yml の事前調査) + PR #32 (claude-review.yml の独立
  sha256 実測) で両アーキ一致確認。release author は `ashwin-ant` (Anthropic
  社員、過去 release と同経路)。
- **2026-05-15 / 2.1.142** — approve (初回 fork + review)。当時の sha256
  実測値: `1249a1dadfe2d48f320bd4e1b657a1a0d82435da76deb11ce509822407cf24ec`
  (x86_64) / `767b13fc28763ca9d663b00f90e501f134b356f1b72dcf0eea59b7e3bed86411`
  (aarch64)。
