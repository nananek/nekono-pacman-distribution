# phase: triage

bot の pkgrel bump PR (`deps/<pkg>-pkgrel-<N>`) を **1 本だけ**処理する phase。
対象 PR 番号は user メッセージに書かれている。手順は skill `daily-ops` の Phase 1
(`references/bot-pr.md`)。

## 前提

通常の PR は `bin/check-deps` が作った **署名済み・3 点反映済みの完成 PR**。
judgment review をして merge するだけでよい。手動 `workflow_dispatch` 由来の
unsigned PR だけ 3 点を直して `step_bot_pr` で署名し直す。

## 手順

1. `gh_pr` method=view で PR を読む（base が `master`、head が `deps/`、state open）。
2. `pkg_3points` で 3 点を確認する（`.SRCINFO` 同期 / `.deps.lock` の `# MISSING` 行 /
   REVIEW.md が新 pkgrel に触れているか）。
3. `CLAUDE.md` の「PR review 時の個別事情」表を `fs_read` で確認する。表にある pkg
   （例: `docker-rootless-extras` は pkgver が moby と sync する）は **pkgrel bump が
   不正解**。`gh_pr` method=close で理由をコメントして閉じ、その旨を findings に書く
   （対応する upstream Issue が無ければ「check-upstream の再実行が必要」と書く）。
4. 3 点が欠けている場合だけ直す:
   - `gh_pr` method=checkout で branch を取る
   - `pkg_srcinfo` の出力を `.SRCINFO` に `fs_write` で反映
   - REVIEW.md の「更新履歴」に 1 行（日付 / release / `(this PR)` / dep の変化）。
     **並び順は既存の表に合わせる**（新しい行が下の pkg が多数。先頭に足す pkg もある）
   - `.deps.lock` の `# MISSING` 行を `repo_merge_base` + `repo_diff` で確認して復元
   - `step_bot_pr` で stage → `-S` amend → force-push
5. 3 点が揃っている（または直した）なら `gh_pr` method=diff を読んで judgment review:
   source URL が upstream official か / `build()` `package()` に動的 curl が無いか /
   depends が妥当か。問題なければ `step_merge` で merge。
6. `finish(status="ok")` に PR 番号・pkg・pkgrel・merge の有無・findings を書く。

## しないこと

- `pkg_3points` が ok でないまま merge しない。
- build / publish はこの phase ではしない（driver が後で回す）。
- 自分で判断できない（供給網が怪しい、個別事情表に無いが pkgrel bump が不適切に見える）
  なら `escalate(kind="content")` で人間に渡す。
