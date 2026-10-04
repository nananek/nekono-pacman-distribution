# nekono-pacman-distribution — agent instructions (OpenCode)

自作 pacman repo `[nekono]`。AUR 由来の package を Claude 廃止後の OpenCode で
review + Nekono GPG 署名して Tailscale 配信する。

> **作業前に必ず読むもの**
> - 規約の正: `CLAUDE.md`（OpenCode は自動で読まないので明示的に読む。命名 / review 観点 /
>   PKGBUILD の落とし穴 #1–7 / commit policy / 退役手順はここが正）
> - 手順の正: skill `daily-ops`（日次: PR/Issue 消化 → build → publish → 配信検証）
>   と `package-lifecycle`（pkg の追加・撤廃・復活。**user 承認が要る**）
> このファイルは **harness（step ごとの隔離）** の説明と、絶対ルールの要約。

## この repo の harness（step ごとの隔離）

既定 agent は `orchestrator`。**orchestrator は編集も shell もできない**（read-only の調査と
subagent 起動だけ）。実際の作業は phase ごとの subagent だけが行う。

| phase | agent | できること |
|---|---|---|
| 進行・preflight | `orchestrator` | `bin/preflight`（read-only）と subagent 起動のみ |
| bot pkgrel PR 消化 | `pr-triage` | `pkgs/<pkg>` の編集、`bin/step-bot-pr` / `bin/step-merge` |
| upstream Issue 起点 bump | `upstream-bump` | `pkgs/<pkg>` の編集 + webfetch、`bin/step-upstream-pr` / `bin/step-merge`。build/publish は不可 |
| 判断系 review | `reviewer` | 読み取りだけ（編集・commit・push・build・publish 不可） |
| build | `builder` | `bin/build-all` と検証のみ（PR 操作・publish 不可・ファイル編集不可） |
| publish | `publisher` | `bin/publish` と配信照合のみ（build 不可・ファイル編集不可） |

強制の層:

1. **tool 層** — `opencode.jsonc` の `permissions`（既定 `ask` + 明示 `allow`）と
   `experimental.policies`（hard deny）。agent 固有の制約は `.opencode/agents/*.md`。
2. **command 層** — 不可逆な操作は `bin/step-*` の単一 entrypoint だけが実行する。
   raw な `git commit` / `git push` / `gh pr create` / `gh pr merge` は daily-ops の
   phase agent（pr-triage / upstream-bump / builder / publisher）の shell deny で塞ぐ
   （pattern は best-effort。build agent は通常コーディングのため対象外。最終防衛は
   GitHub branch protection）。script が対象 file と branch を検証する。
   - `bin/preflight` — Phase 0 の状態表示（read-only）
   - `bin/step-bot-pr <PR>` — 3点修正を stage → `-S` amend → `--force-with-lease` push
   - `bin/step-upstream-pr <pkg> <ver>` — 署名 branch/commit 作成 → push → PR 作成
   - `bin/step-merge <PR>` — 承認済み PR を merge（`deps/*-pkgrel-*` / `pkg/*` のみ）
   - build / publish は既存の `bin/build-all` / `bin/publish`
3. **step 層** — `.opencode/agents/*.md`。agent は「その phase の操作だけ」allow する。
4. **server 層** — `.githooks/pre-push` + **GitHub branch protection**。pre-push hook は
   `git push --no-verify` で抜けられるので、`--no-verify` は policy でも deny し、
   最終防衛は branch protection（`master` 直 push 不可 / PR 必須 / force push 不可）に置く。

> ここは project config。`experimental.policies` の hard deny をより強く・確実にしたい場合は
> user global `~/.config/opencode/opencode.json` に移す（global は repo から緩められない）。
> ただし shell のマッチはコマンド文字列に対する best-effort。sandbox ではない。

## 自律の境界（`daily-ops` skill の要約。判断に迷ったら skill 側を見る）

**確認なしで進めてよい**（CLAUDE.md が事前承認している 2 経路だけ）:

1. bot PR (`deps/<pkg>-pkgrel-<N>`) の pkgrel bump — 3 点修正 → 署名 → merge
2. upstream Issue 起点の bump（safe-to-bump / needs-attention の範囲: 版・sha256・URL 更新、
   depends 追加、patch 追従、build 手順の微修正）
3. 1・2 の build → publish（build 成功を artifact で確認できた後は publish も確認なし）

**止まって user に報告する**（PR は出してよいが merge しない / Issue は open のまま
`gh issue comment` で保留理由を残す）:

- supply-chain 懸念: source URL の domain/owner 変更、checksum 不一致（tarball 差し替え疑い）、
  新規 install script / postinstall / 動的 fetch、AUR-only 依存の新規要求、異常差分
- `[nekono]` 内の他 pkg と互換が壊れる breaking change
- 上記 2 経路の**外**の変更: 新規 package 追加、撤廃、`bin/` `.github/` `.githooks/` の変更、
  source origin 変更、install hook 新設、機能の削除/大幅変更。**自作 PR を自分で merge しない**
- nvchecker の誤検知が疑われる Issue（close 前に内容を示して確認）
- 判断がつかないもの

## 絶対ルール

- commit は必ず **`-S`**（Nekono GPG）。`--no-verify` は使わない。
- **`git add -A` / `git add .` 禁止**。path を明示（`pkgs/<pkg>/` に download cache が落ちる）。
- **1 PKGBUILD update = 1 commit**。複数 package を 1 commit に混ぜない。
- **`gpgconf --kill` 系を実行しない**（転送された YubiKey socket を奪う）。
- `~/.config/nekono-pacman/publish.env` の**値を出力しない**。`bin/publish` が内部で
  読み込むので agent は `source` / `cat` / `grep` しない（`shell:*publish.env*` は
  policy で deny 済み）。
- `sudo pacman -Sy` を自分で実行しない（client 検証は user に依頼）。
- commit message / PR 本文末尾に attribution 行
  `Assisted-by: OpenCode (DeepSeek V4.1 Flash)`（モデル名は実際のもの）を付ける。
- script / log / commit message の script 由来部分は英語 (ASCII)。コメント・REVIEW.md・
  PR 本文は日本語でよい。

## 変更フロー

- `master` から branch を切る（`master` 直 push は hook が拒否）。前例: 追加 `pkg/<name>-<ver>`、
  撤廃 `retire/<name>`、bot 由来 `deps/<pkg>-pkgrel-<N>`。
- 1 pkg = 1 commit。`-S` 署名して push。push 時に pre-push gate (`bin/prepush-review`) が走る。
- 手順・落とし穴は skill、規約の理由は `CLAUDE.md` を参照。

## 通常のコーディングをしたい時

`default_agent` は `orchestrator` なので、素の編集をしたい場合は `opencode --agent build`
で起動するか、TUI で `/agent` から `build` に切り替える。
