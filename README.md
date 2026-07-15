# nekono-pacman-distribution

自作 pacman repository。AUR で配布されているけど自分で signing chain を
握りたい (= 公式 repo + 自分の GPG trust chain だけで完結したい) パッケージ
を、AUR PKGBUILD を fork + Claude review + Nekono GPG で署名 + Tailscale
経由で配布する。

このリポジトリは **PKGBUILD と build 用スクリプトのソース** のみ。実際の
署名済み `.pkg.tar.zst` と repo db は build host 上の `repo/` (gitignore)
に置かれ、Tailscale 越しの HTTP で配信される。

## 信頼モデル

- AUR の PKGBUILD を **そのまま信頼しない**。fork して
  `/security-review` skill 等で確認した上で commit する。
- 各 PKGBUILD は対応する `REVIEW.md` を持ち、何を見たか・何を改変したかを
  記録する。
- ビルドした成果物は **Nekono GPG**
  (`483DC691DF9F29327EA106BD030130E2F156CD74`) で署名する。Build host に
  NFC リーダー + YubiKey が刺さってる前提で、master 鍵で signing する
  (= disk 上に秘密鍵を置かない)。

## Repo 構造

```
.
├── README.md                 — この文書
├── CLAUDE.md                 — 規約 (review 手順 / 命名 / commit policy)
├── nvchecker.toml            — upstream version 監視設定 (cron で daily 実行)
├── pkgs/
│   ├── <name>/
│   │   ├── PKGBUILD          — AUR fork + 改変版
│   │   ├── .SRCINFO          — makepkg --printsrcinfo 出力 (AUR diff 用)
│   │   ├── REVIEW.md         — review log (誰が・いつ・何を)
│   │   ├── .deps.lock        — depends/makedepends の version snapshot
│   │   └── *.install         — (任意) post-install hook
│   └── ...
├── bin/
│   ├── build-all             — pkgs/* を makepkg + sign + repo-add で一括処理
│   ├── sign-pkg              — 個別 .pkg.tar.zst を gpg --detach-sign
│   ├── update-repo           — repo-add wrapper
│   ├── prune                 — db 未登録の古い .pkg.tar.zst を削除
│   ├── serve                 — build host 上で nginx が repo/ を Tailscale 配信
│   └── publish               — repo/ を別配信ホストの WebDAV へ container rclone で push
├── deploy/
│   └── dist/                 — 配信を別ホストへ分離する docker-compose (Caddy + WebDAV)
└── repo/                     — gitignore (built + signed artifacts)
    └── x86_64/
        ├── nekono.db.tar.gz
        ├── nekono.db.tar.gz.sig
        └── *.pkg.tar.zst{,.sig}
```

## 配布パッケージ一覧

### デスクトップアプリ

| package | source 方針 | 役割 |
|---|---|---|
| `moonlight-qt`   | from-source (Qt6/FFmpeg/SDL2) | NVIDIA GameStream / Sunshine client |
| `sunshine-bin`   | -bin (LizardByte 公式 pkg)    | Sunshine 配信ホスト (本機を Moonlight でリモート操作) |
| `nekono-btkeycast` | from-source (Python)        | キーボードを BLE HOGP で iPad へ転送 |
| `qview`          | from-source (Qt6)             | 画像ビューア |
| `localsend-bin`  | -bin (Flutter prebuilt)       | AirDrop 代替 (LAN ファイル転送) |
| `pwvucontrol`    | from-source (Rust/GTK4)       | PipeWire 音量コントロール |

### システム・インフラ

| package | source 方針 | 役割 |
|---|---|---|
| `claude-code`            | -bin (npm tarball)      | Anthropic Claude Code CLI |
| `docker-rootless-extras` | from-source             | Docker rootless mode 補助バイナリ |
| `nekonopaw`              | from-source (Go)        | PipeWire → PAM 連携デーモン |
| `pam_pkcs11`             | from-source (C)         | PIV/PKCS#11 PAM 認証モジュール |
| `pass-secret-service`    | from-source (Rust)      | pass を D-Bus secrets backend に |

### 音声・メディア

| package | source 方針 | 役割 |
|---|---|---|
| `electron37-bin`          | -bin (prebuilt zip)      | voicevox-bin 用 electron37 (Arch 公式 extra が rolling で落としたため自家配布) |
| `icecast`                 | from-source (C)          | HTTP audio streaming server |
| `libigloo`                | from-source (C)          | Icecast 2.5.x 必須共通フレームワーク |
| `shairport-sync-airplay2` | from-source (C)          | AirPlay 2 受信 (nqptp 連携) |
| `voicevox-bin`            | -bin (Electron AppImage) | VOICEVOX エディタ GUI |

### MCP サーバー

| package | source 方針 | 役割 |
|---|---|---|
| `nekono-pipewire-mcp`  | from-source (Go)     | PipeWire 操作 MCP server |
| `nekono-voicevox-mcp`  | from-source (Python) | VOICEVOX 音声合成 MCP server |

### MCP サーバー Python 依存

`nekono-voicevox-mcp` / `nekono-pipewire-mcp` の依存 chain。Arch 公式・AUR 未収録のため自家管理。

`python-fastmcp` / `python-mcp` / `python-cyclopts` / `python-beartype` /
`python-rich-rst` / `python-docstring-parser` / `python-sse-starlette` /
`python-openapi-pydantic` / `python-jsonschema-path` /
`python-jsonref` / `python-pdm-pep517` / `python-py-key-value-aio` /
`python-uv-dynamic-versioning` / `python-uncalled-for`

