# pam_pkcs11 review

## 状態

**review 済み、approve** (2026-05-16、Claude review 指摘の fix 反映)

AUR の `pam_pkcs11` PKGBUILD (pkgver=0.6.13, pkgrel=1) を fork。
**改変 4 箇所** (= AUR からの divergence、Claude review 指摘で追加):
1. `md5sums` → `sha256sums`
   (md5 は衝突攻撃に対し暗号学的に破綻、fork する目的が「強化」のため
   sha256 化が筋。新値: `8a853f4e6e136ceecdcffad798570e3d6af2fde08e975656b2dc931989c35aff`)
2. `makedepends` 明示 (`autoconf`, `automake`, `libtool`, `pkgconf`)
   (`./bootstrap` が autotools 系を要求、AUR PKGBUILD は base-devel 暗黙
   依存だが明示する方が re-build / clean build で安全)
3. `license` 値を `LGPL` → `LGPL-2.1-or-later` (SPDX 準拠、upstream COPYING と一致)
4. `depends` に `openssl` を追加 (= `configure.ac` が `with_nss=no` の時
   OpenSSL を hard requirement で要求、`libcrypto.so` / `libssl.so` は
   runtime link 対象。Arch `core` に常駐するため build/run には影響ない
   が、直接 link する runtime dep は明示するのが PKGBUILD の正解)

## debug build について

PKGBUILD の `./configure` 呼び出しで `--disable-debug` がコメントアウト
されている (= AUR から継承):

```bash
./configure --prefix=/usr ... --disable-static
#    --disable-debug # build is broken
```

`--disable-debug` を有効化すると upstream のビルドが壊れる旨 AUR
maintainer が comment 残しており、本 fork でもそのまま継承。結果として
production binary に debug code が含まれるが、セキュリティ上の問題は
なし (= 単に binary size が増える)。upstream で `--disable-debug` が
fix されたら revisit。

## 用途

PAM 経由で X.509 (PIV) cert ベースの認証を可能にする library。
ansible-nekonodesk の `roles/pam_yubikey` が require、Arch クライアント
(= kirigiri / 将来の ayaka Arch 化版) で YubiKey PIV による sudo / swaylock /
login の認証を有効化するための必須前提。

Arch 公式 (core/extra) には無く、AUR の `pam_pkcs11` のみ。AUR を直接
使わない方針なので nekono-pacman-distribution に取り込む。

## Source

- AUR (GitHub mirror 経由 fetch、TLS handshake が AUR 直で fail したため):
  https://github.com/archlinux/aur/blob/pam_pkcs11/PKGBUILD
  - maintainers: Andrea Scarpino / Oleg Smirnov (contributors)
- Upstream: https://github.com/OpenSC/pam_pkcs11
  - OpenSC org の sub-project (= 公式 OpenSC スイートの一部)、LGPL-2.1、77 stars
  - 主開発者 Paul Wolneykien

## 検証結果

- [x] `source` URL = `github.com/OpenSC/pam_pkcs11/archive/pam_pkcs11-0.6.13.tar.gz`
  - OpenSC 公式 organization、PAM PKCS#11 認証スタックの正規 upstream
- [x] `sha256sums` (= 我々が AUR の md5 → sha256 に切替) が upstream tarball と一致
  - 実測: `8a853f4e6e136ceecdcffad798570e3d6af2fde08e975656b2dc931989c35aff`
  - PKGBUILD 値: `8a853f4e6e136ceecdcffad798570e3d6af2fde08e975656b2dc931989c35aff`
  - 一致
  - 参考: AUR PKGBUILD の md5sums `329426f89f13a5374828c35199c54d73` も
    同 tarball で一致確認済み (= 改ざんなし)
- [⚠] Tag `pam_pkcs11-0.6.13` の git commit (`b8dbe6370d36a6a11a466d5f0ee285804103e030`)
      は **GPG verified: false** — OpenSC project は commit signing を運用
      していない。tarball md5 pin で integrity 確保 (= GitHub release CI に
      依存)、author/committer = Paul Wolneykien (= 主開発者) と GitHub
      identity で確認
- [x] `build()`: `./bootstrap && ./configure --prefix=/usr --sysconfdir=/etc
      --with-ldap=no --without-docbook --disable-static && make` — autotools
      標準、network fetch / eval なし
- [x] `package()`: `make DESTDIR install` + `install -d /etc/pam_pkcs11/{cacerts,crls}`
      標準
- [x] `depends`: `pcsclite` — PAM module が PC/SC で smart card を扱う、妥当
- [x] `makedepends`: `autoconf`, `automake`, `libtool`, `pkgconf` — `./bootstrap`
      が autotools を要求、明示 (= AUR は省略してた、Claude review 指摘で追加)
- [x] `backup` に `/etc/pam_pkcs11/{pam_pkcs11.conf,subject_mapping,card_eventmgr.conf}`
      — pacman -Rn で消さない、ansible role pam_yubikey の deploy と整合
- [x] license `LGPL-2.1-or-later` (= SPDX 準拠に修正、upstream COPYING と一致。
      AUR は `LGPL` だけで version 不明だった)

## 結論

**approve** — そのまま build host で `bin/build-all pam_pkcs11` で build +
sign + repo db 追加可。

これにより:
- `apt_packages/vars/Archlinux.yml` の Self-hosted block に `pam_pkcs11`
  を追加して、kirigiri (Arch) でも `pacman -S pam_pkcs11` で install 可能
- `roles/pam_yubikey` の Arch dispatch (= lineinfile で `/etc/pam.d/system-auth`
  に `auth sufficient pam_pkcs11.so` 注入) が動作可能になる

## 更新方針

upstream の新 release (0.6.14 等) が出たら:
1. AUR PKGBUILD の pkgver / md5sums (or sha256sums) を確認
2. 本 dir の PKGBUILD を差し替え
3. md5 / sha256 を独立再計算 (= `curl -fsSL <url> | md5sum` で照合)
4. tag commit author を GitHub API で確認 (= Paul Wolneykien が変わってないか)
5. REVIEW.md に確認日 + 結論 update
