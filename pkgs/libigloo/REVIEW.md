# libigloo review

## 状態

**review 済み、approve** (2026-05-25)

AUR の `libigloo` PKGBUILD (pkgver=0.9.5, pkgrel=2) を fork。
**AUR からの改変なし** — チェックサム・ビルド手順とも問題なし。

## 用途

Icecast 2.5.x が必須とする共通 C フレームワーク (I/O 抽象化層)。
Arch 公式 repo 未収録、AUR のみ。`icecast` の依存として
nekono-pacman-distribution に取り込む。

## Source

- AUR: https://aur.archlinux.org/packages/libigloo
  - Maintainer: Drew Nutter (icecast AUR と同一メンテナ)
- Upstream: Xiph.Org Foundation (icecast と同一組織)
  - tarball: https://ftp.osuosl.org/pub/xiph/releases/igloo/
  - GitHub: https://github.com/xiph/igloo

## 検証結果

- [x] `source` URL = `https://ftp.osuosl.org/pub/xiph/releases/igloo/libigloo-0.9.5.tar.gz`
  - OSU Open Source Lab は Xiph.Org の公式 CDN/ミラー
  - HTTPS 通信、typosquat リスクなし
- [x] `sha256sums` が tarball と一致
  - 独立計算: `ea22e9119f7a2188810f99100c5155c6762d4595ae213b9ac29e69b4f0b87289`
  - PKGBUILD 値: 同上
- [x] `build()`: `./configure --prefix=/usr && make` — 標準 autotools。
  tarball に configure 生成済みのため autoreconf 不要。
  ネットワーク fetch / eval なし。
- [x] `package()`: `make DESTDIR install` — 標準。
- [x] `depends`: `rhash` のみ (Arch 公式 extra、ハッシュライブラリ)
- [x] makedepends 明示なし — tarball が configure を同梱するため
  base-devel (gcc/make) のみで充分。

## 結論

**approve** — build host で `bin/build-all libigloo` 実行可。

## 更新履歴

| 日付 | pkgver-pkgrel | PKGBUILD SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-25 | 0.9.5-2 | (初回 fork) | — | approve、AUR 改変なし |
