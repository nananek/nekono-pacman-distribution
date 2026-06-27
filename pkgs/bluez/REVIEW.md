# bluez review

## 状態

**review 済み、approve** (2026-06-12、official-shadow 初版 5.86-6.1)

**本 repo 初の official-shadow パッケージ**。AUR fork ではなく Arch 公式
extra/bluez 5.86-6 を同 pkgname のまま fork し、upstream の未リリース修正
1 本だけを乗せて `pkgrel=6.1` で配布する。クライアントは pacman.conf の
repo 並び順 (`[nekono]` が `[extra]` より前) で本 pkg を優先取得する。

shairport-sync-airplay2 で採った「別 pkgname + provides/conflicts/replaces」
方式とは意図的に異なる: bluez は 8 split packages (bluez / bluez-utils /
bluez-libs / bluez-cups / bluez-deprecated-tools / bluez-hid2hci / bluez-mesh /
bluez-obex) が相互に version 整合を要求するため、公式と同 pkgname / 同 split
構成を維持して公式版との混在・齟齬を防ぐ (Issue #208 の指示)。**暫定配布**
であり、撤去条件 (後述) を満たし次第 retire して公式に戻す。

## 用途

ayaka (linux 7.0.11.arch1-1 + bluez 5.86-6) で `RegisterAdvertisement` が
全クライアント・全アダプタで `org.bluez.Error.Failed` になる LE 広告
リグレッションの修正。ayaka で開発中の btkeycast (BLE HOGP キーボード転送)
がブロックされている。

根本原因 (Issue #208 の btmon トレース調査で確定済み):

- bluetoothd が MGMT `Add Extended Advertising Data (0x0055)` のヘッダ長を
  誤った構造体 (`struct mgmt_cp_add_advertising` = 11B、正しくは
  `mgmt_cp_add_ext_adv_data` = 3B) で計算し、毎回 8 バイト余分に送る
- kernel 7.0 の検証強化 `d3f7d17` "Bluetooth: MGMT: validate Add Extended
  Advertising Data length" (mainline 2026-05-20) が長さ不一致を厳格拒否
  するようになり、従来黙認されていた bug が顕在化

upstream の修正状況 (2026-06-12 時点、いずれも未リリース):

- bluez 側: `2a6968b` "advertising: Fix sending extra bytes with
  MGMT_OP_ADD_EXT_ADV_DATA" (2026-06-02)。5.86 リリース未収録 → **本 pkg で
  これを先取り適用**
- kernel 側: `149324f` "Bluetooth: MGMT: Fix backward compatibility with
  userspace" (mainline 2026-06-03、検証を `expected_len > data_len` に緩和)。
  7.0.12 (core-testing) 未収録

## Source

- 元 PKGBUILD: Arch 公式 extra/bluez 5.86-6 を fork
  https://gitlab.archlinux.org/archlinux/packaging/packages/bluez
  - tag `5.86-6` = commit `624b4486bc99e218632a3d40f4beb055e08d014c`
  - maintainer: Andreas Radke / Robin Candau
- Upstream tarball: https://www.kernel.org/pub/linux/bluetooth/bluez-5.86.tar.xz
  - tag `5.86` = commit `74770b1fd2be612f9c2cf807db81fcdcc35e6560`
    (https://github.com/bluez/bluez)
  - PGP 署名 (.sign) + `validpgpkeys` Marcel Holtmann
    `E932D120BC2AEC444E558F0106CA9F5D1DCF2659` で検証 (公式と同一)
- 補助ファイル:
  - `bluetooth.modprobe` は公式 packaging tag 5.86-6 から直 copy
    (sha256 `46c021be...` が公式 PKGBUILD の pin と一致することを確認済み)
  - `fix-ext-adv-data-extra-bytes.patch` = upstream commit
    `2a6968b40378dca5650e18e03ad0407738c47be5` を `?full_index=1` 形式で
    取得して vendor (公式の他 3 patch と同じ取得形式、ただし本 repo 内に
    ファイル同梱 + sha256 pin)

## 検証結果

- [x] `source` tarball = kernel.org official path、実測 sha256
      `99f144540c6070591e4c53bcb977eb42664c62b7b36cb35a29cf72ded339621d`
      が公式 PKGBUILD の pin と一致 (2026-06-12 取得で再確認)
- [x] `bluetooth.modprobe` sha256 `46c021be659c9a1c4e55afd04df0c059af1f3d98a96338236412e449bf7477b4`
      = 公式 pin と一致
- [x] `fix-ext-adv-data-extra-bytes.patch` sha256
      `01c0126f430596c1cca643e9466c1aa760228a8176800c13c48d7b1f107b5ffe`
  - 中身: `src/advertising.c` の 1 hunk のみ。`param_len` の計算を
    `sizeof(struct mgmt_cp_add_advertising)` → `sizeof(*cp)`
    (= `mgmt_cp_add_ext_adv_data`) に直す 1 行修正。network / exec 無し
  - 5.86 tarball 展開後 `patch -Np1 --dry-run` 成功を確認済み
    (hunk 行番号 1498 → 実際 1487、offset -11、context 完全一致)
  - 同型の `param_len = sizeof(struct mgmt_cp_add_advertising)` が
    advertising.c の legacy `MGMT_OP_ADD_ADVERTISING` 経路 (line 1032、
    そちらは正しい構造体) にもあるが、context が異なり誤適用しないことを
    適用後 grep で確認済み
- [x] 公式既存 3 patch (fix-broken-stdin-handling / revert_e73bf58 /
      a2dp-connect-source-profile-after-sink) との衝突なし — 触るファイルは
      `Makefile.tools` / `src/shared/shell.c` / `profiles/audio/a2dp.c` で
      `src/advertising.c` に触れるのは本 patch のみ
- [x] `prepare()` / `build()` / `check()` / `package_*()`: 公式 5.86-6 から
      patch 適用 3 行 (コメント込み) の追加以外、一切変更なし
- [x] `depends` / `makedepends`: 公式から変更なし
- [x] `makepkg --printsrcinfo` で .SRCINFO 生成、parse 正常
- [x] `vercmp`: `5.86-6 < 5.86-6.1 < 5.86-7 < 5.87` を確認 (公式の次
      リリースが必ず本 shadow より新しい version になる)

## 公式 (extra) との意図的差分

| 変更 | 理由 |
|---|---|
| `pkgrel=6` → `6.1` | 公式 6 と次の 7 の間に挟む。撤去後にクライアントが公式の次 release へ自然に upgrade できる version 設計 |
| `fix-ext-adv-data-extra-bytes.patch` を source[] + prepare() に追加 | kernel 7.0 で LE 広告全滅となるリグレッションの upstream fix (2a6968b) 先取り。これが本 shadow の目的そのもの |
| patch をリモート URL でなくファイル同梱 | 本 repo 規約 (自前ファイルは sha256 必須 / source は repo 内で完結) |

## 撤去条件

**本 pkg は暫定配布**。以下のいずれかが Arch に降りたら CLAUDE.md
「Package の退役」手順で `pkgs/bluez/` を削除して公式に戻す:

1. **extra に bluez ≥ 5.87 (2a6968b 収録) が入る** — nvchecker が 5.87 を
   検知して Issue が立ったら、それは bump ではなく**撤去のトリガー**として
   扱うこと (nvchecker.toml の [bluez] section コメントにも明記)
2. **kernel 7.0.y に `149324f` の backport が入る** — この場合 bluez 5.86
   のままでも動くため shadow 不要。kernel の ChangeLog
   (https://cdn.kernel.org/pub/linux/kernel/v7.x/) で確認

注意: 条件 2 (kernel 側) で撤去した場合、extra がまだ 5.86-6 のままだと
クライアントは installed 5.86-6.1 (or bot bump 後の 7.1 系) > repo 5.86-6 の
「local is newer」状態になる。実害は無く、公式が 5.86-7 / 5.87 に進んだ
時点で自然に解消される。

## 更新方針

- **upstream 5.87+ の検知 = 撤去** (上記)。pkgver bump はしない
- dep-version-pr bot の pkgrel bump は通常フローで処理してよい (rebuild は
  公式の soname rebuild 相当)。bot の regex (`^pkgrel=(\d+)`) は `6.1` の
  整数部のみ match するため bump 結果は `7.1`, `8.1`, ... と `.1` suffix が
  保存される。branch / PR title は `pkgrel-7` 表記になるが動作に問題なし
- 公式が 5.86-7 等の rebuild を出しても repo 並び順で本 shadow が勝ち続ける
  (クライアントには公式 rebuild が届かない)。dep-version-pr が同等の
  rebuild を担うので追従不要

## 結論

**approve** — build host で `bin/build-all bluez` で build + sign + repo db
追加可。split 8 packages なので署名 (YubiKey タッチ) が package 数分必要。

build host 初回のみ、tarball の PGP 検証用に Marcel Holtmann の公開鍵
import が必要: `gpg --recv-keys E932D120BC2AEC444E558F0106CA9F5D1DCF2659`

## 更新履歴

| 日付 | release | PKGBUILD repo SHA | upstream tag commit | 確認内容 |
|---|---|---|---|---|
| 2026-06-12 | 5.86-6.1 | (初版) | `74770b1` (bluez 5.86) | 新規追加 (official-shadow 初版)。extra 5.86-6 (packaging tag `624b448`) fork + upstream `2a6968b` patch 先取り。tarball / bluetooth.modprobe sha256 が公式 pin と一致確認、patch dry-run OK、既存 3 patch と衝突なし。Issue #208 |
| 2026-06-15 | 5.86-7.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libical 4.0.2-1 → 4.0.3-1 |
| 2026-06-17 | 5.86-8.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): alsa-lib 1.2.16-1 → 1.2.16.1-1 |
| 2026-06-18 | 5.86-9.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python-docutils 1:0.22.4-1 → 1:0.23-1 |
| 2026-06-21 | 5.86-10.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libical 4.0.3-1 → 4.0.3-2 |
| 2026-06-25 | 5.86-11.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): systemd / systemd-libs 260.2-2 → 261-1 |
| 2026-06-27 | 5.86-12.1 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): glib2 2.88.1-1 → 2.88.2-1 |
