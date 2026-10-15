---
name: daily-ops
description: nekono-pacman-distribution の日次運用を最初から最後まで実行する runbook — ローカル検知〜生成 (bin/check-upstream / check-deps) → bot PR / upstream Issue の消化 → merge → bin/build-all (Nekono GPG 署名) → bin/publish → 配信検証。「PR/Issue 消化して」「bot PR 片付けて」「upstream 更新を bump して」「build して publish して」「朝 routine」「配信まで反映して」で使う。repo 初見・記憶なしの agent 向けに前提・罠を全部含む。package の新規追加/撤廃は package-lifecycle skill。
---

# daily-ops — Issue/PR 消化 → build → publish

自作 pacman repo `[nekono]` の日次運用 runbook。`CLAUDE.md` の「Claude への委任フロー」を、
実際に手を動かす順に具体化したもの。規約の *理由* は CLAUDE.md、ここは *手順*。

> CLAUDE.md の「build host 作業 (Claude には実行できない)」は **古い**。build host は 2026-07-14 に
> nekono-pacman0 から「Claude Code が動いているこのマシン」へ移った。build host 上なら
> `bin/build-all` / `bin/publish` は Claude が直接実行する (下記 preflight で判定)。

## 使い分け

| やりたいこと | 使うもの |
|---|---|
| Issue/PR 消化、bump、build、publish | この skill |
| build だけ / publish だけ | Phase 3 / 4 の references だけ読めばよい |
| package の新規追加・撤廃・復活 | `package-lifecycle` skill (user 承認が要る) |

## Phase 一覧

| Phase | 内容 | 詳細 |
|---|---|---|
| 0 | preflight (本ファイル) | 下記 |
| 0.5 | 今日分の検知〜生成 (`bin/check-upstream` / `bin/check-deps`) | 下記 |
| 1 | bot PR `deps/<pkg>-pkgrel-<N>` → 検証 (不足あれば 3 点修正 + 署名) → merge | `references/bot-pr.md` |
| 2 | upstream Issue → 自分で調査 → bump PR → merge | `references/upstream-issue.md` |
| 3 | `bin/build-all --pending` (署名付き build + repo db 更新) | `references/build.md` |
| 4 | `bin/publish` → 配信側との hash 照合 | `references/publish.md` |
| 5 | 最終報告 (本ファイル) | 下記 |

署名 (YubiKey / gpg-agent) は commit `-S` にも build にも要る。罠が多いので Phase 0.5 / Phase 1 の
前に `references/signing.md` の readiness check を通すこと (check-deps は署名 commit を作る)。

## Phase 0: preflight

`bin/preflight` を実行する（read-only。branch/file は変更しない）。working tree /
`core.hooksPath` / `gh auth` / open bot PR / open upstream Issue / build host /
forward socket の状態が 1 回で出る:

```sh
bin/preflight
```

- `core.hooksPath` が `.githooks` でなければ pre-push gate が無効。orchestrator の shell は
  entrypoint 3 本だけなので、user に `git config core.hooksPath .githooks` を依頼する。
- build host: NO なら Phase 2 の merge までで止め、user に
  `git pull --ff-only && bin/build-all --pending` と `bin/publish` を build host で実行するよう
  提示する。

## Phase 0.5: 今日分の検知〜生成 (旧 GitHub Actions 相当)

cron で回っていた 2 本の workflow (`upstream-version-issue` / `dep-version-pr`) は、
GitHub の schedule 遅延 (実測 +2〜7h、中央値 +5h) で daily trigger の体を成さないため
**ローカル実行に移した** (workflow は手動 `workflow_dispatch` の fallback として残る)。

```sh
bin/check-upstream --dry-run    # nvchecker の結果だけ見る (GitHub を変更しない)
bin/check-upstream              # upstream 更新 → Issue を立てる
bin/check-deps --dry-run        # 差分の一覧だけ見る
bin/check-deps                  # 差分 → pkgrel bump の署名済み完成 PR を立てる
```

- どちらも**冪等**: 既存 Issue (title fingerprint) / 既存 branch・PR は skip される。
- 生成物は Phase 1 (deps PR) / Phase 2 (upstream Issue) がそのまま消化する。
- 前提ツールは preflight が warning で出す (`nvchecker` / `fakeroot` / `gh`)。`bin/check-deps`
  は fresh な Arch 公式 repo db を throwaway で同期して比較する (root 不要)。
