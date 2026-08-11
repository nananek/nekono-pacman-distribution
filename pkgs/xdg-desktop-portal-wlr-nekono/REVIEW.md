# xdg-desktop-portal-wlr-nekono review

## 状態

**review 済み、approve** (2026-08-06、fork + 改変あり、詳細下記)

## 用途

ayaka (Sway) で HDMI キャプチャ (AVerMedia Live Gamer MINI) を、ウィンドウを
経由せず直接 Discord の画面共有に流すため。

xdg-desktop-portal の ScreenCast interface のソース型は `MONITOR` /
`WINDOW` / `VIRTUAL` のみで、「V4L2 デバイスを直接ソースとして選ぶ」口が無い
(client 側を何をどういじっても出てこない、これは client ではなく OS 側 API
の制約)。そこで xdg-desktop-portal-wlr の chooser に第三の返り値
`Node: <pipewire node id>` を追加する downstream patch を自家 fork に載せ、
これを package 化して extra 版を置換する。

根拠は xdg-desktop-portal 1.22.1 `src/screen-cast.c` の
`open_pipewire_screen_cast_remote()`: backend が `Start()` で返した node id
に `PW_PERM_RWX` を与え、それ以外を `PW_ID_ANY, 0` で隠すのみで、node の
所有者検証は行っていない。したがって backend (xdpw) が既存の任意の
PipeWire node id をそのまま返せば、client はそれを自分が作った capture
node と区別せず受け取る。

ayaka で実測済み (2026-08-06): Firefox 上の Discord に 1920x1080 BGRx 60fps
が compositor を一切経由せず届いた。ログにも `wlroots: pipewire node: 177`
→ `dbus: start: returning node 177` と出て、wlroots 側のコードパスを通らず
抜けている。

## Source

- **原本は AUR ではなく Arch 公式 packaging repo**:
  `https://gitlab.archlinux.org/archlinux/packaging/packages/xdg-desktop-portal-wlr`
  (review 時点 `main` HEAD commit `ca8ab29`, pkgver=0.8.3-1、
  maintainer: Christian Rebischke, Carl Smedstad)
- **本 package の実際の source は Arch 公式 tarball ではなく自家 fork の
  署名付き git tag**:
  `git+https://github.com/nananek/xdg-desktop-portal-wlr.git#tag=v0.8.4-nekono1?signed`
  - upstream (`emersion/xdg-desktop-portal-wlr`) の tag `v0.8.4`
    (commit `34153094662acd713241ca6cbbb20003ef67da5f`) を基点に
    `nekono-node-source` branch へ downstream patch 1 commit を積んだもの
  - fork tag `v0.8.4-nekono1` (commit `5459ef00857009f5963f455e180c9da430fe4658`)
    は Nekono GPG (`483DC691DF9F29327EA106BD030130E2F156CD74`) で署名済み。
    `git tag -v v0.8.4-nekono1` で ultimate trust の正しい署名を確認した
    (2026-08-06)
  - fork repo 直下の `CLAUDE.md` に「upstream には issue も PR も出さない」
    方針が明記されている。このパッチは wlroots portal backend というスコープ
    外の提案であり、fork はローカル用途限定

## 検証結果

- [x] `source` の主要 URL = 自家 fork `github.com/nananek/xdg-desktop-portal-wlr`
  の署名付き git tag。typosquat / domain spoof リスク無し (自分の repo)
- [x] tag 署名 (`?signed` + `validpgpkeys`) を独立検証:
  `git tag -v v0.8.4-nekono1` → `EDDSA鍵483DC691DF9F29327EA106BD030130E2F156CD74
  ... 正しい署名 [究極]`。GitHub API でも同 commit sha
  (`5459ef00857009f5963f455e180c9da430fe4658`) を確認
- [x] `wlr-portals.conf` の sha256 を独立検証:
  `85b84ba8ba69e41295e7f02045c67b547df6eab1b137502b089ddf354f2a04de`
  (Arch 公式 packaging repo の同名ファイルを `git clone` して
  `sha256sum` で照合、内容は2行 `[preferred]` / `default=wlr` のみ)
- [x] patch 本体 (`7105198 screencast: add "Node: <id>" chooser source type`)
  の diff を目視確認。`include/screencast_common.h` /
  `src/screencast/chooser.c` / `src/screencast/screencast.c` /
  `xdg-desktop-portal-wlr.5.scd` の4ファイルのみ。動的コマンド構築 / eval /
  network fetch 無し。既存 PipeWire node id を `Start()` の `node_id` に
  そのまま流用しているだけで、**新しい攻撃面 (任意コード実行や新規
  ネットワーク経路) を追加していない**
- [x] `build()`: `arch-meson -Dsd-bus-provider=libsystemd build && ninja -C build`
      — extra 版と無改変
- [x] `package()`: `ninja -C build install` + `wlr-portals.conf` install
      — extra 版と無改変 (`cd "$pkgname-$pkgver"` → `cd "$pkgname"` のみ、
      git source でのディレクトリ名変更に伴う機械的差し替え)
- [x] `depends`: extra 版のまま無改変。patch は既存の PipeWire / sd-bus しか
      触っておらず新規依存を増やしていない
