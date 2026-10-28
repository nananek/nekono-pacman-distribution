# 署名 (remote-gpg-agent / rga) の前提と罠

署名が要る操作: `git commit -S` / `--amend -S`、`bin/build-all` (makepkg `--sign`、`repo-add --sign`)、
`bin/update-repo`、`bin/sign-pkg`。

## 構成 (2026-10 以降)

- 署名鍵 = remote-gpg-agent の鍵、fingerprint `2641FA2BFF801BBA75E8633AEF7566A07D865C0F`。
  秘密鍵は **iPhone の PWA の中だけ**にあり、サーバは passkey で暗号化した envelope を預かるだけ。
- build VM の `/usr/local/bin/gpg` はシム。`--detach-sign` だけを SSH 越しに host teyvat の
  `rga gpg-compat` に流し、それ以外 (`--verify` など) は VM の本物の gpg に委譲する。
- host の `rga` が mTLS (`https://gpg.tail2c8c7.ts.net:8443`) で署名要求を送り、
  iPhone の PWA が Face ID で承認する。**秘密鍵もクライアント証明書も VM には置かない。**
- VM は tailnet に出ない。署名は host の SSH bridge key (`~/.ssh/host_bridge`) 経由。
- `--local-user` の末尾 `!` は gpg の primary key 固定指示。rga が fingerprint 解決前に落とす。

## readiness check (署名系の操作の前に)

VM の中:

```sh
ssh -o BatchMode=yes -i ~/.ssh/host_bridge kts_sz@10.0.2.2 '$HOME/.local/bin/rga ping'
```

host の上:

```sh
rga ping
rga keys
```

| 状態 | 判定 |
|---|---|
| `rga ping` が `{"mode":"client","ok":true}` | OK。`rga keys` に `personal` が見えることを確認 |
| SSH が `Permission denied` / `host key` | VM の `~/.ssh/host_bridge` と host の authorized_keys を確認 |
| `rga ping` が TLS/証明書エラー | host の `~/.config/remote-gpg-agent/` (ca/client cert/key) と有効期限を確認 |
| `rga keys` が空 | PWA で鍵を作っていない。PWA の設定を確認 |
| 承認待ちがタイムアウト | iPhone の PWA を開く。通知が来ない時は Web Push の設定 (無効→有効→テスト) を確認 |

## 禁止

- host の `~/.config/remote-gpg-agent/` を VM にコピーしない (秘密鍵と証明書は host に閉じる)。
- VM から `gpg.tail2c8c7.ts.net:8443` へ直接接続しない (VM は tailnet にいない)。
- PWA を閉じたまま `bin/build-all` を回さない。署名ごとに承認待ち (最大 180 秒) で止まる。

## 復旧手順 (署名が失敗した)

`bin/build-all` は build 自体が成功した後に署名で落ちるのが典型。

1. host で `rga ping` / `rga keys` を見る。
2. VM で `ssh -i ~/.ssh/host_bridge kts_sz@10.0.2.2 '$HOME/.local/bin/rga ping'` を見る。
3. host の `~/.config/remote-gpg-agent/config.json` の `url` が `:8443` 付きか確認する。
4. PWA を開き、承認待ちが残っていれば署名/拒否する。
5. 中断した処理を再実行。`bin/build-all --pending` は冪等 (未署名の同 version pkg を消して再 build)。

## 承認の運用

- 1 package = 1 承認。`bin/build-all --pending` で複数 build するときは、PWA を開いたまま
  承認を続ける。通知は補助で、来なくても PWA の承認画面に出る。
- 署名要求の `description` は作業ディレクトリ (例: `/home/kts_sz/repos/nekono-pacman-distribution`)。
- 署名が拒否された場合、rga は exit 1 で止まる。build 済みの pkg は `repo/x86_64/` に残るので、
  原因を直して `bin/build-all --pending` を再実行する。
