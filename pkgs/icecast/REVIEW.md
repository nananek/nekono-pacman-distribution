# icecast review

## 状態

**review 済み、approve** (2026-05-25)

AUR の `icecast` PKGBUILD (pkgver=2.5.0, pkgrel=1) を fork。
**AUR からの改変なし** — sha512+b2 二重ハッシュが全ファイルに適用済みで
チェックサム・ビルド手順とも問題なし。

## 用途

HTTP/ICY プロトコルによる音声ストリーミングサーバー。
PipeWire monitor → ffmpeg → Icecast → iPhone VLC (Tailscale 経由) の
構成で ansible-nekonodesk の `audio_stream` role が使用する。
Arch 公式 repo 未収録、AUR のみ。依存の `libigloo` も nekono repo に
取り込む（同時追加）。

## Source

- AUR: https://aur.archlinux.org/packages/icecast
  - Maintainer: Drew Nutter。Contributor に David Runge (dvzrv@archlinux.org、
    Arch Linux Trusted User) を含む。
- Upstream: Xiph.Org Foundation
  - tarball: https://downloads.us.xiph.org/releases/icecast/
  - GitHub: https://github.com/xiph/icecast-server

## 検証結果

### Source URL

- [x] tarball: `https://downloads.us.xiph.org/releases/icecast/icecast-2.5.0.tar.gz`
  - Xiph.Org 公式 CDN、HTTPS、typosquat リスクなし
- [x] 4 つの補助ファイル (icecast.service / .logrotate / .sysusers / .tmpfiles)
  は AUR repo に同梱 (ネットワーク fetch なし)

### チェックサム

- [x] tarball sha512 を独立計算:
  `d92ce5d8ae1cd011eaa8c7424adea744f35e5c2d3e8244d362743be1c6bbc8fc44d76d7a212cf1eebe79da9b7d83b2ed5ab8659fb97929af316674b5ddf590b5`
  PKGBUILD 値と一致。
- [x] 補助ファイル sha512 を独立計算 (全4ファイル PKGBUILD 値と一致):
  - icecast.service:   `89f656...d878b5`
  - icecast.logrotate: `1727ec...6ad14a`
  - icecast.sysusers:  `ca0c6e...04c8`
  - icecast.tmpfiles:  `db3cf0...b7813`
- [x] b2sums も全ファイル pin 済み (sha512 と二重保護)

### ビルド手順

- [x] `prepare()`: `autoreconf -vfi` — 標準 autotools bootstrap
- [x] `build()`: `./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var && make`
  ネットワーク fetch / eval / inject なし
- [x] `package()`: `make DESTDIR install` + 補助ファイルの `install -vDm`
  PATH への書き込みは `/usr/bin/icecast` のみ

### 補助ファイル内容

- [x] `icecast.service`: 強固な systemd hardening
  (CapabilityBoundingSet, ProtectSystem=strict, NoNewPrivileges,
  PrivateUsers, RestrictAddressFamilies=AF_INET AF_INET6 のみ 等)
- [x] `icecast.sysusers`: `icecast` system user を作成 (UID 自動)
- [x] `icecast.tmpfiles`: `/etc/icecast.xml` を 0640 root:icecast に設定
  (ansible role の template deploy と整合)
- [x] `icecast.logrotate`: 標準 log rotation、postrotate は空

### 依存関係

- [x] depends: glibc / libxml2 / libxslt / openssl / rhash / speex /
  libtheora (全て Arch 公式 extra/core)
- [x] depends: libigloo (nekono repo で別途提供)
- [x] depends (dynamic): libcurl.so / libogg.so / libvorbis.so — runtime link、
  package() で `depends+=` している (Arch 慣行)
- [x] makedepends: autoconf / automake / curl / libogg / libtool / libvorbis
  (全て Arch 公式)
- [x] optdepends: libmaxminddb (GeoIP、未使用)

## 依存方針 (AUR との diff なし)

依存は AUR 通り。`libigloo` は nekono repo で提供するので、
`pacman -S icecast` 時に nekono repo から自動解決される。

## 結論

**approve** — `libigloo` を先に build + sign 後、
build host で `bin/build-all icecast` 実行可。

## 更新履歴

| 日付 | pkgver-pkgrel | PKGBUILD SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-25 | 2.5.0-1 | (初回 fork) | — | approve、AUR 改変なし |
| 2026-06-06 | 2.5.0-2 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libtool 2.6.0+r23+gb08cb0a0-1 → 2.6.1-1 |
| 2026-06-10 | 2.5.0-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): openssl 3.6.2-2 → 3.6.3-1 |
| 2026-06-27 | 2.5.0-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): curl 8.20.0-7 → 8.21.0-1 |
