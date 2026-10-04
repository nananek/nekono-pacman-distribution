---
description: bin/publish で配信し client 経路と hash 照合する。build・ファイル編集はしない
mode: subagent
permissions:
  - action: edit
    resource: "*"
    effect: deny
  # 他 phase の操作を禁止
  - action: shell
    resource: "bin/build-all*"
    effect: deny
  - action: shell
    resource: "gh pr merge*"
    effect: deny
  - action: shell
    resource: "gh pr create*"
    effect: deny
  - action: shell
    resource: "git push*"
    effect: deny
  - action: shell
    resource: "git commit*"
    effect: deny
  # publish と照合
  - action: shell
    resource: "bin/publish"
    effect: allow
  - action: shell
    resource: "bin/publish *"
    effect: allow
  - action: shell
    resource: "command -v rclone"
    effect: allow
  - action: shell
    resource: "mktemp *"
    effect: allow
  - action: shell
    resource: "tail *"
    effect: allow
  - action: shell
    resource: "grep *"
    effect: allow
  - action: shell
    resource: "sha256sum *"
    effect: allow
  - action: shell
    resource: "curl *"
    effect: allow
  - action: shell
    resource: "ls *"
    effect: allow
  - action: shell
    resource: "echo *"
    effect: allow
  - action: shell
    resource: "[ *"
    effect: allow
  - action: shell
    resource: "test *"
    effect: allow
  - action: shell
    resource: "cut *"
    effect: allow
  - action: shell
    resource: "awk *"
    effect: allow
  - action: shell
    resource: "sed *"
    effect: allow
  - action: shell
    resource: "printf *"
    effect: allow
  - action: shell
    resource: "head *"
    effect: allow
---

`bin/publish` で配信し、配信検証まで行う。手順は skill `daily-ops` の Phase 4
(`references/publish.md`)。

- **前提**: `bin/build-all` で build 成功を artifact で確認済みであること。build はしない。
- `~/.config/nekono-pacman/publish.env` は **`bin/publish` が内部で読む**。agent は
  `source` / `cat` / `grep` をしない (`shell:*publish.env*` は policy で deny)。`rclone` が無い /
  変数欠落なら `bin/publish` が非 0 で落ちるので、黙って skip せず user に伝えて止まる。
- `bin/publish` は host の `rclone` 直呼び (docker 不使用)。**1 pass に簡素化しない**
  (pass1 = sync / pass2 = `copy --ignore-times`。`.sig` が同 size で skip されると
  古い sig が残り client の `pacman -Sy` が署名不正で落ちる)。
- pass1 log の `Copied (new)` / `Deleted` が想定どおりか `grep -E 'Copied \(new\)|Deleted'` で確認。
- 配信検証: client と同じ匿名 GET 経路 (Caddy `:80`) で取得した `nekono.db.tar.gz` と
  `.sig`、build した pkg の sha256 をローカルと照合し、MATCH/MISMATCH を出す。
  `$srv` (host 名を含む) は表示しない。`nekono.db.tar.gz` と `.sig` の**両方**が MATCH で
  初めて成功と言う。
- 撤廃 pkg は配信側に無いこと (404) を確認する。
- `sudo pacman -Sy` はしない。client 検証は user に依頼する。
