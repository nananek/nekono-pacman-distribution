# nqptp review

## 状態

**review 済み、approve** (2026-05-23)

新規 PKGBUILD (AUR には `nqptp-git` のみで、nekono 規約「`-git` 抑制 / stable
tag pin」と整合しないため自作)。Mike Brady の上流 release tag 1.2.8 を pin。

## 用途

`shairport-sync` を AirPlay 2 mode で動かす際に必須となる timing/sync
daemon。PTPv2 (UDP 319/320) を排他的に握り、AirPlay 2 sender (iPhone / Mac /
HomePod 等) との同期 clock 源を提供する。

ansible-nekonodesk の `roles/airplay` (= ayaka) が `shairport-sync-airplay2`
と組で require する system service。`/usr/lib/systemd/system/nqptp.service`
として install され、root 権限で 319/320 を listen する (= raw socket /
privileged port 確保のため user 化はできない)。

Arch 公式 (core / extra) にも AUR にも stable tag pin の選択肢が無く、
nekono-pacman-distribution に取り込む。

## Source

- Upstream: https://github.com/mikebrady/nqptp
  - 作者: Mike Brady (shairport-sync 主開発者と同一)、GPL-3.0-or-later
  - 1.2.8 (2026-05-13 release)
- AUR `nqptp-git` (= rolling) は今回参照のみで code は使わない

## 検証結果

- [x] `source` URL = `github.com/mikebrady/nqptp/archive/1.2.8.tar.gz`
  - shairport-sync 主開発者の正規 upstream
- [x] `sha256sums` が upstream tarball と一致
  - 実測 (2026-05-23): `3a2882a299c21605f53bb215ce537f9cc7a1e894476f639ab28562c68fd183a9`
  - PKGBUILD 値: 同上
- [⚠] Tag `1.2.8` の git commit は **GPG verified: false** — mikebrady project
      は commit signing を運用していない。tarball sha256 pin で integrity 確保、
      author/committer = Mike Brady (GitHub identity) と shairport-sync 側と
      同一作者であることで確認
- [x] `prepare()`: `autoreconf -fi` (autotools 再生成のみ、network / eval 無し)
- [x] `build()`: `./configure --prefix=/usr --with-systemd-startup
      --with-systemdsystemunitdir=/usr/lib/systemd/system && make` — autotools
      標準、追加 fetch 無し
- [x] `package()`: `make DESTDIR="$pkgdir" install` のみ
- [x] `depends`: `glibc` のみ — configure.ac は pthread / librt のみ要求、
      どちらも glibc が provide
- [x] `makedepends`: `autoconf`, `automake`, `libtool`, `pkgconf` — `autoreconf
      -fi` が autotools を要求、明示
- [x] 上流 Makefile.am で `BUILD_FOR_LINUX && INSTALL_SYSTEMD_STARTUP` が両方
      満たされて初めて systemd unit が `$(systemdsystemunitdir)` (= 我々が
      `/usr/lib/systemd/system` を渡す) に install される

## AUR / extra との差分

- AUR `nqptp-git` は git HEAD を rolling 追従する pkg。本 fork は安定 tag pin
  に切替 (nekono 規約)。

## 結論

**approve** — そのまま build host で `bin/build-all nqptp` で build + sign +
repo db 追加可。

完了後、ansible-nekonodesk の `roles/airplay` から `pacman -S nqptp` で取得
可能になり、`systemctl enable --now nqptp` で UDP 319/320 を listen 開始
する。

## 更新方針

upstream の新 release (1.2.9 等) が出たら:

1. 本 dir の PKGBUILD の `pkgver` を更新
2. `curl -fsSL <archive URL> | sha256sum` で sha256 再計算し
   `sha256sums` 更新
3. `.SRCINFO` の `pkgver` と `sha256sums` を同期
4. tag commit author が引き続き Mike Brady であることを GitHub UI で確認
5. configure.ac / Makefile.am に新 dep や install path 変更がないか diff 確認
6. REVIEW.md「更新履歴」に 1 行追加

## 更新履歴

| 日付 | release | PKGBUILD repo SHA | 確認内容 |
|---|---|---|---|
| 2026-05-23 | 1.2.8 | (初版) | 新規追加、Mike Brady upstream、sha256 OK |
