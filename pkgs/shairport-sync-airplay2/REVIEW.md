# shairport-sync-airplay2 review

## 状態

**review 済み、approve** (2026-05-23)

Arch 公式 extra/shairport-sync 4.3.5-4 の PKGBUILD を fork。pkgname を
`shairport-sync-airplay2` に rename した上で、configure flags に
`--with-airplay-2` を追加し AirPlay 2 receiver 化。同時に AirPlay 2 mode で
必要となる runtime / build-time 依存 (`nqptp`, `libplist`, `libsodium`,
`libsndfile`, `ffmpeg`) を depends に追加。

extra 版 (= 同 pkgname `shairport-sync`) は `--with-airplay-2` を付けず
ビルドされているため AirPlay 1 のみ。本 fork で AirPlay 2 を有効化した
別 pkgname を作り、`provides`/`conflicts`/`replaces=shairport-sync` を
宣言して extra 版を **明示的に置換 install** する形を取る (= `[nekono]` と
`[extra]` の pacman.conf 並び順に依存しない決定論的切替)。

## 用途

ayaka (Arch desktop、Sway / Wayland) で iPhone / Mac / iPad から AirPlay 2
receiver として ayaka を見えるようにする。同時にもう一つの nekono package
`nqptp` を system service として動かす必要があり、その依存関係を本 PKGBUILD
の `depends` で表現している。

`ansible-nekonodesk` の `roles/airplay` が本 pkg を `pacman -S
shairport-sync-airplay2` で install し、shairport-sync 本体は **systemd-user**
として PipeWire default sink にルーティングする (= 既存 ayaka の運用形態を
維持。AirPlay 2 化しても systemd-user 起動は変えない)。

## Source

- 元 PKGBUILD: Arch 公式 extra/shairport-sync 4.3.5-4
  https://gitlab.archlinux.org/archlinux/packaging/packages/shairport-sync
  - maintainer: Anatol Pomozov
- Upstream tarball: https://github.com/mikebrady/shairport-sync (tag 4.3.5)
  - 作者: Mike Brady、GPL
  - extra 4.3.5 が release 中の現行 stable に追従
- 補助ファイル: extra packaging に同梱されている
  `shairport-sync.sysusers` / `remove_useradd.patch` を **そのまま** 採用
  (= shairport-sync system user / group の作成、makepkg 中の useradd 回避)

## 検証結果

- [x] `source` 1 番目 = `github.com/mikebrady/shairport-sync/archive/4.3.5.tar.gz`
  - extra は `.zip` を使うが本 fork では `.tar.gz` を採用 (= nekono 他 pkg
    と慣習統一)、archive 形式違いに伴い sha256 を再計算
  - 実測 (2026-05-23): `66e985e8e51e8e2c5883c95f68063e2b27d81eb372c8f048acf46dd80a94c118`
- [x] `source` 2 番目 = `shairport-sync.sysusers` (= extra packaging から直 copy)
  - sha256: `bc2d92254910996e837d1c4c7dd81eddfb96a9f5f0cb2faad9fcb0414ea79a1d`
  - 中身: `u shairport-sync - "ShairportSync AirPort receiver" /var/lib/shairport-sync` 等、
    sysusers.d format 標準。改変なし
- [x] `source` 3 番目 = `remove_useradd.patch` (= extra packaging から直 copy)
  - sha256: `3973049b1a92c729efacd3312636a8c7f3d51cd62c7d5af214668c8104378dc6`
  - 中身: `Makefile.am` の `install-systemd-local` target から `$(INSTALL_USER_TARGET)`
    依存を削るだけの 1 行 diff。makepkg 中の root 不要 (= sysusers.d 経由で代替)
  - 4.3.5 upstream の Makefile.am に当該 hunk が一致することを確認済 (= 適用可)
- [⚠] Tag `4.3.5` の git commit GPG 署名は **無し** — mikebrady project は
      commit signing を運用していない。tarball sha256 pin で integrity 確保、
      author = Mike Brady を nqptp 側と同一人物として確認
- [x] `prepare()`: `patch -p1 < remove_useradd.patch` のみ、network / eval 無し
- [x] `build()`: extra と同じ `autoreconf -i -f && ./configure ... && make` +
      `sed` で systemd unit 内の `/usr/local/bin/` → `/usr/bin/` 置換
  - configure_args は extra と同じ flag list に **`--with-airplay-2` 1 個を
    追加しただけ**。他は全て extra のまま
