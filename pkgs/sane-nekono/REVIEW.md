# sane-nekono review log

## 2026-08-06

- **source**: AUR ではなく自前 fork <https://github.com/nananek/sane-backends>
  tag `1.4.0-nekono1`。tag は Nekono GPG
  (`483DC691DF9F29327EA106BD030130E2F156CD74`) で署名済み。
  upstream `gitlab.com/sane-project/backends` の 1.4.0 からの差分は
  **2 ファイル 5 行**のみ (`git diff 1.4.0..1.4.0-nekono1 -- include backend`)。
- **PKGBUILD**: Arch 公式 `extra/sane` (1.4.0-4) をベースに、
  `source[0]` を上記 fork に差し替え + `provides`/`conflicts` を追加しただけ。
  build/package の処理は公式と同一 (パス変数の使い分けのみ変更)。
- **補助ファイル 4 つ** (`66-saned.rules` `sane.sysusers` `saned.service`
  `saned.socket`) は Arch packaging repo の原本を無改変で使用。checksum も公式値。
- **gcc16.patch** は upstream commit `d04d17b` をそのまま参照 (公式と同一)。
- findings: **受入**。動的取得なし、SKIP checksum なし、
  build/package 内に curl/wget/eval 無し。
- 差分の中身: ① `include/sane/sane.h` の `#if 0` → `#if 1`
  (`SANE_FRAME_JPEG` の封印解除) ② `backend/fujitsu.c` の
  `usleep(500000)` → `usleep(20000)` (リトライ上限 120 → 3000 でタイムアウト総量は不変)。
  どちらも外部と通信せず、他 backend の動作を変えない。

## 依存方針

`extra/sane` からの意図的 diff は `source[0]` の fork 差し替えと
`provides`/`conflicts` の追加のみ。depends/makedepends は公式と完全一致。
