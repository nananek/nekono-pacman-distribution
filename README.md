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
├── pkgs/
│   ├── <name>/
│   │   ├── PKGBUILD          — AUR fork + 改変版
│   │   ├── REVIEW.md         — review log (誰が・いつ・何を)
│   │   └── *.install         — (任意) post-install hook
│   └── ...
├── bin/
│   ├── build-all             — pkgs/* を makepkg + sign + repo-add で一括処理
│   ├── sign-pkg              — 個別 .pkg.tar.zst を gpg --detach-sign
│   ├── update-repo           — repo-add wrapper
│   └── serve                 — nginx / caddy docker-compose 起動
└── repo/                     — gitignore (built + signed artifacts)
    └── x86_64/
        ├── nekono.db.tar.gz
        ├── nekono.db.tar.gz.sig
        └── *.pkg.tar.zst{,.sig}
```

## 配布パッケージ一覧

| package | source 方針 | 役割 |
|---|---|---|
| `vesktop-bin`    | -bin (upstream AppImage) | Discord client (Vencord patched) |
| `moonlight-qt`   | from-source (Qt6/FFmpeg/SDL2) | NVIDIA GameStream / Sunshine client |
| `qview`          | from-source (Qt6 small)  | 画像ビューア |
| `pika-backup`    | from-source (GTK4/Rust)  | borg backup GUI frontend |
| `localsend-bin`  | -bin (Flutter prebuilt)  | AirDrop 代替 (iOS / Android / Linux LAN ファイル転送) |
| `claude-code`    | -bin (npm tarball)       | Anthropic Claude Code CLI (Arch 公式 repo 無し) |

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