- [x] `package()`: extra と同じ `make DESTDIR install` + sysusers.d / LICENSE
      配置 + sample conf 削除
- [x] `depends`: extra の depends に **`nqptp`, `libplist`, `libsodium`,
      `libsndfile`, `ffmpeg`** を追加
  - 追加分は upstream BUILD.md の AirPlay 2 mode 必須 dep に対応
    (`pacman -Sy git base-devel alsa-lib popt libsoxr avahi libconfig
    libsndfile libsodium ffmpeg vim libplist` のうち、extra で抜けていた分)
  - `nqptp` は同じ nekono repo の自家 pkg、`# MISSING nqptp (nekono-internal)`
    として `.deps.lock` に注釈
- [x] `makedepends`: extra と同じ (`glib2-devel`, `xmltoman`, `vim` (= xxd))
- [x] `provides=(shairport-sync)` / `conflicts=(shairport-sync)` /
      `replaces=(shairport-sync)` を 3 点セットで宣言
  - 初回 install で extra/shairport-sync が自動 uninstall + 本 pkg install
    される (= ansible role 側で明示的に extra 版を消す手順を書かなくて済む)
- [x] `backup=(etc/shairport-sync.conf)` を extra から継承 (= 既存 ayaka の
      `/etc/shairport-sync.conf` (もしあれば) が package 置換時に保護される)

## extra との意図的差分

| 変更 | 理由 |
|---|---|
| `pkgname=shairport-sync-airplay2` | extra の同名 install 衝突を避け、user 指示「名前被りは混乱の元」に従う |
| `provides`/`conflicts`/`replaces=shairport-sync` | 別 pkgname にしつつ extra 版を明示的に置換 install させる (`/etc/pacman.conf` の repo 並び順に依存しない) |
| configure flag `--with-airplay-2` を追加 | 本 fork の目的そのもの (AirPlay 2 mode の有効化) |
| depends に `nqptp`/`libplist`/`libsodium`/`libsndfile`/`ffmpeg` を追加 | AirPlay 2 mode が link / require、upstream BUILD.md 準拠 |
| source archive 形式を `.zip` → `.tar.gz` | nekono 他 pkg と慣習統一 (sha256 は再計算) |

## 結論

**approve** — そのまま build host で `bin/build-all shairport-sync-airplay2`
で build + sign + repo db 追加可。

build 順は **nqptp が先**で次に shairport-sync-airplay2 (= depends に nqptp
を持つため、`bin/build-all` の topological sort で自然にこの順になる)。

完了後、ansible-nekonodesk から `pacman -S shairport-sync-airplay2` で install
すると `provides`/`conflicts`/`replaces` の自動処理で extra/shairport-sync
が無くなり、本 pkg だけが残る。/usr/bin/shairport-sync の中身が AirPlay 2
対応版に置き換わり、ayaka 側 systemd-user の shairport-sync.service
restart で反映される。

## 更新方針

upstream の新 release (4.3.6, 5.0.4 等) が出たら:

1. extra 側 (= Arch 公式 packaging) の追従を**先に確認**:
   `gitlab.archlinux.org/archlinux/packaging/packages/shairport-sync` の
   PKGBUILD diff を読み、`depends` / `configure_args` の構成変化を反映する
2. 本 dir の PKGBUILD の `pkgver`, `sha256sums` (tar.gz 用に再計算), 必要に
   応じて `depends`/`configure_args` を更新
3. `remove_useradd.patch` が新版 Makefile.am に当たるか確認 (`patch --dry-run`)、
   当たらなければ extra 側の更新版 patch を採用
4. `.SRCINFO` も同期
5. REVIEW.md「更新履歴」に 1 行追加

## 更新履歴

| 日付 | release | PKGBUILD repo SHA | 確認内容 |
|---|---|---|---|
| 2026-05-23 | 4.3.5 | (初版) | 新規追加、extra 4.3.5-4 fork + `--with-airplay-2`、sha256 OK、`remove_useradd.patch` 4.3.5 Makefile.am に適用可確認 |
