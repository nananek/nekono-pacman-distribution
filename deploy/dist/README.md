# nekono 配信ホスト (Tailscale + Caddy + dufs, docker-compose)

ビルドホスト (nekono-pacman0) から**配信を切り離す**ための compose 一式。
配信ホスト (例: `nekono-dist0`) で `docker compose up -d` すると、tailnet に
tailscale sidecar で参加した上で 2 経路を出す:

- **GET (匿名)** … pacman client 向け静的配信。**Caddy**, `:80`, volume は
  **read-only** mount。
- **PUT (認証)** … build host からの publish 用 **WebDAV**。**dufs**, `:8080`,
  volume は read-write mount。

read 側 (Caddy) を `:ro` mount にすることで「配信は書けない土管」という
security 境界を作る。**build 道具・Nekono 署名鍵はこのホストに置かない**
(署名は artifact に付随し、client が `SigLevel=Required` で検証するため)。

> 検証済み (2026-07-03): `:80` 匿名 GET / `:8080` 認証 GET とも HTTP 200。

## なぜ 2 サーバか

- pacman は匿名 HTTP GET しかできない → GET は無認証で出す (Caddy `:80`)。
- publish は書き込みが要る → dufs の WebDAV に認証付きで PUT (`:8080`)。
- dufs は `--auth user:pass@/:rw` の `:rw` **だけでは書けない**。global の
  `--allow-all` (または `--allow-upload`/`--allow-delete`) を併記して初めて
  upload/delete が有効になる (dufs 仕様: auth 権限は global permission に制限
  される)。匿名 rule を置かないので `:8080` は authed のみ (匿名 GET も 401)。

## セットアップ (配信ホスト)

```sh
cd deploy/dist
cp .env.example .env
# .env を編集:
#   TS_AUTHKEY … tailnet の auth key (初回 join 用)
#   DUFS_USER / DUFS_PASS … publish 認証 (dufs は平文。 bcrypt 不要)

# ./volumes/repodata を dufs の uid(65532) が書けるようにしておく
mkdir -p volumes/repodata volumes/ts_state
sudo chown -R 65532:65532 volumes/repodata

docker compose up -d

# 疎通確認
curl -sI http://nekono-dist0.<tailnet>.ts.net/x86_64/nekono.db          # :80 匿名 GET → 200
curl -sI -u "$DUFS_USER:$DUFS_PASS" http://nekono-dist0.<tailnet>.ts.net:8080/   # :8080 → 200
```

## publish (build host 側)

build host で `bin/build-all` 後に `bin/publish`。ホストに rclone を入れず、
rclone コンテナで dufs (`:8080`) の WebDAV へ sync する (docker が要る):

```sh
export NEKONO_DAV_URL=http://nekono-dist0.<tailnet>.ts.net:8080   # dufs の port
export NEKONO_DAV_USER=nekono-publish        # = DUFS_USER
export NEKONO_DAV_PASS='...'                 # = DUFS_PASS (平文)
bin/publish
```

`bin/publish` は `rclone sync --copy-links --delay-updates` で:

- `--copy-links` … `nekono.db` (symlink → `nekono.db.tar.gz`) を実体として上げる
  (WebDAV は symlink を持てないため)
- `--delay-updates` … 全 pkg を上げてから db を最後に rename。同期途中に
  「db が参照する pkg がまだ無い」状態を client に見せない (dufs は MOVE 対応)

初回だけ、ビルドホストの `repo/x86_64/` をそのまま volume に seed しておくと速い
(このリポジトリの publish 経路が固まる前の bootstrap 用)。

## client 側 (ayaka 等)

pacman の GET は Caddy (`:80`) 側。`Server` を配信ホストに向けるだけ
(SigLevel は不変):

```ini
[nekono]
SigLevel = Required DatabaseRequired
Server  = http://nekono-dist0.<tailnet>.ts.net/$arch
```

## 分担 (ansible-nekonodesk 側の follow-up)

- 配信ホストのプロビジョニング (docker / この compose のデプロイ / TS_AUTHKEY 投入) →
  ansible-nekonodesk の新規 role
- client の `Server` URL 差し替え → ansible-nekonodesk の `pacman_nekono_repo` role
