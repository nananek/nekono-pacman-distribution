# Phase 4: publish (`bin/publish`) と配信検証

`bin/publish` は署名済みの `repo/x86_64/` を**配信ホスト (dufs WebDAV, `:8080`)** へ rclone コンテナで push する。
配信ホストは署名鍵を持たない土管で、client は `SigLevel=Required` で Nekono GPG を検証する
(信頼モデルは不変)。build 成功を `build.md` の手順で検証できた後は、**確認なしで続けて実行してよい**
(user が「アップロードまで自動で」と決めた運用)。

## 前提チェック

1. **docker (rootless)**: `systemctl --user is-active docker.service`。`inactive` なら
   `systemctl --user start docker.service`。`sudo systemctl start docker` / system unit の `is-active docker` は使わない
   (この環境の docker は user-level の rootless。system unit が inactive なのは正常)。
2. **認証情報**: `~/.config/nekono-pacman/publish.env` (repo 外、chmod 600)。3 変数が非空かを**値を出さずに**確認:

   ```sh
   f=~/.config/nekono-pacman/publish.env
   for v in NEKONO_DAV_URL NEKONO_DAV_USER NEKONO_DAV_PASS; do
     grep -Eq "^(export )?$v=.+" "$f" && echo "$v: set" || echo "$v: MISSING"
   done
   ```

   ファイルが無い / 空 / 変数欠落 / docker が起動できない → 黙って skip せず、何が足りないかを user に伝えて停止。
   値は絶対に表示・log しない。

## 実行

```sh
LOG=$(mktemp -t nekono-publish-XXXXXX.log)
( set -a; source ~/.config/nekono-pacman/publish.env; set +a
  bin/publish > "$LOG" 2>&1; rc=$?; echo "publish exit=$rc" >> "$LOG"; exit $rc )
tail -n 15 "$LOG"       # `[OK] publish complete` と `publish exit=0`
```

(pkg が大きいと時間がかかる。長いなら `run_in_background`。`build.md` と同じく rc を捕捉する。)

**`bin/publish` を「1 pass に簡素化」しない**。dufs WebDAV は mtime を保存できず rclone が size-only 比較になる。
`*.sig` は常に 119 byte なので、db や pkg を作り直しても sig が同 size で `sync` に skip され、古い sig が
新しい db の上に残って client の `pacman -Sy` が署名不正で落ちる (2026-07-04 / 2026-08-22 に実発生)。
そのため pass 1 = `sync` (pkg の追加・撤廃の削除)、pass 2 = `copy --ignore-times` (db と sig の強制更新)。

pass 1 の log で `Copied (new)` (今回 build した pkg と `.sig`) と `Deleted` (prune された旧版) が想定どおりか見る:
`grep -E 'Copied \(new\)|Deleted' "$LOG"`。旧版の `Deleted` が出ていれば配信側の掃除も済んでいる。

## 配信検証 — pacman client が実際に取得する物がローカルと一致するか

exit 0 だけでは足りない。**client と同じ匿名 GET 経路** (Caddy `:80`) で取得した物の hash をローカルと比べる。
build host は自身も `[nekono]` を pacman.conf に持つので、その `Server` をそのまま使える (認証情報不要):

```sh
srv=$(grep -A4 '^\[nekono\]' /etc/pacman.conf | awk -F' *= *' '/^Server/{print $2; exit}' | sed 's/\$repo/nekono/; s/\$arch/x86_64/')
for f in nekono.db.tar.gz nekono.db.tar.gz.sig <今回 build した pkg の .pkg.tar.zst と .sig ...>; do
  l=$(sha256sum "repo/x86_64/$f" | cut -c1-16)
  r=$(curl -fsS --max-time 60 "$srv/$f" | sha256sum | cut -c1-16)
  [ "$l" = "$r" ] && echo "MATCH    $f" || echo "MISMATCH $f"
done
```

- `$srv` (host 名を含む) は表示しない。MATCH/MISMATCH だけ出す。
- `nekono.db.tar.gz` と `.sig` の**両方**が MATCH であること (sig の stale が過去の事故)。
- 取得に失敗 (`curl -f` が非 0) しても空 stream の hash になり MISMATCH と出る。
- 撤廃した pkg は、配信側に**もう無い**ことを確認 (build 前に `ls repo/x86_64/<name>-*` で file 名を控えておく):

  ```sh
  curl -s -o /dev/null -w '%{http_code}\n' "$srv/<retired-file>"      # 404 が期待値
  ```

`/etc/pacman.conf` に `[nekono]` が無いマシンでは、代わりに publish.env の認証付き経路 (dufs `:8080`) で同じ照合をする:
`curl -fsS -K <(printf 'user = "%s:%s"\n' "$NEKONO_DAV_USER" "$NEKONO_DAV_PASS") "$NEKONO_DAV_URL/x86_64/$f"`
(password を argv に載せないため `-K` + process substitution。`source` は subshell 内で行う)。

## 「build 済みだが未 publish」かもしれない時 (build が `nothing to build` だった等)

上の検証を先にやる。`nekono.db.tar.gz` と `.sig` が MATCH なら配信は最新で publish 不要。
MISMATCH なら前回の build 後に publish されていないので、そのまま publish する。

## その先 (user に依頼する範囲)

client (ayaka 等) での `sudo pacman -Sy && pacman -Si <pkg>` による新 version の解決確認は sudo を伴うので
user に依頼する (`pacman -Sy` を自分で実行しない)。
