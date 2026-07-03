# nekono 配信ホスト (Caddy + WebDAV, docker-compose)

ビルドホスト (nekono-pacman0) から**配信を切り離す**ための scaffold。
配信ホスト (例: `nekono-dist0`) 上で docker-compose を 1 本上げるだけで、

- **GET (匿名)** … pacman client 向けの静的配信 (`file_server`)
- **PUT (認証)** … build host からの publish 用 WebDAV

を 1 コンテナ (Caddy + `hacdias/caddy-webdav`) で賄う。ホスト側に nginx 設定も
ssh+rsync も要らない。**build 道具・Nekono 署名鍵はこのホストに置かない**
(署名は artifact に付随し、client が `SigLevel=Required` で検証するため、
配信ホストは土管でよい)。

> ⚠ これは未検証の scaffold。実ホストに上げる初回に、Caddy の webdav module
> ビルド・directive・rclone WebDAV 疎通を必ず検証すること。

## セットアップ (配信ホスト)

```sh
# 1) tailnet に join 済みの docker ホストで:
cd deploy/dist
cp .env.example .env
# DAV_USER を決め、bcrypt hash を生成して .env の DAV_HASH に貼る:
docker run --rm caddy:2 caddy hash-password --plaintext 'YOUR_STRONG_PASS'

# 2) 起動 (webdav module を xcaddy でビルドして焼く)
docker compose up -d --build

# 3) 匿名 GET 疎通 (まだ空でも 200/404 が返れば OK)
curl -sI http://<dist-host>.<tailnet>.ts.net/x86_64/nekono.db
```

本番は `compose.yaml` の `ports` を Tailscale IP に bind する
(`"100.x.y.z:80:80"`)。0.0.0.0 に晒さない。

## publish (build host 側)

build host で `bin/build-all` 後に `bin/publish` を叩く。ホストに rclone を
入れず、rclone コンテナで WebDAV へ sync する。creds は環境変数で注入:

```sh
export NEKONO_DAV_URL=http://nekono-dist0.<tailnet>.ts.net
export NEKONO_DAV_USER=nekono-publish
export NEKONO_DAV_PASS='YOUR_STRONG_PASS'   # .env の DAV_HASH の元の平文
bin/publish
```

`bin/publish` は `rclone sync --copy-links --delay-updates` で:

- `--copy-links` … `nekono.db` (symlink → `nekono.db.tar.gz`) を実体として上げる
  (WebDAV/S3 は symlink を持てないため)
- `--delay-updates` … 全 pkg を上げてから db を最後に rename。同期途中に
  「db が参照する pkg がまだ無い」状態を client に見せない

## client 側 (ayaka 等)

`/etc/pacman.conf` の `Server` を配信ホストに向けるだけ (SigLevel は不変):

```ini
[nekono]
SigLevel = Required DatabaseRequired
Server  = http://nekono-dist0.<tailnet>.ts.net/$arch
```

## 分担 (ansible-nekonodesk 側の follow-up)

- 配信ホストのプロビジョニング (docker / tailscale / この compose のデプロイ) →
  ansible-nekonodesk の新規 role
- client の `Server` URL 差し替え → ansible-nekonodesk の `pacman_nekono_repo` role
