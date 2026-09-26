# Phase 1: bot PR (`dep-version-pr`) の処理

## 何の PR か

`.github/workflows/dep-version-pr.yml` が毎日 (JST 11:55) 立てる**機械 PR**。

- branch `deps/<pkg>-pkgrel-<N>` / title `<pkg>: pkgrel bump to <N> (deps changed)` / author `github-actions[bot]`
- 中身は `PKGBUILD` の `pkgrel` +1 と `.deps.lock` の更新だけ (Arch 公式 repo の依存 version が動いた →
  ABI 追従の rebuild 指示)。本文の「Dep changes」に変わった依存が列挙される。
- commit は **unsigned** (Actions に署名鍵を置かない方針)。merge 前に build host で署名し直す。

## 列挙

```sh
gh pr list --state open --json number,headRefName,title \
  --jq '.[] | select(.headRefName|startswith("deps/")) | "\(.number)\t\(.headRefName)\t\(.title)"'
```

`deps/` 以外の open PR (人間/他 agent の PR) は**この phase の対象外**。触らず最終報告に載せる。

## Step 1: 個別事情テーブルを確認

`CLAUDE.md` の「PR review 時の個別事情」表を読む。表にある pkg (現状 `docker-rootless-extras`) は
pkgver が depends と sync しているので pkgrel bump は不正解。

```sh
gh pr close <N> --comment "pkgver が <dep> と sync する pkg のため pkgrel bump は不適切。upstream-version-issue 経路で pkgver を上げる (CLAUDE.md 個別事情表)"
```

閉じたら次の PR へ。対応する upstream Issue がまだ無ければ最終報告で user に伝える
(`workflow_dispatch` で `upstream-version-issue.yml` を手動起動する手もある)。

同じ pkg に upstream bump (Phase 2) が同時に絡むなら、bump 側で `.deps.lock` も最新化して含め、
bot PR は close (理由をコメント) するのが素直。迷ったら user に確認。

## Step 2: 3 点修正 (repo 規約。bot は下の 3 つを満たさない)

```sh
gh pr checkout <N>
```

### (a) `pkgs/<pkg>/REVIEW.md` の「更新履歴」に 1 行追記

```sh
grep -n '更新履歴' -A8 pkgs/<pkg>/REVIEW.md      # 既存の形式と並び順を確認してから書く
```

表形式なら 5 列。前例 (electron37-bin):

```
| 2026-09-27 | 37.10.3-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): nss 3.129-1 → 3.130-1 |
```

日付は `date +%F`、release 列は `<pkgver>-<新 pkgrel>`、findings は PR 本文の「Dep changes」を転記。

**並び順は既存の表に合わせる**。表は **新しい行が下** が大多数 (2026-09 時点で 15 pkg 中 13。sunshine-nekono /
nekono-btkeycast だけ新しい行が上)。bullet 形式の REVIEW.md (claude-code) は新しい方が上。先頭に足すか
末尾に足すかを決め打ちせず、既存の日付の並びを見て同じ向きに足す。

### (b) `.SRCINFO` の `pkgrel` を PKGBUILD に揃える

```sh
sed -i -E 's/^([[:space:]]*pkgrel = ).*/\1<N>/' pkgs/<pkg>/.SRCINFO
( cd pkgs/<pkg> && makepkg --printsrcinfo | diff - .SRCINFO && echo SRCINFO-in-sync )
```

`printsrcinfo` との diff が空ならOK。空でなければ差分の中身を見て判断 (並び順だけの cosmetic 差は
pkg によって既存)。

### (c) `.deps.lock` の `# MISSING ...` 行を復元

bot は header + 更新後の `name=version` だけで lock を書き直すので、AUR-only / virtual provide / nekono repo 内 pkg を
示す `# MISSING ...` 行が**消える**。bot PR の**分岐元**の lock をベースに、bot が変えた version 行だけ当て直すのが確実
(`origin/master` ではなく merge-base を使う。bot が branch を切った後に master が進んでいても、bot の差分だけを取れる):

```sh
pkg=<pkg>; base=$(git merge-base origin/master HEAD)
diff <(git show $base:pkgs/$pkg/.deps.lock | grep -E '^[^#[:space:]]') \
     <(grep -E '^[^#[:space:]]' pkgs/$pkg/.deps.lock) | grep '^>' | sed 's/^> //' > /tmp/changed.$$
git checkout $base -- pkgs/$pkg/.deps.lock
while IFS== read -r n v; do sed -i "s|^${n}=.*|${n}=${v}|" pkgs/$pkg/.deps.lock; done < /tmp/changed.$$
rm -f /tmp/changed.$$
git diff $base -- pkgs/$pkg/.deps.lock      # 変わった version 行だけで、MISSING 行が残っていること
```

## Step 3: 署名して push、merge

```sh
git add pkgs/<pkg>/REVIEW.md pkgs/<pkg>/.SRCINFO pkgs/<pkg>/.deps.lock
git commit --amend -S --no-edit
git push --force-with-lease            # pre-push gate (bin/prepush-review) が走る。pkgrel bump は素通りが正常
git log -1 --format='%G? %an | %s'     # %G? が N (未署名) でないこと。U/G は署名済み (U = trust 未設定表示で正常)
gh pr merge <N> --merge --delete-branch
git checkout master && git pull --ff-only
```

待つべき CI は無い。`--no-edit` (message は bot のまま。author が bot のまま署名だけ付く) は既存の運用。
gate が BLOCK したら内容を読んで直す (pkgrel bump では通常起きない)。`--no-verify` は使わない。

## 全 PR 処理後の sanity check (任意だが推奨)

`.SRCINFO` が PKGBUILD と食い違った pkg が無いか (bot PR の 3 点修正漏れの検出):

```sh
for d in pkgs/*/; do ( cd "$d" && makepkg --printsrcinfo 2>/dev/null | diff -q - .SRCINFO >/dev/null || echo "SRCINFO drift: $d" ); done
```

`okaguchi-lawyer-dicterm` は makedepends の並び順だけの既知の cosmetic 差。それ以外が出たら user に報告
(自分の担当外の pkg なら勝手に直さず、報告 + 修正 PR の提案)。