- ここで失敗が出ても、取れた分の Issue / PR は有効。失敗は最終報告に載せる。

## 自律の境界 (最重要)

**確認なしで進めてよい** (CLAUDE.md が事前承認している範囲だけ):

0. 今日分の検知〜生成 (`bin/check-upstream` / `bin/check-deps`、Phase 0.5) — 冪等・機械的
1. bot PR (`deps/*-pkgrel-*`) の pkgrel bump — 検証 (不足あれば 3 点修正 + 署名) → merge
2. upstream Issue 起点の bump — 調査結果が safe-to-bump / needs-attention (既存 pkg を今まで通り build 可能に
   保つ範囲: 版・sha256・URL の更新、depends 追加、patch の追従、build 手順の微修正) のもの
3. 1・2 の結果を build → publish (build 成功を artifact で確認できた後は publish も確認なし。user 承認済みの運用)

**止まって user に報告する** (PR は出してよいが `gh pr merge` しない / Issue は open のまま
`gh issue comment` で保留理由を残す):

- supply-chain 上の懸念: source URL の domain/owner 変更、upstream checksum と実測の不一致 (tarball 差し替え疑い)、
  新規 install script / postinstall / 動的 fetch、AUR-only 依存の新規要求、release の取り下げ・maintainer 交代直後の異常な差分
- [nekono] 内の他 pkg と互換が壊れる breaking change
- 上記 2 経路の**外**の変更: 新規 package の追加、撤廃、`bin/` `.github/` `.githooks/` の変更、source origin の変更、
  install hook の新設、機能の削除/大幅変更。自作 PR は gate を通っても・review しても自己承認しない
  (gate は機械チェック。過去に自作の新規 pkg PR を自分で merge して権限分類に止められた)
- nvchecker の誤検知が疑われる Issue (close する前に内容を示して確認)
- 自分では判断がつかないもの

## 守ること

- **全 commit `-S`**。`git commit --no-verify` / `git push --no-verify` は使わない (user が明示した時だけ)。
- **`git add -A` / `git add .` 禁止**。`pkgs/<pkg>/` に makepkg の download cache が落ちる。path を明示して add。
- 1 PKGBUILD update = 1 commit。複数 package を 1 commit に混ぜない。
- **`gpgconf --kill` 系を絶対に実行しない**。転送された YubiKey socket を奪う rogue agent を生む (→ `signing.md`)。
- `bin/publish` は host の `rclone` を直接使う (docker 不使用)。
- `~/.config/nekono-pacman/publish.env` の**値を出力しない** (`cat` / `echo` / `set -x` 禁止)。
- background 実行の通知 "exit 0" を信じない。rc を捕捉し、log 末尾と artifact で確認してから成功と言う (→ `build.md`)。
- コマンド出力・log・commit message の script 由来部分は英語 (ASCII)。コメント・REVIEW.md・PR 本文は日本語 OK。
- **非対話 run (ask ゼロ) 前提**: `question` は使わない (orchestrator は permission deny 済み)。
  止まる時は Phase 5 報告に理由を書いて終了する。`ask` になる操作が出たら harness の
  allow 漏れなので、止めて報告に残す (`--no-verify` 等で迂回しない)。
- commit message / PR 本文末尾には harness 指定の attribution 行
  `Assisted-by: OpenCode (DeepSeek V4.1 Flash)` を付ける（モデルが違えばその名前で）。

## Phase 5: 最終報告

user には表で返す。省略せず、未完・保留を必ず出す:

| 項目 | 内容 |
|---|---|
| 検知〜生成 (Phase 0.5) | 立った Issue / deps PR の番号 (dry-run 時はその内容) |
| merge 済み | PR 番号 / pkg / 版 (bot PR は pkgrel) |
| 保留・停止 | Issue/PR 番号と理由 (block、誤検知疑い、承認待ち) |
| build | 対象 pkg と版、`[cascade-warn]` があればその pkg (pkgrel bump PR が要る) |
| publish | 実行有無、配信側 hash 照合の結果 (MATCH/MISMATCH) |
| user にお願い | client で `sudo pacman -Sy && pacman -Si <pkg>` して新 version が見えるか。ansible-nekonodesk 側の変更があれば |

失敗・スキップした step はそのまま書く (「build は署名失敗で未完、publish 未実行」等)。
