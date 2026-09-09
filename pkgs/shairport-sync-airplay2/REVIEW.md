# shairport-sync-airplay2 review

## 状態

**review 済み、approve** (最新: 2026-09-09 / upstream 5.5.1)

upstream の最新 stable を pin。初版 (2026-05-23、4.3.5、当時の
extra/shairport-sync 4.3.5-4 fork) で立ち上げた後、Arch 公式の更新周期から
独立して security fix / bug fix を早期に取り込むため、**upstream release を
直接追従する方針**へ切り替えた (= nvchecker の監視先とも整合)。

configure flags は本 fork の目的 `--with-airplay-2` を追加した上で、 extra
4.3.5 の build 構成 (= 全 backend / metadata / dbus / mpris / mqtt 込み) を
継承。 5.0.x で生じた flag rename (= `--with-pa` → `--with-pulseaudio`、
`--with-pw` → `--with-pipewire`、 `--with-systemd` → `--with-systemd-startup`)
を反映。

Arch 公式版も現在は 5.0.4-3 で `--with-airplay-2` と nqptp / libplist /
ffmpeg を採用しており、「公式版は AirPlay 1 のみ」という初版時の差は解消済み。
本 fork は別 pkgname と upstream 最新追従を維持し、`provides`/`conflicts`/
`replaces=shairport-sync` により公式版を **明示的に置換 install** する
(= pacman.conf の repo 順に依存しない決定論的切替)。

## 用途

ayaka (Arch desktop、Sway / Wayland) で iPhone / Mac / iPad から AirPlay 2
receiver として ayaka を見えるようにする。同時にもう一つの nekono package
`nqptp` を system service として動かす必要があり、 その依存関係を本
PKGBUILD の `depends` で表現している。

`ansible-nekonodesk` の `roles/airplay` が本 pkg を `pacman -S
shairport-sync-airplay2` で install し、 shairport-sync 本体は **systemd-user**
として PipeWire default sink にルーティングする (= 既存 ayaka の運用形態を
維持。 AirPlay 2 化しても systemd-user 起動は変えない)。