- [x] `makedepends`: extra 版 (`meson scdoc wayland-protocols`) + `git`
      (git source のため追加)
- [x] `provides`: `xdg-desktop-portal-wlr=$pkgver` (extra 版置換用) +
      `xdg-desktop-portal-impl` (**extra 版から必ず引き継ぐ**。落とすと
      `xdg-desktop-portal` の依存解決が壊れる)
- [x] `conflicts` / `replaces`: `xdg-desktop-portal-wlr` — extra 版と共存
      不可にして置換する
- [x] secrets 混入なし (`.git` 削除済み、秘密鍵 / token 等なし)

## extra 版との意図的 diff (repo 規約: 「なぜ AUR/extra と違うのか」を明記)

| 箇所 | 変更 | 理由 |
|------|------|------|
| `pkgname` | `xdg-desktop-portal-wlr` → `xdg-desktop-portal-wlr-nekono` | `pacman.conf` の repo 順は `[core]` → `[extra]` → `[nekono]`。同名だと extra 版が勝ち `[nekono]` 版がインストールされない。repo 順の入れ替えは影響範囲が広すぎるため不採用、別名 + provides/conflicts/replaces で置換する方針とした |
| `provides` / `conflicts` / `replaces` | `provides=("xdg-desktop-portal-wlr=$pkgver" 'xdg-desktop-portal-impl')` / `conflicts=('xdg-desktop-portal-wlr')` / `replaces=('xdg-desktop-portal-wlr')` を追加 | extra 版の置換に必須。`xdg-desktop-portal-impl` の provide を落とすと `xdg-desktop-portal` パッケージの依存解決 (「backend を1つ以上要求」) が壊れる |
| `source` | upstream tarball (+ sha512sums) → 自家 fork の署名付き git tag (`?signed` + `validpgpkeys`) | fork なので GitHub archive tarball の hash pin より tag の GPG 署名検証の方が強い証跡になる。この repo の GPG 中心の運用 (Nekono trust chain) と一貫させる |
| `makedepends` | `git` を追加 | git source を使うため makepkg が `git` binary を要求する |

`pkgver` は upstream (emersion) の値をそのまま使い、nekono 印は入れていない
(= `0.8.4.nekono1` のような mangling はしない)。理由は nvchecker: upstream
の release を追わせて「そろそろ fork を rebase する時期」のシグナルとして
機能させるため。fork の世代は git tag 側 (`v$pkgver-nekonoN`) で表現する。
`nvchecker.toml` の追跡対象も **fork ではなく upstream (emersion)** にして
ある。

## 依存方針

`depends` / `build()` / `package()` は Arch 公式 extra 版から無改変
(patch が新規依存を増やしていないため)。`pipewire-session-manager` は
virtual provide なので `.deps.lock` では監視対象外 (`# MISSING` 行)。

## 結論

**approve** — build host で `bin/build-all xdg-desktop-portal-wlr-nekono`
で build + sign + repo db 追加可。

これにより:
- ayaka で `pacman -S xdg-desktop-portal-wlr-nekono` により extra 版
  `xdg-desktop-portal-wlr` を置換し、Node source chooser patch 入りの
  xdpw を配布可能にする

## 更新方針

upstream (`emersion/xdg-desktop-portal-wlr`) が新 release を出したら
(= nvchecker が検知):
1. fork repo (`~/repos/xdg-desktop-portal-wlr`) で
   `git fetch upstream --tags && git rebase v<new> nekono-node-source`
   (conflict は主に `chooser.c` / `screencast.c`)
2. build して動作確認 (画面共有 chooser に `Node:` が出て映像が流れること)
3. `git tag -s v<new>-nekono1` + push、GPG 署名を独立検証
4. 本 PKGBUILD の `pkgver` を上げ、`source=` の tag 参照を更新、
   `wlr-portals.conf` の sha256 は Arch 公式 packaging repo の最新版と
   再照合 (変更されていれば再計算)
5. tag 署名 + build host `makepkg --verifysource` で検証後 commit
6. REVIEW.md 更新履歴に追記

**upstream に issue / PR を出さないこと。** fork repo の CLAUDE.md にも
明記されている方針。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-08-06 | 0.8.4-1 (fork tag v0.8.4-nekono1) | (this commit) | `34153094662acd713241ca6cbbb20003ef67da5f` (emersion v0.8.4) / fork tag commit `5459ef00857009f5963f455e180c9da430fe4658` | 新規追加。extra 版 (Arch 公式 packaging repo, main HEAD `ca8ab29`, 0.8.3-1) をベースに pkgname/provides/conflicts/replaces/source/makedepends を改変。patch 本体・tag 署名・wlr-portals.conf sha256 を独立検証、approve |
| 2026-08-10 | 0.8.4-2 | — (pkgrel bump のみ) | — | `wayland` 1.25.0-1 → 1.26.0-1 に伴う rebuild、既存 review 内容に変更なし |
| 2026-08-11 | 0.8.4-3 | — (pkgrel bump のみ) | — | `meson` 1.11.2-1 → 1.12.0-1 に伴う rebuild、既存 review 内容に変更なし |
