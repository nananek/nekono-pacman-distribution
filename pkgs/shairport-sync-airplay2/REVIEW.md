# shairport-sync-airplay2 review

## 状態

**review 済み、approve** (2026-05-24、upstream 5.0.4 へ bump)

upstream の最新 stable (5.0.4、 2026-04-27 release) を pin。 初版 (2026-05-23、
4.3.5、 extra/shairport-sync 4.3.5-4 fork) で立ち上げたが、 extra はまだ 4.3.5
止まりで upstream の bug fix / feature を取り込めないため、 本 fork は extra に
追従するのではなく **upstream を直接追従する方針** に切替える (= nvchecker
の monitoring が「upstream の release」なのと整合)。

configure flags は本 fork の目的 `--with-airplay-2` を追加した上で、 extra
4.3.5 の build 構成 (= 全 backend / metadata / dbus / mpris / mqtt 込み) を
継承。 5.0.x で生じた flag rename (= `--with-pa` → `--with-pulseaudio`、
`--with-pw` → `--with-pipewire`、 `--with-systemd` → `--with-systemd-startup`)
を反映。

extra 版 (= 同 pkgname `shairport-sync`) は `--with-airplay-2` を付けずに
ビルドされている (AirPlay 1 のみ)。 本 fork で AirPlay 2 を有効化した別
pkgname を作り、 `provides`/`conflicts`/`replaces=shairport-sync` を宣言して
extra 版を **明示的に置換 install** する形を取る (= `[nekono]` と `[extra]`
の pacman.conf 並び順に依存しない決定論的切替)。

## 用途

ayaka (Arch desktop、Sway / Wayland) で iPhone / Mac / iPad から AirPlay 2
receiver として ayaka を見えるようにする。同時にもう一つの nekono package
`nqptp` を system service として動かす必要があり、 その依存関係を本
PKGBUILD の `depends` で表現している。

`ansible-nekonodesk` の `roles/airplay` が本 pkg を `pacman -S
shairport-sync-airplay2` で install し、 shairport-sync 本体は **systemd-user**
として PipeWire default sink にルーティングする (= 既存 ayaka の運用形態を
維持。 AirPlay 2 化しても systemd-user 起動は変えない)。

