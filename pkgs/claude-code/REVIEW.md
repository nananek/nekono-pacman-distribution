# claude-code review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `claude-code` PKGBUILD (pkgver=2.1.142, pkgrel=1) を fork。改変なし。

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

- [x] `source_x86_64` URL = `downloads.claude.ai/claude-code-releases/2.1.142/linux-x64/claude`
  - Anthropic 公式 CDN、典型的な mirror spoof / DNS hijack に脆弱だが TLS で守られる
- [x] `sha256sums_x86_64` が upstream binary と一致
  - 実測: `1249a1dadfe2d48f320bd4e1b657a1a0d82435da76deb11ce509822407cf24ec`
  - PKGBUILD 値: `1249a1dadfe2d48f320bd4e1b657a1a0d82435da76deb11ce509822407cf24ec`
  - 一致
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
5. REVIEW.md に確認日 + 結論 update