**iOS 26 系の "isRemoteControlOnly" 互換性問題 (upstream issue #2179)** は
直接 fix されないまま 2026-06-11 に stale close。5.1 以降の
`service_type=auto` fallback が運用回避策になり得るが、ayaka 側での再有効化は
実機検証と ansible の切替を別途行う。

## Source

- 元 PKGBUILD: Arch 公式 extra/shairport-sync 4.3.5-4 を歴史的 baseline 雛形
  として fork。公式 main は現在 5.0.4-3 で AirPlay 2 対応済み
  https://gitlab.archlinux.org/archlinux/packaging/packages/shairport-sync
  - maintainer: Anatol Pomozov
- Upstream tarball: https://github.com/mikebrady/shairport-sync (tag 5.5)
  - 作者: Mike Brady、 GPL
  - 2026-09-04 release、複数の未認証 remote vulnerability を修正する security release
- 補助ファイル:
  - `shairport-sync.sysusers` は extra packaging から直 copy (4.3.5 ↔ 5.0.4
    で内容変化無し)
  - `remove_useradd.patch` は extra 4.3.5 用の hunk を 5.0.4 Makefile.am の
    行ずれに合わせて再生成 (内容は同一、 行番号 297 → 286)

## 検証結果

- [x] `source` 1 番目 = `github.com/mikebrady/shairport-sync/archive/5.5.tar.gz`
  - 実測 (2026-09-06): `5fcce2ee6b6fbda5fcfb381d0000ab799ec7ce4c285098da6bef82a17237945f`
- [x] `source` 2 番目 = `shairport-sync.sysusers` (= extra packaging から直 copy、
      4.3.5 / 5.5 で同一)
  - sha256: `bc2d92254910996e837d1c4c7dd81eddfb96a9f5f0cb2faad9fcb0414ea79a1d`
- [x] `source` 3 番目 = `remove_useradd.patch` (= 5.0.4 用に hunk 行ずれ再生成、5.5 にも適用可)
  - sha256: `38f5c7aa7a35d1fd1a591b28d50293798cb29c3d9c600c0a750cb4209aed42f1`
  - 中身: `Makefile.am` の `install-systemd-local` target から
    `$(INSTALL_USER_TARGET)` 依存を削るだけの 1 行 diff。 内容は前版と同一
  - 5.5 upstream の Makefile.am で `patch -p1 --dry-run` が offset +21 で成功
- [⚠] Tag `5.5` の target commit に GPG 署名は **無し**。tarball sha256 pin で
      integrity 確保、
      target commit `663499543b535de0e61e0b67afce27f2d637f938`、author = Mike Brady を確認
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
  - 5.2.3 → 5.5 で新たに required な lib は無し (= 既存追加分でカバー)
- [x] `makedepends`: glib2-devel xmltoman vim (= xxd)、 変更なし
- [x] `provides=(shairport-sync)` / `conflicts=(shairport-sync)` /
      `replaces=(shairport-sync)` を 3 点セットで宣言、 変更なし

## Arch 公式 5.0.4-3 との主な差分

| 変更 | 理由 |
|---|---|
| `pkgname=shairport-sync-airplay2` | extra の同名 install 衝突を避け、 user 指示「名前被りは混乱の元」に従う |
| `pkgver=5.5` (公式は 5.0.4) | 本 fork は upstream を直接追従、Arch 公式の更新周期に lock しない (= nvchecker も upstream を見ている) |
| `provides`/`conflicts`/`replaces=shairport-sync` | 別 pkgname にしつつ extra 版を明示的に置換 install させる (`/etc/pacman.conf` の repo 並び順に依存しない) |
| depends に `libsodium` / `libsndfile` を明示 | AirPlay 2 build の upstream 要件として直接表現。nqptp / libplist / ffmpeg と `--with-airplay-2` は公式版も現在採用済み |
| `remove_useradd.patch` は本 fork snapshot を維持 | 4.3.5 用 hunk を 5.0.4 で再生成したもの。5.5 にも offset のみで適用可能 |

## 結論

**approve** — そのまま build host で `bin/build-all shairport-sync-airplay2`
で build + sign + repo db 追加可。

build 順は **nqptp が先**で次に shairport-sync-airplay2 (= depends に nqptp
を持つため、 `bin/build-all` の topological sort で自然にこの順になる)。

完了後、 ansible-nekonodesk の `roles/airplay` を AirPlay 2 化に再切替して
実機検証する時に、そのまま `pacman -S
shairport-sync-airplay2` で install すれば `provides`/`conflicts`/`replaces`
の自動処理で extra/shairport-sync が無くなり本 pkg だけが残る。

## 更新方針

upstream の新 release (5.5.1, 5.6.x 等) が出たら:

1. 本 dir の PKGBUILD の `pkgver` を更新
2. `curl -fsSL <archive URL> | sha256sum` で sha256 再計算し `sha256sums[0]` 更新
3. `cd /tmp/shairport-sync-<new>; patch -p1 --dry-run < remove_useradd.patch`
   で patch 適用性確認。 失敗時は同 hunk を新 Makefile.am の行に合わせて
   再生成 + sha256 再計算 + `sha256sums[2]` 更新
4. configure.ac を grep してbuild flag の rename / deprecation / 新 flag を
   確認、 必要に応じて configure_args を更新
5. `.SRCINFO` を `makepkg --printsrcinfo` で同期
6. REVIEW.md「更新履歴」に 1 行追加
7. upstream の release notes / iOS 26 互換性を確認し、ayaka 側 ansible で
   AirPlay 2 再挑戦できるか判断

extra/shairport-sync の追従は **本 fork の更新方針からは外す**。 extra が
5.x に上がった時の依存変化 (pacman 公式 lib の SONAME 変化等) は
`.deps.lock` + dep-version-pr workflow が自動検出するので、 そちらに任せる。

## 更新履歴

| 日付 | release | PKGBUILD repo SHA | upstream tag commit | 確認内容 |
|---|---|---|---|---|
| 2026-05-23 | 4.3.5 | (初版) | — | 新規追加、 extra 4.3.5-4 fork + `--with-airplay-2`、 sha256 OK、 `remove_useradd.patch` 4.3.5 Makefile.am に適用可確認 |
| 2026-05-24 | 5.0.4 | (bump) | — | upstream latest stable 追従、 configure flag rename 反映 (pa/pw/systemd)、 `--with-ffmpeg` は airplay-2 で auto-on、 `--with-apple-alac` は upstream deprecated のため非採用、 `remove_useradd.patch` を 5.0.4 用に再生成、 全 sha256 verify OK。 ※ ayaka 運用は AirPlay 1 のまま (iOS 26 + issue #2179 待ち) |
| 2026-05-25 | 5.0.4-2 | `620aacc` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libpipewire 1:1.6.5-1 → 1:1.6.5-2 |
| 2026-05-28 | 5.0.4-3 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libpipewire 1:1.6.5-2 → 1:1.6.6-1 |
| 2026-06-01 | 5.0.4-4 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): vim 9.2.0511-1 → 9.2.0573-1 |
| 2026-06-04 | 5.0.4-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): alsa-lib 1.2.15.3-2 → 1.2.16-1 |
| 2026-06-06 | 5.0.4-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): vim 9.2.0573-1 → 9.2.0600-1 |
| 2026-06-10 | 5.0.4-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): openssl 3.6.2-2 → 3.6.3-1 |
| 2026-06-13 | 5.0.4-8 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): vim 9.2.0600-1 → 9.2.0623-1 |
| 2026-06-16 | 5.0.4-9 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): vim 9.2.0623-1 → 9.2.0653-1 |
| 2026-06-17 | 5.0.4-10 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): alsa-lib 1.2.16-1 → 1.2.16.1-1、 vim 9.2.0653-1 → 9.2.0663-1 |
| 2026-06-18 | 5.0.4-11 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): vim 9.2.0663-1 → 9.2.0670-1 |
| 2026-06-20 | 5.0.4-12 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): ffmpeg 2:8.1.1-2 → 2:8.1.2-1 |
| 2026-06-21 | 5.0.4-13 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libpipewire 1:1.6.6-1 → 1:1.6.7-1 |
| 2026-06-23 | 5.0.4-14 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): ffmpeg 2:8.1.2-1 → 2:8.1.2-6、 vim 9.2.0670-1 → 9.2.0699-1 |
| 2026-06-27 | 5.0.4-15 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): avahi 1:0.9rc4-1 → 1:0.9rc5-1、 ffmpeg 2:8.1.2-6 → 2:8.1.2-7 |
| 2026-06-28 | 5.0.4-16 | bot PR #304 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): glib2-devel 2.88.1-1 → 2.88.2-1、 libsoxr 0.1.3-4 → 0.1.3-5、 vim 9.2.0699-1 → 9.2.0735-1 |
| 2026-07-03 | 5.1-1 | (this PR) | upstream tag `5.1` (`d6ac53bf4c6a1ebc55a03177537765ff42dec919`) | safe-to-bump (Issue #334)。 **security update**: AirPlay 2 pairing の TLV parser (`pair_ap/pair-tlv.c`) + DMAP parser (`rtsp.c`) の OOB read / 無限ループ修正 (外部 researcher TristanInSec report, PR #2218, +19/-3)。 `configure.ac` の `--with-*` フラグ群・`PKG_CHECK_MODULES`/`AC_CHECK_LIB` 増減なし、 depends/makedepends 無変化 (`build()` 無改変)。 man page 生成有効化 (xmltoman は既に makedepends)、 systemd unit の nqptp 依存 Requires→Wants 緩和 (自動生成) — build 影響なし。 sha256sums[0] のみ更新 (独立実測 `d85b5ad2…`)、 sysusers/patch (sha[1]/[2]) 据え置き。 `remove_useradd.patch` は 5.1 に dry-run 適用確認済み (Hunk #1 offset +20 で成功、 再生成不要)。 pkgrel reset 16→1。 (参考: iOS 26 `isRemoteControlOnly` upstream issue #2179 は fix ではなく 2026-06-11 stale close、 5.1 の `service_type=auto` フォールバックが運用回避策になり得る) |
| 2026-07-14 | 5.1-1 (deps.lock 訂正、build 影響なし) | (this commit) | — | `.deps.lock` の `# MISSING nqptp (nekono-internal)` は誤り。`nqptp` は Arch 公式 `extra` に実在する (`extra/nqptp 1.2.8-1`、mikebrady/nqptp)。 build-host migration での初回 from-scratch build で `pacman -S nqptp` により正常に解決確認、`.deps.lock` に `nqptp=1.2.8-1` として追記。`jack` は引き続き virtual provide (`pipewire-jack` または `jack2` が満たす) のため MISSING 表記を維持。PKGBUILD/sha256/pkgver は無変更 |
| 2026-07-29 | 5.2-1 | (this PR) | upstream tag `5.2` (`42bd29d6dac605d37356f847334246a78a569276`) | safe-to-bump (Issue #452)。feature + bugfix release (security fix 無し): MQTT `queue_next` コマンド追加、新 exit handler (cleanup 付き終了処理)、metadata 受信方式の互換性回復、convolution/loudness の rate 取り違えバグ修正、`log-to-syslog` オプション廃止 (systemd unit の `ExecStart` から自動除去、build 影響なし)。`configure.ac`/`Makefile.am` は version 文字列 diff のみで `--with-*` フラグ・`PKG_CHECK_MODULES`/`AC_CHECK_LIB` 増減なし、depends/makedepends 無変化。`remove_useradd.patch` は 5.2 でも dry-run 適用確認済み (context 一致、offset のみ増)。sha256sums[0] のみ更新 (独立実測 `17bd4c2d…`)、sysusers/patch (sha[1]/[2]) 据え置き。commit author 一覧に不審な新規 maintainer 無し (Mike Brady 本人 + 既存 contributor + dependabot のみ)。pkgrel reset 1→1 (旧版は 5.1-1 のまま bump なしだったため実質変化なし) |
| 2026-07-30 | 5.2.1-1 | (this PR) | upstream tag `5.2.1` (`08af668a5d17b4714da38981dea4c9039263a4cc`) | safe-to-bump (Issue #459)。pure build-fix release (機能追加/security fix 無し): `--with-libdaemon` 使用時のみ発生する `log_to_syslog()` 未定義参照の残骸 1 箇所を削除 (外部 contributor Daeho Ro 初回貢献、Mike Brady が review・merge)。本 PKGBUILD は `--with-libdaemon` 不使用のため影響無し。diff は `configure.ac`(version 文字列のみ) と `shairport.c`(該当 1 行コメントアウト) の 2 ファイルのみ、depends/makedepends/configure_args 変更不要。tag commit は本 project 従来通り unsigned (既知の運用、tarball sha256 pin で integrity 確保)。sha256sums[0] のみ更新 (独立実測 `8f97d1a6…`)、sysusers/patch (sha[1]/[2]) 据え置き。`remove_useradd.patch` は 5.2.1 でも dry-run 適用確認済み (context 一致、offset のみ増) |
| 2026-08-30 | 5.2.3-1 | (this PR) | upstream tag `5.2.3` | safe-to-bump (Issue #563/#565/#569、5.2.2/5.2.2.1/5.2.2.2/5.2.3 の累積 diff を一括 bump)。`compare/5.2.1...5.2.3` (28 commits) は audio rate/format matching のバグ修正 2 件、convolution コードのメモリ確保バグ修正 (frame length 基準に変更)、AirPlay 2 buffered latency offset の補正修正、MPRIS ドキュメント更新、GitHub Actions dependabot bump のみ。security fix なし。`configure.ac`/`Makefile.am` は version 文字列のみで `--with-*` フラグ・`PKG_CHECK_MODULES`/`AC_CHECK_LIB` 増減なし、depends/makedepends/configure_args 変更不要。tag commit は従来通り unsigned (tarball sha256 pin で integrity 確保)。sha256sums[0] のみ更新 (独立実測 `890eacbc…`)、sysusers/patch (sha[1]/[2]) 据え置き。`remove_useradd.patch` は 5.2.3 でも dry-run 適用確認済み (Hunk #1 offset +21 で成功、context 一致、再生成不要)。commit author 一覧に不審な新規 maintainer 無し (Mike Brady 本人 + 既存 contributor + dependabot のみ)。Closes #563, #565, #569。 |
| 2026-09-06 | 5.5-1 | (this PR) | upstream tag `5.5` (`663499543b535de0e61e0b67afce27f2d637f938`) | **security update、優先 publish** (Issue #600)。5件の High advisory (GHSA-3v9c-6fg5-25pp / GHSA-jgrm-g4c3-wq2r / GHSA-6g79-wrcj-h8xw / GHSA-hf47-mx7r-cr8q / GHSA-536j-295w-5jxr) を修正し、現 AirPlay 2 build は少なくとも前4件の未認証 OOB read / NULL dereference / stack overflow / daemon crash の影響あり。5.2.3...5.5 は17 commits、実効6 files +51/-9。`configure.ac` は version のみ、Makefile / build flags / dependency check / install target は不変で depends/makedepends 改変不要。tag は従来どおり unsigned lightweight tag、release author/target commit は Mike Brady。GitHub archive と codeload の byte一致、tar埋込commit一致、sha256sums[0] を独立実測 `5fcce2ee…` へ更新。sysusers/patch据え置き、patch dry-runはoffset +21で成功。Arch公式mainも現在AirPlay 2対応済みのため古い説明を訂正。`.deps.lock` は現行 Arch repo (ffmpeg 9.0.1-4等) へrefresh。Closes #600。 |
| 2026-09-09 | 5.5.1-1 | (this PR) | upstream tag `5.5.1` (`9229697db2fc30587510a428ae095b84da8b73bf`) | safe-to-bump (Issue #612)。upstream release body: 「This version is identical to Version 5.5, but has been bumped to 5.5.1 to invoke the Docker image-generation workflow.」実際に `compare/5.5...5.5.1` は2 commits (`663499543…5.5` targetの直後に `configure.ac` の version文字列 bump 1 commit + `.github/workflows/codeql.yml` 追加 1 commit、いずれも `[skip ci]`) のみで、変更ファイルは `configure.ac` と `.github/workflows/codeql.yml` の2つだけ。`Makefile.am` / build flags / dependency check / install target に影響する差分なし、depends/makedepends 改変不要。tag target commit は Mike Brady 本人、GitHub archive tar埋込の `AC_INIT([shairport-sync], [5.5.1], ...)` で version 一致確認。sha256sums[0] を独立実測 `5f56571f…` へ更新、sysusers/patch(sha[1]/[2])据え置き。`remove_useradd.patch` は 5.5.1 でも dry-run 適用確認済み (Hunk #1 offset +21 で成功、5.5 と同一)。Closes #612。 |