**iOS 26 系の "isRemoteControlOnly" 互換性問題 (upstream issue #2179、 open)**
が現時点 (2026-05-24) で fix 未 merge のため、 5.0.4 でも audio が流れない
状態は継続見込み。 本 PKGBUILD は upstream の修正が降りた時に bump で取り
込める受け皿として整備する位置付け。 ayaka 側 ansible は AirPlay 1 (=
extra/shairport-sync) に戻して運用している。

## Source

- 元 PKGBUILD: Arch 公式 extra/shairport-sync 4.3.5-4 を baseline 雛形として
  fork 済み (configure_args 構成 / makedepends / sysusers.d 経路 / patch 経路
  はそのまま継承)
  https://gitlab.archlinux.org/archlinux/packaging/packages/shairport-sync
  - maintainer: Anatol Pomozov
- Upstream tarball: https://github.com/mikebrady/shairport-sync (tag 5.0.4)
  - 作者: Mike Brady、 GPL
  - 2026-04-27 release、 PulseAudio bugfix release (= 5.0.x 系)
- 補助ファイル:
  - `shairport-sync.sysusers` は extra packaging から直 copy (4.3.5 ↔ 5.0.4
    で内容変化無し)
  - `remove_useradd.patch` は extra 4.3.5 用の hunk を 5.0.4 Makefile.am の
    行ずれに合わせて再生成 (内容は同一、 行番号 297 → 286)

## 検証結果

- [x] `source` 1 番目 = `github.com/mikebrady/shairport-sync/archive/5.0.4.tar.gz`
  - 実測 (2026-05-24): `b89d4af74cffadd83d1be6eaf4e967180aa5a6aed32f561c937ae1d787909c25`
- [x] `source` 2 番目 = `shairport-sync.sysusers` (= extra packaging から直 copy、
      4.3.5 / 5.0.4 で同一)
  - sha256: `bc2d92254910996e837d1c4c7dd81eddfb96a9f5f0cb2faad9fcb0414ea79a1d`
- [x] `source` 3 番目 = `remove_useradd.patch` (= 5.0.4 用に hunk 行ずれ再生成)
  - sha256: `38f5c7aa7a35d1fd1a591b28d50293798cb29c3d9c600c0a750cb4209aed42f1`
  - 中身: `Makefile.am` の `install-systemd-local` target から
    `$(INSTALL_USER_TARGET)` 依存を削るだけの 1 行 diff。 内容は前版と同一
  - 5.0.4 upstream の Makefile.am で `patch -p1 --dry-run` が成功すること
    を確認済み
- [⚠] Tag `5.0.4` の git commit GPG 署名は **無し** — mikebrady project は
      commit signing を運用していない。 tarball sha256 pin で integrity 確保、
      author = Mike Brady を nqptp 側と同一人物として確認
- [x] `prepare()`: `patch -p1 < remove_useradd.patch` のみ、 network / eval 無し
- [x] `build()`: `autoreconf -i -f && ./configure ... && make` + `sed` で
      systemd unit 内の `/usr/local/bin/` → `/usr/bin/` 置換
  - configure_args は **extra 4.3.5 と同じ flag list + 5.0.x の flag rename
    対応 + `--with-airplay-2`**。 `--with-ffmpeg` は `--with-airplay-2` で
    auto-on (configure.ac の "if test airplay_2 -o ffmpeg then using_ffmpeg=true"
    で判定) のため明示せず
  - `--with-apple-alac` は upstream で deprecated 化 (= issue #2178 で macOS
    Realtime AirPlay stream を壊す報告) のため採用せず
- [x] `package()`: `make DESTDIR install` + sysusers.d / LICENSE 配置 +
      sample conf 削除 (extra 継承、 変更なし)
- [x] `depends`: openssl avahi libsoxr popt alsa-lib libconfig libpipewire
      libpulse jack mosquitto nqptp libplist libsodium libsndfile ffmpeg
  - 5.0.x で新たに required な lib は無し (= 既存追加分でカバー)
- [x] `makedepends`: glib2-devel xmltoman vim (= xxd)、 変更なし
- [x] `provides=(shairport-sync)` / `conflicts=(shairport-sync)` /
      `replaces=(shairport-sync)` を 3 点セットで宣言、 変更なし

## extra との意図的差分

| 変更 | 理由 |
|---|---|
| `pkgname=shairport-sync-airplay2` | extra の同名 install 衝突を避け、 user 指示「名前被りは混乱の元」に従う |
| `pkgver=5.0.4` (extra は 4.3.5) | 本 fork は upstream を直接追従、 extra に lock しない (= nvchecker も upstream を見ている) |
| `provides`/`conflicts`/`replaces=shairport-sync` | 別 pkgname にしつつ extra 版を明示的に置換 install させる (`/etc/pacman.conf` の repo 並び順に依存しない) |
| configure flag `--with-airplay-2` を追加 | 本 fork の目的そのもの (AirPlay 2 mode の有効化) |
| configure flag `--with-pa` → `--with-pulseaudio`、 `--with-pw` → `--with-pipewire`、 `--with-systemd` → `--with-systemd-startup` | 5.0.x の rename 反映 (extra の flag のままだと configure error) |
| depends に `nqptp`/`libplist`/`libsodium`/`libsndfile`/`ffmpeg` を追加 | AirPlay 2 mode が link / require、 upstream BUILD.md 準拠 |
| source archive 形式を `.zip` → `.tar.gz` | nekono 他 pkg と慣習統一 (sha256 は再計算) |
| `remove_useradd.patch` を 5.0.4 用に hunk 行ずれ再生成 | 4.3.5 用の `-297` 文脈は 5.0.4 で `-286` にずれた (内容は同一の 1 行 diff) |

## 結論

**approve** — そのまま build host で `bin/build-all shairport-sync-airplay2`
で build + sign + repo db 追加可。

build 順は **nqptp が先**で次に shairport-sync-airplay2 (= depends に nqptp
を持つため、 `bin/build-all` の topological sort で自然にこの順になる)。

完了後、 ansible-nekonodesk の `roles/airplay` を AirPlay 2 化に再切替する
時 (= upstream issue #2179 が fix されたら) に、 そのまま `pacman -S
shairport-sync-airplay2` で install すれば `provides`/`conflicts`/`replaces`
の自動処理で extra/shairport-sync が無くなり本 pkg だけが残る。

## 更新方針

upstream の新 release (5.0.5, 5.1.x 等) が出たら:

1. 本 dir の PKGBUILD の `pkgver` を更新
2. `curl -fsSL <archive URL> | sha256sum` で sha256 再計算し `sha256sums[0]` 更新
3. `cd /tmp/shairport-sync-<new>; patch -p1 --dry-run < remove_useradd.patch`
   で patch 適用性確認。 失敗時は同 hunk を新 Makefile.am の行に合わせて
   再生成 + sha256 再計算 + `sha256sums[2]` 更新
4. configure.ac を grep してbuild flag の rename / deprecation / 新 flag を
   確認、 必要に応じて configure_args を更新
5. `.SRCINFO` を `makepkg --printsrcinfo` で同期
6. REVIEW.md「更新履歴」に 1 行追加
7. upstream の release notes / open issue (= 特に iOS 26 関連 issue #2179) を
   確認、 ayaka 側 ansible で AirPlay 2 再挑戦できるか判断

extra/shairport-sync の追従は **本 fork の更新方針からは外す**。 extra が
5.x に上がった時の依存変化 (pacman 公式 lib の SONAME 変化等) は
`.deps.lock` + dep-version-pr workflow が自動検出するので、 そちらに任せる。

## 更新履歴

| 日付 | release | PKGBUILD repo SHA | upstream tag commit | 確認内容 |
|---|---|---|---|---|
| 2026-05-23 | 4.3.5 | (初版) | — | 新規追加、 extra 4.3.5-4 fork + `--with-airplay-2`、 sha256 OK、 `remove_useradd.patch` 4.3.5 Makefile.am に適用可確認 |
| 2026-05-24 | 5.0.4 | (bump) | — | upstream latest stable 追従、 configure flag rename 反映 (pa/pw/systemd)、 `--with-ffmpeg` は airplay-2 で auto-on、 `--with-apple-alac` は upstream deprecated のため非採用、 `remove_useradd.patch` を 5.0.4 用に再生成、 全 sha256 verify OK。 ※ ayaka 運用は AirPlay 1 のまま (iOS 26 + issue #2179 待ち) |
| 2026-05-25 | 5.0.4-2 | `620aacc` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libpipewire 1:1.6.5-1 → 1:1.6.5-2 |
