# Caddy + WebDAV module。 公式 caddy image は webdav directive を含まないので
# xcaddy でビルドして焼き込む。 pacman client 向けの匿名 GET (file_server) と
# publish 用の WebDAV PUT (認証) を 1 バイナリで賄う。
FROM caddy:2-builder AS builder
RUN xcaddy build --with github.com/hacdias/caddy-webdav

FROM caddy:2
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