### 辞書

| package | source 方針 | 役割 |
|---|---|---|
| `genshin-dict`            | from-source        | 原神固有名詞辞書 |
| `debuyoko-seiyuu-dicterm` | scrape (WordPress) | でぶよこ声優辞書 |
| `okaguchi-lawyer-dicterm` | scrape (WordPress) | 岡口法律辞書 |

## Build host セットアップ (要件)

新規 Arch マシン 1 台で完結。必要なもの:

- `base-devel git pacman-contrib nginx` (+ from-source build dep の
  qt6-base / gtk4 / ffmpeg / sdl2 / nodejs / npm 等)
- NFC リーダー (例: ACS ACR1252) + YubiKey
- pcscd + gpg-agent + scdaemon を立てて Nekono GPG が `gpg --card-status`
  で見える状態
- Tailscale で `<host>.<tailnet>.ts.net` が解決できる
- `~/.gnupg/gpg-agent.conf` で `default-cache-ttl 3600` 程度に伸ばす
  (build 中 PIN cache が切れないように)

ホスト構築自体は ansible-nekonodesk の Arch 経路 (= `site.yml` + 専用
inventory entry) で管理する想定。

## クライアント側設定

Arch クライアント (ayaka 等) で `/etc/pacman.conf` に追加:

```
[nekono]
SigLevel = Required DatabaseRequired
Server  = http://<build-host>.<tailnet>.ts.net/$arch
```

`pacman-key` で Nekono pubkey を local trust:

```sh
sudo pacman-key --add /etc/pacman.d/nekono.gpg
sudo pacman-key --lsign-key 483DC691DF9F29327EA106BD030130E2F156CD74
```

これらは ansible-nekonodesk の `pacman_nekono_repo` role
(`apt_packages` の前段) で自動配備する想定。

## ビルド手順 (build host で実行)

```sh
git pull                    # 最新の PKGBUILD を取得
bin/build-all               # makepkg + gpg sign + repo-add
sudo systemctl reload nginx # (任意) cache 飛ばす場合
```

build-all は対話プロンプトを起こさない (gpg PIN は事前 cache 必須):

```sh
gpg-connect-agent "SCD SERIALNO" /bye   # YubiKey を warm up
bin/build-all
```

## 配信の分離 (別ホスト, docker-compose) — optional

既定ではビルドと配信を同一ホスト (nekono-pacman0) で行い、`bin/serve` の
nginx が Tailscale 越しに `repo/` を配信する。配信を**別ホストに切り離す**
compose 一式を `deploy/dist/` に用意している (nekono-dist0 でデプロイ済み)。

- 配信ホストで docker-compose を上げると、tailscale sidecar で tailnet に
  参加した上で 2 経路を出す:
  - **匿名 GET** (pacman) = **Caddy** `:80`、volume は read-only mount
  - **認証 PUT** (publish) = **dufs** WebDAV `:8080`、volume は read-write mount
  - read 側の `:ro` mount が「配信は書けない土管」の security 境界。
- build host は `bin/build-all` 後に **`bin/publish`** を叩く。ホストに
  rclone を入れず、rclone コンテナで署名済み `repo/x86_64/` を dufs (`:8080`)
  の WebDAV へ sync する (`--copy-links` で `nekono.db` symlink を実体化、
  `--delay-updates` で pkg→db の順序を保証)。
- **信頼モデルは不変**: `SigLevel=Required` で client が Nekono GPG を検証する
  ため、配信ホストは署名鍵を持たない土管でよい。
- client は `Server` を配信ホスト (Caddy `:80`) に向け替えるだけ。

手順は `deploy/dist/README.md` 参照。配信ホストのプロビジョニングと client の
`Server` URL 差し替えは ansible-nekonodesk 側で行う。

## License

本リポジトリは [**0BSD (Zero-clause BSD)**](https://opensource.org/license/0bsd)
で配布する (= 全文は `LICENSE`)。 attribution / 帰属表記不要、 商用利用 OK、
改変・再配布自由、 無保証。

選定理由は **AUR メンテナとの摩擦回避**。 `pkgs/<name>/PKGBUILD` は AUR
fork に Nekono review + 改変を加えたものなので、 AUR メンテナがこの
リポジトリの改変を upstream の AUR PKGBUILD に取り込む (= cherry-pick) 経路
を license 観点で妨げない / 阻害しないことを最優先にした:

- MIT / BSD-3-Clause 等 attribution 要件付きの license だと、 AUR メンテナが
  back-port 時に license header 等の遵守義務を負う形になる (= AUR 慣行と衝突)
- 0BSD は attribution すら要らないので、 AUR メンテナは何も気にせず diff を
  cherry-pick できる
- AUR PKGBUILD 自体は慣行的に license を明示しない (= do-what-you-want 扱い)
  ので、 0BSD はこの de facto に最も近い formal license

`pkgs/<name>/` 配下の PKGBUILD で扱う **配布対象の application 本体** は、
それぞれの PKGBUILD の `license=()` 行に記載された upstream license に従う
(= 本リポジトリの 0BSD はあくまで Nekono オリジナル部分 = `bin/`, `.github/`,
review/build 仕組み + 各 PKGBUILD への Nekono 改変分にかかる)。

**AUR メンテナへ**: cherry-pick 大歓迎、 何の連絡も不要。 そのまま AUR に
持って帰ってください。
