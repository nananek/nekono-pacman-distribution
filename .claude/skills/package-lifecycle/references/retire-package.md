# package の撤廃 (retire)

`pkgs/<name>/` を消すだけでは **db (`nekono.db.tar.gz`) と `repo/x86_64/` の `.pkg.tar.zst`、配信側のファイルが残る**。
`bin/build-all` が retire を検出して `repo-remove --sign` + `bin/prune` を自動でやり、`bin/publish` の pass 1
(`sync`) が配信側の削除を行う。手順はその 3 段。

**user が名指しで撤廃を依頼した pkg だけ**。理由の例: Arch 公式に昇格した (`LC_ALL=C pacman -Si <name>` の
Repository が core/extra)、upstream が終了した、不要になった。

## Step 0: 影響調査 (消す前に)

```sh
# [nekono] 内の他 pkg がこれに依存していないか (depends / makedepends / optdepends)
grep -lE '^[[:space:]]*(depends|makedepends|optdepends) = <name>([<>=: ]|$)' pkgs/*/.SRCINFO
# provides / replaces / conflicts で名前が絡んでいないか
grep -nE '^(provides|replaces|conflicts)' pkgs/*/PKGBUILD | grep -w '<name>'
ls -la pkgs/<name>/                     # 消す前に中身を見る (tracked 以外に何があるか)
ls repo/x86_64/<name>-* 2>/dev/null     # 後で「配信側から消えた」確認に使う file 名を控える
```

- **他の [nekono] pkg が依存しているなら、そのまま消さない**。依存元も同時に撤廃するか、依存を外す改修が要るので
  user に確認する (前例: python chain の撤廃、`icecast` → `libigloo`)。
- 公式に昇格した pkg: client の pacman.conf で `[nekono]` と公式のどちらが優先されるか、撤廃後に version が
  下がる (downgrade) 可能性がないかを user に伝える。
- CLAUDE.md の「個別事情」表に該当行があれば、それも消す対象。

## Step 1: repo 側 (branch `retire/<name>`)

```sh
git checkout -b retire/<name> origin/master
git rm -r pkgs/<name>/
rm -rf pkgs/<name>       # git rm は untracked / ignored file (src/, download cache, *.pkg.tar.zst) を消さない。
                         # 直前に ls で中身を確認済みのこと。build-all は PKGBUILD 無し dir を WARN で skip するだけだが残さない
```

- `README.md` の「配布パッケージ一覧」から該当行を削除。
- `nvchecker.toml` の `[<name>]` section を、直前のコメントブロックごと削除。
- `.gitignore` に `pkgs/<name>/...` 個別 pattern があれば削除。
- CLAUDE.md「個別事情」表の該当行 (あれば)。

```sh
git add -u pkgs/<name> README.md nvchecker.toml [.gitignore CLAUDE.md]
git status --short
git commit -S -m "<name>: retire (<reason>)"
```

依存 chain を撤廃するなら **依存元 → 依存先の順**に 1 pkg 1 commit で並べ、1 PR にまとめてよい。

```sh
git push -u origin HEAD
gh pr create --base master --title "<name>: retire (<reason>)" --body "<撤廃理由 / Step 0 の調査結果 (reverse-deps 無し等) / client への影響 / merge 後に build host で build-all → publish が必要>"
```

PR を出したら **user に報告して承認を待つ** (自分で merge しない)。

## Step 2: 承認 → merge → build host で反映

```sh
gh pr merge <N> --merge --delete-branch && git checkout master && git pull --ff-only
bin/build-all --pending --dry-run
```

dry-run に `[*] Will retire N package(s) (no PKGBUILD in pkgs/): <name>` が出ること (出なければ db に元々無い =
既に反映済み)。実行は `daily-ops` の `references/build.md` (background + rc 捕捉) に従う:

```sh
( bin/build-all --pending > "$LOG" 2>&1; rc=$?; echo "build-all exit=$rc" >> "$LOG"; exit $rc )
```

retire は `--pending` や個別指定と独立に**常に**走り、`repo-remove --sign` で db から外して `bin/prune` で
`.pkg.tar.zst{,.sig}` を消す (署名が要るので `signing.md` の前提が必要)。

検証:

```sh
tar -xzO -f repo/x86_64/nekono.db.tar.gz --wildcards '*/desc' | awk '/^%NAME%$/{getline n; print n}' | grep -qx '<name>' \
  && echo "STILL IN DB" || echo "removed from db"
ls repo/x86_64/<name>-* 2>/dev/null && echo "FILES REMAIN" || echo "files pruned"
LC_ALL=C gpg --verify repo/x86_64/nekono.db.tar.gz.sig repo/x86_64/nekono.db.tar.gz 2>&1 | grep -E 'Good signature|BAD'
```

## Step 3: publish

`daily-ops` の `references/publish.md`。pass 1 の `sync` が配信側の該当ファイルを削除し、pass 2 が db と sig を更新する。
検証では **Step 0 で控えた file 名が配信側で 404 になること**も確認する (`publish.md` の「撤廃した pkg」)。

## Step 4: user に依頼 (最終報告に書く)

- client (ayaka 等): まだ install されていれば `sudo pacman -Rns <name>`、その後 `sudo pacman -Sy`
  (撤廃しても install 済みの pkg は自動では消えず、repo に無い "foreign" になるだけ)。
  公式に昇格した pkg なら、次回の更新で公式版に切り替わる。
- ansible-nekonodesk: "Self-hosted (nekono repo)" の対象リストから該当行を消す PR
  (別 repo。この checkout には無い。実際の file パスは ansible 側で確認)。次回の ansible run で client から消える。
