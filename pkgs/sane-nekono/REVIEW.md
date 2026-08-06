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

## 2026-08-06 (build fix)

- **findings**: 1.4.0-1 (build fix)。`bin/build-all` (glibc 2.44 / gcc 16.1.1
  ホスト) で build 失敗: `configure` の `AC_CHECK_TYPES([u_char, u_short,
  u_int, u_long])` が `_DEFAULT_SOURCE`/`_GNU_SOURCE` 無し環境で実行されるため
  誤検知 (= 型が既にシステムヘッダにあるのに「無い」と判定)、結果
  `include/sane/config.h` が `#define u_char unsigned char` 等の互換 macro を
  発行 → 実ビルド時 (`_GNU_SOURCE` 有効下で glibc が本物の `typedef` を提供) に
  `frontend/scanimage.c` で二重定義衝突しコンパイルエラー。
- **原因は upstream sane-backends 1.4.0 側の configure スクリプトの不備**
  (新しめの glibc で顕在化)。nekono fork の 5 行差分 (JPEG 封印解除 /
  polling interval) とは無関係、supply-chain 上の懸念なし。
- **対処**: `build()` の `./configure` 呼び出し時のみ
  `CPPFLAGS="${CPPFLAGS} -D_DEFAULT_SOURCE"` を付与し、configure の型検出を
  実ビルド時の条件と一致させる (= config.h が誤った互換 macro を発行しなく
  なる)。build host でクリーンビルドし直し、成功確認済み
  (`libsane-fujitsu.so` に `JPEG` 文字列あり、`fujitsu.c` に
  `usleep(20000)` が 2 箇所、上記パッチ効果も再確認済み)。
- 外部通信 / SKIP checksum / curl・wget・eval 追加は無し。`depends` /
  `makedepends` / `source[]` / checksum は無変更 (`build()` の 1 行追加のみ)。

## 依存方針

`extra/sane` からの意図的 diff は `source[0]` の fork 差し替えと
`provides`/`conflicts` の追加のみ。depends/makedepends は公式と完全一致。
