---
name: daily-ops
description: nekono-pacman-distribution の日次運用を最初から最後まで実行する runbook — bot PR / upstream Issue の消化 → merge → bin/build-all (Nekono GPG 署名) → bin/publish → 配信検証。「PR/Issue 消化して」「bot PR 片付けて」「upstream 更新を bump して」「build して publish して」「朝 routine」「配信まで反映して」で使う。repo 初見・記憶なしの agent 向けに前提・罠を全部含む。package の新規追加/撤廃は package-lifecycle skill。
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
| 1 | bot PR `deps/<pkg>-pkgrel-<N>` → 3 点修正 → 署名 amend → merge | `references/bot-pr.md` |
| 2 | upstream Issue → 自分で調査 → bump PR → merge | `references/upstream-issue.md` |
| 3 | `bin/build-all --pending` (署名付き build + repo db 更新) | `references/build.md` |
| 4 | `bin/publish` → 配信側との hash 照合 | `references/publish.md` |
| 5 | 最終報告 (本ファイル) | 下記 |

署名 (YubiKey / gpg-agent) は commit `-S` にも build にも要る。罠が多いので Phase 1 の前に
`references/signing.md` の readiness check を通すこと。

## Phase 0: preflight

```sh
cd "$(git rev-parse --show-toplevel)"
git status --short                    # 空であること
git config core.hooksPath             # `.githooks`。空なら: git config core.hooksPath .githooks
git checkout master && git pull --ff-only && git fetch origin   # pre-push gate は origin/master を base にする。古いと誤 block
gh auth status                        # 未 login なら user に `! gh auth login` を依頼
gh pr list --state open
gh issue list --state open
```

build host 判定 (Phase 3/4 をここで実行できるか):

```sh
[ -f ~/.config/nekono-pacman/publish.env ] && [ -S /run/user/$(id -u)/gnupg/S.gpg-agent ] \
  && [ -f repo/x86_64/nekono.db.tar.gz ] && echo BUILD_HOST || echo NOT_BUILD_HOST
```

`NOT_BUILD_HOST` なら Phase 2 の merge までで止め、user に
`git pull --ff-only && bin/build-all --pending` と `bin/publish` を build host で実行するよう提示する。

## 自律の境界 (最重要)

**確認なしで進めてよい** (CLAUDE.md が事前承認している 2 経路だけ):

1. bot PR (`deps/*-pkgrel-*`) の pkgrel bump — 3 点修正 → 署名 → merge
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
- docker は **rootless**。`systemctl --user ...` で扱う。`sudo systemctl start docker` は使わない。
- `~/.config/nekono-pacman/publish.env` の**値を出力しない** (`cat` / `echo` / `set -x` 禁止)。
- background 実行の通知 "exit 0" を信じない。rc を捕捉し、log 末尾と artifact で確認してから成功と言う (→ `build.md`)。
- コマンド出力・log・commit message の script 由来部分は英語 (ASCII)。コメント・REVIEW.md・PR 本文は日本語 OK。
- commit message / PR 本文末尾には harness 指定の attribution 行を付ける (前例: 578e54a, PR #652)。

## Phase 5: 最終報告

user には表で返す。省略せず、未完・保留を必ず出す:

| 項目 | 内容 |
|---|---|
| merge 済み | PR 番号 / pkg / 版 (bot PR は pkgrel) |
| 保留・停止 | Issue/PR 番号と理由 (block、誤検知疑い、承認待ち) |
| build | 対象 pkg と版、`[cascade-warn]` があればその pkg (pkgrel bump PR が要る) |
| publish | 実行有無、配信側 hash 照合の結果 (MATCH/MISMATCH) |
| user にお願い | client で `sudo pacman -Sy && pacman -Si <pkg>` して新 version が見えるか。ansible-nekonodesk 側の変更があれば |

失敗・スキップした step はそのまま書く (「build は署名失敗で未完、publish 未実行」等)。
