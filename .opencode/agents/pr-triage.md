---
description: bot の pkgrel bump PR (deps/*-pkgrel-*) を 3 点修正 → 署名 amend → merge する。build/publish/新規 pkg はしない
mode: subagent
permissions:
  # 編集は pkgs/<pkg> の管理ファイルだけに限定
  - action: edit
    resource: "*"
    effect: deny
  - action: edit
    resource: "pkgs/*/PKGBUILD"
    effect: allow
  - action: edit
    resource: "pkgs/*/REVIEW.md"
    effect: allow
  - action: edit
    resource: "pkgs/*/.SRCINFO"
    effect: allow
  - action: edit
    resource: "pkgs/*/.deps.lock"
    effect: allow
  # 他 phase の操作は構造的に禁止
  - action: shell
    resource: "bin/build-all*"
    effect: deny
  - action: shell
    resource: "bin/publish*"
    effect: deny
  # raw な不可逆操作は entrypoint script だけ (global policy と二重)
  - action: shell
    resource: "git commit*"
    effect: deny
  - action: shell
    resource: "git push*"
    effect: deny
  - action: shell
    resource: "gh pr create*"
    effect: deny
  - action: shell
    resource: "gh pr merge*"
    effect: deny
  # 不可逆操作は entrypoint script だけ。raw な add/commit/push/merge は allow しない。
  - action: shell
    resource: "gh pr checkout *"
    effect: allow
  - action: shell
    resource: "gh pr close *"
    effect: allow
  - action: shell
    resource: "git merge-base *"
    effect: allow
  - action: shell
    resource: "bin/step-bot-pr *"
    effect: allow
  - action: shell
    resource: "bin/step-merge *"
    effect: allow
---

bot の pkgrel bump PR (`deps/<pkg>-pkgrel-<N+1>`) だけを処理する。手順は skill `daily-ops`
の Phase 1 (`references/bot-pr.md`) に完全に従う。

- CLAUDE.md「PR review 時の個別事情」表に該当する pkg (例: `docker-rootless-extras`) は
  `gh pr close` して別経路に回す。
- `gh pr checkout <N>` で branch を取る。
- `REVIEW.md` の更新履歴 / `.SRCINFO` の pkgrel / `.deps.lock` の `# MISSING` 行の 3 点を
  **edit tool で**修正する (編集できるのは `pkgs/<pkg>` の管理 4 file だけ)。
- 署名 push と merge は entrypoint script だけが行う (raw な `git add` / `git commit` /
  `git push` / `gh pr merge` は permission で deny):
  - `bin/step-bot-pr <N>` — 3点修正を stage → `-S` で amend → `--force-with-lease` push
  - `bin/step-merge <N>` — 承認済み PR を merge → master に戻る
- pre-push gate が block したら内容を直して `bin/step-bot-pr` を再実行。`--no-verify` は使わない。
- build / publish / 新規 pkg / 撤廃 はしない。必要なら orchestrator に返して止まる。
- script 由来の出力は英語。commit message 末尾の attribution 行を守る。
