---
name: package-lifecycle
description: nekono-pacman-distribution の package を新規追加する手順 (AUR から fork → security review → PKGBUILD / .SRCINFO / REVIEW.md / .deps.lock / nvchecker.toml / README 整備 → PR) と、撤廃 (retire) して repo db・配信・client から外す手順。撤廃済み pkg の復活も含む。「package 追加して」「AUR から fork して」「○○ を [nekono] に入れたい」「○○ を外したい/撤廃/retire」で使う。日次の bump・bot PR・build・publish は daily-ops skill。
---

# package-lifecycle — package の追加 / 撤廃 / 復活

`[nekono]` に載せる package 集合そのものを変える作業。日次の bump (`daily-ops`) と違い、
**自律 merge の事前承認の対象外**。

## 承認境界 (最重要)

- **どの package を追加/撤廃するかは user が決める**。user の依頼なしに追加・撤廃しない。
  (「この pkg は公式に昇格したので撤廃できそう」等の提案は報告に書くだけ。)
- PR は作ってよいが、**review が通っても pre-push gate が通っても、user の明示承認を得るまで `gh pr merge` しない**。
  CLAUDE.md が事前承認している自律 merge は「bot の pkgrel bump PR」と「upstream Issue 起点の bump」だけで、
  自作の新規 pkg / 撤廃はそこに含まれない。gate は機械チェック (SKIP checksum / build 内 curl 等 / secret /
  REVIEW.md 未更新) にすぎず、判断系の review の代わりにならない。自分で書いた PR を自分で承認しない
  (過去に自作の新規 pkg PR を自分で merge して権限分類に止められた)。
- 承認後の merge → build → publish は `daily-ops` の Phase 3 / 4 (`../daily-ops/references/build.md`, `publish.md`)。

## どの手順か

| 依頼 | 手順 |
|---|---|
| 新規追加 (AUR 由来、または自家 pkg) | `references/add-package.md` |
| 撤廃 (retire) | `references/retire-package.md` |
| 撤廃済みの復活 | `references/add-package.md` の末尾「復活」 |

## 共通ルール

- 全 commit **`-S`** (Nekono GPG)。署名の罠は `../daily-ops/references/signing.md`。`--no-verify` は使わない。
- **1 package = 1 commit**。複数 package を 1 commit に混ぜない。依存 chain を撤廃する等で 1 PR に複数 commit を
  束ねるのは可 (過去に複数 pkg を 1 commit に束ねた撤廃があるが、新規作業は CLAUDE.md の規約に従う)。
- `git add` は **path を明示**。`git add -A` / `.` 禁止 (`pkgs/<name>/` に source の download cache が落ちる)。
- branch は `master` から切る (master への直 push は pre-push hook が拒否)。名前の前例:
  追加 `pkg/<name>-<pkgver>`、撤廃 `retire/<name>`。
- 規約の詳細 (命名、AUR fork の作法、review 項目、落とし穴 #1–7) は `CLAUDE.md`。ここは実行順。
- ansible-nekonodesk (client の `pacman -S` 対象リスト) は**別 repo で、この checkout には無い**。
  変更が要る場合は最終報告で user に具体的に依頼する (勝手に探して触らない)。CLAUDE.md 内で file 名の表記が
  揺れている (`roles/os_packages/vars/Archlinux.yml` / `apt_packages/vars/Archlinux.yml`) ので、実際のパスは
  ansible 側で確認するよう添える。
