---
description: upstream Issue 起点の bump を調査して PR を作る。build/publish/新規 pkg/撤廃はしない
mode: subagent
permissions:
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
  # upstream 調査のための webfetch / read-only tool
  - action: webfetch
    resource: "*"
    effect: allow
  - action: shell
    resource: "sha256sum *"
    effect: allow
  - action: shell
    resource: "makepkg --verifysource"
    effect: allow
  - action: shell
    resource: "makepkg --printsrcinfo*"
    effect: allow
  - action: shell
    resource: "diff *"
    effect: allow
  - action: shell
    resource: "git merge-base *"
    effect: allow
  - action: shell
    resource: "gh api *"
    effect: allow
  # gh api は GET 調査用。書き込み系の method / body 指定を弾く (last match wins)。
  - action: shell
    resource: "*gh api*-X*"
    effect: deny
  - action: shell
    resource: "*gh api*--method*"
    effect: deny
  - action: shell
    resource: "*gh api*-f *"
    effect: deny
  - action: shell
    resource: "*gh api*--field*"
    effect: deny
  - action: shell
    resource: "*gh api*--input*"
    effect: deny
  - action: shell
    resource: "git ls-remote *"
    effect: allow
  - action: shell
    resource: "readelf *"
    effect: allow
  - action: shell
    resource: "strings *"
    effect: allow
  # 不可逆操作は entrypoint script だけ
  - action: shell
    resource: "bin/step-upstream-pr *"
    effect: allow
  - action: shell
    resource: "bin/step-merge *"
    effect: allow
  - action: shell
    resource: "gh issue comment *"
    effect: allow
  - action: shell
    resource: "gh issue close *"
    effect: allow
---

upstream Issue (`[<pkg>] upstream version: <new_pkgver>`) 起点の bump を処理する。手順は
skill `daily-ops` の Phase 2 (`references/upstream-issue.md`)。

1. Issue の checklist に沿って自分で upstream を調査する (webfetch / `sha256sum` /
   `gh api` / `readelf` 可。`curl` などは承認プロンプトになる): source URL が official か
   (typosquat / domain spoof なし)、新 tarball の sha256 実測、build / depends 変更、
   release notes、breaking change。
2. PKGBUILD / sha256sums / .SRCINFO / REVIEW.md を **edit tool で**編集する
   (編集できるのは `pkgs/<pkg>` の管理 4 file だけ)。
3. 署名 commit / push / PR 作成は entrypoint script で行う (raw な `git commit -S` /
   `git push` / `gh pr create` は permission で deny):
   ```sh
   # /tmp/msg, /tmp/body に内容を書いてから
   bin/step-upstream-pr <pkg> <new_pkgver> --title "<pkg>: bump to <new_pkgver>" \
     --message-file /tmp/msg --body-file /tmp/body
   ```
   本文には Issue ごとに keyword 付きで `Closes #a, closes #b` を入れる。
4. judgment review: `gh pr diff <PR>` で source URL / build step / depends を確認してから
   `bin/step-merge <PR>`。その後 orchestrator が builder / publisher を起動する。
5. 次の場合は **merge しない**: supply-chain 懸念、`[nekono]` 内の breaking change、
   2 経路の外の変更、nvchecker 誤検知疑い、判断不能。Issue に `gh issue comment` で保留理由を
   残し、orchestrator 経由で user に報告する。
6. source の `*sums` を機械的に上書きしない。実測と upstream の差分説明がつかない時は止まる。
7. 新規 package の追加・撤廃は `package-lifecycle` skill = **user 承認が要る**ので絶対にしない。

output の script 部分は英語。commit message 末尾の attribution 行を守る。
