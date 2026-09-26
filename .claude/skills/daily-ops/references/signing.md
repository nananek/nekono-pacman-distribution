# 署名 (Nekono GPG / YubiKey) の前提と罠

署名が要る操作: `git commit -S` / `--amend -S`、`bin/build-all` (makepkg `--sign`、`repo-add --sign`)、
`bin/update-repo`、`bin/sign-pkg`。

## 構成 (2026-09 時点の build host)

- 署名鍵 = Nekono GPG、fingerprint `483DC691DF9F29327EA106BD030130E2F156CD74`。**primary を `!` 付きで指定**
  (subkey に流れて "wrong card" になる事故の防止。スクリプトは対応済み)。秘密鍵は disk に無く **YubiKey 上**。
- YubiKey は user の手元 workstation に刺さっている。build host へは SSH の `RemoteForward` で
  **`/run/user/<uid>/gnupg/S.gpg-agent` (full socket)** を転送して使う (`extra-socket` は制限付きで card 操作が通らない)。
  → **user の SSH セッションが生きている間だけ署名できる**。
- 設定は user の `~/.ssh/config` (`RemoteForward ... S.gpg-agent`、`StreamLocalBindUnlink yes`)。
  もしこの設定が無く card がこのマシンに直接刺さっているなら、ローカルの gpg-agent は正常で、
  下記の rogue 判定は当てはまらない。

## readiness check (署名系の操作の前に)

```sh
S=/run/user/$(id -u)/gnupg/S.gpg-agent
[ -S "$S" ] && echo "forward socket: present" || echo "forward socket: MISSING"
ps -eo pid,args | grep -E '[g]pg-agent|[s]cdaemon'
```

| 状態 | 判定 |
|---|---|
| socket あり、`--homedir $HOME/.gnupg` の gpg-agent/scdaemon が**居ない** | OK。`gpg --card-status` で card (OpenPGP) が見えることを確認して進む |
| `--homedir /run/user/<uid>/ccv-gpg-agent-*` の agent | Claude Code 自身の helper。**触らない** |
| socket 無し、または `--homedir $HOME/.gnupg` の agent が居る (= rogue) | **gpg を呼ばない**。下の復旧手順へ |

## 禁止

- `gpgconf --kill all` / `gpgconf --kill gpg-agent` — 「署名が通らないから agent を再起動」は逆効果。
  gpg は生きた agent を socket に見つけられないとローカルに `gpg-agent --homedir ~/.gnupg --daemon` を
  自動起動し、それが**転送 socket と同じ path を奪う**。以後 card 系だけ「そのようなデバイスはありません」で失敗する
  (`gpg-connect-agent 'GETINFO version'` は OK を返すので一見正常に見える = 診断が難しい)。
- forward が落ちている間に `gpg` / `git commit -S` / `bin/build-all` を繰り返さない。1 回ごとに rogue が再生成され得る。
- pcscd / lsusb / udev を疑って時間を使わない (原因は大抵 rogue agent)。

## 復旧手順 (署名が失敗した / card が見えない)

`bin/build-all` は build 自体が成功した後に `署名に失敗しました` (exit 16) で落ちるのが典型
(forward が build 中に落ち、makepkg の署名 step が rogue agent を生む)。

1. `ps` と `ls -la /run/user/$(id -u)/gnupg/S.gpg-agent` だけで状況を見る (**gpg は呼ばない**)。
2. `--homedir $HOME/.gnupg` の gpg-agent / scdaemon の PID を控える。
3. **agent 自身が kill すると権限分類に止められる**ことが多い。retry せず user に頼む:
   `! kill <pid> <pid>`。stale な socket file が残る場合はそれも user に。
4. user が workstation から build host に SSH し直す (forward が再 bind される)。
5. `S.gpg-agent` が再作成され、rogue が居ないことを (1) と同じ方法で確認 → その後に `gpg --card-status`。
6. 中断した処理を再実行。`bin/build-all --pending` は冪等 (未署名の同 version pkg を消して再 build。部分結果は無害)。

## PIN cache

build 全体 (数十分〜) を覆うため `~/.gnupg/gpg-agent.conf` の `default-cache-ttl` は 3600 以上が前提。
`bin/build-all` は冒頭で agent を warm up する。途中で PIN prompt が出る構成ではない (出たら止まる)。
PIN 入力が要る状況になったら user に伝える。
