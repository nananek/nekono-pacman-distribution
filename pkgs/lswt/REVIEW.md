# lswt review

## 状態

**review 済み、approve** (2026-08-06、純 fork、改変なし)

AUR の `lswt` PKGBUILD (pkgver=2.0.0, pkgrel=1) を **純 fork**。唯一の diff は
`# Maintainer:` → `# Contributor:` の置換 + fork 説明コメント追加、および
`package()` 内の trailing space 除去のみ (= 関数本体・依存・hash とも無改変)。

## 用途

Wayland toplevel (= 開いているウィンドウ) 一覧を取得する CLI ツール。ayaka
(Sway) で Discord の画面共有を「窓単位」で行うために必要。

`xdg-desktop-portal-wlr 0.8.2` (extra、ayaka では導入済み) の OUTPUT CHOOSER
は `Window: <ext-foreign-toplevel-list-v1 identifier as given by lswt(1)>`
形式の chooser output を要求する。この identifier を取得できるツールが
`lswt` 以外に無い。

2026-05-06 に Debian trixie 上で同種の試みを断念した経緯があるが (xdpw
0.7.x は output 全体キャプチャのみ)、Arch 移行後は xdpw 0.8.2 + sway 1.12
(wlroots 0.20.2) となり窓単位キャプチャが可能になっている。ayaka の sway
binary が `wlr_ext_foreign_toplevel_image_capture_source_manager_v1_create` /
`wlr_ext_image_copy_capture_manager_v1_create` を advertise していることを
実機で確認済み。

Arch 公式 (core/extra) には無く AUR のみ存在するため nekono-pacman-
distribution に取り込む。

## Source

- AUR: https://aur.archlinux.org/packages/lswt
  - maintainer: Peter Kaplan `<peter@pkap.de>`
- Upstream: https://sr.ht/~leon_plickat/lswt (Leon Henrik Plickat)
  - AUR `url` フィールドと一致 (typosquat / なりすましリスクなし)
- `-git` 版 (`lswt-git`) もあるが repo 規約どおり不採用 (rolling は avoid)

## 検証結果

- [x] `source` URL = `https://git.sr.ht/~leon_plickat/lswt/archive/v2.0.0.tar.gz`
  - `git ls-remote --tags https://git.sr.ht/~leon_plickat/lswt` で確認した
    最新 tag は `v2.0.0` (他は v1.0.0〜v1.0.4)。AUR は最新に追随済み
- [x] `sha256sums` 独立検証
  - 実測: `curl -fsSL <url> | sha256sum` = `8e23cc5c00bb715b0a1610938111cb76eb9efe1eea87408123620a8a7155e6ab`
  - PKGBUILD 値と一致 (2026-08-06)
- [⚠] Tag `v2.0.0` の GPG 検証は未実施 (= sr.ht 上で commit signing を運用
      している様子は無い)。tarball sha256 pin で integrity 確保
  - **既知リスク**: source は sr.ht の自動生成 archive tarball
    (`/archive/vX.Y.Z.tar.gz`)。sr.ht 側のアーカイブ生成方式が変われば
    (= tar フォーマット・timestamp 等) 同一 tag でも hash がズレうる。
    bump 時は必ず sha256 を再計算すること
- [x] tarball 中身: `lswt-v2.0.0/{LICENSE,Makefile,README,lswt.c,lswt.1,
      bash-completion,.gitignore}` + `ext-foreign-toplevel-list-v1.xml` /
      `wlr-foreign-toplevel-management-unstable-v1.xml`
      → protocol XML は tarball に vendor 同梱済み
- [x] `build()`: `make` のみ。Makefile は `wayland-scanner` で XML →
      C/H を生成し `-lwayland-client` で link するだけ。network fetch /
      動的コマンド構築 / eval / exec なし
- [x] `package()`: `make DESTDIR="$pkgdir" PREFIX="/usr" install` +
      LICENSE install。標準的な install target、怪しい step なし
- [x] `depends`: `wayland` のみ、妥当 (`libwayland-client.so` runtime dep)
- [x] `makedepends`: 空。**意図通り** (下記「依存方針」参照)
- [x] `secrets` 混入なし (`.git` 削除済、秘密鍵 / token / `.gpg` ファイル
      等なし)
- [x] AUR との diff: 純 fork (`# Maintainer:` → `# Contributor:` +
      fork コメント追記、`package()` の trailing space 除去のみ)

## 依存方針 (AUR との意図的 diff)

`makedepends` が空なのは AUR PKGBUILD どおりで、意図的に正しい:

- `wayland-scanner` (Makefile が XML → C/H 生成に使用) と
  `wayland-client.h` (build 時 header) は両方とも Arch `wayland` package
  が提供する (`pacman -Qo /usr/bin/wayland-scanner` /
  `pacman -Qo /usr/include/wayland-client.h` で実機確認済み)。`wayland`
  は既に `depends` にあるため makedepends への重複記載は不要
- protocol XML (`ext-foreign-toplevel-list-v1.xml` /
  `wlr-foreign-toplevel-management-unstable-v1.xml`) は tarball に同梱
  されており `wayland-protocols` package は不要
- gcc / make は `base-devel` group で担保 (repo の build host 前提)

## 機能要件との突き合わせ

`lswt.1` によると情報取得の protocol 優先順位は ext-foreign-toplevel-list-v1
(preferred) → wlr-foreign-toplevel-management-unstable-v1 の順で、
「As of now, only the ext- protocol supports unique toplevel IDs」との記載
あり。2.0.0 でこの要件 (unique identifier 取得) を満たす。1.x 系は ext-
protocol 未対応のため **downgrade 不可**。

出力形式: `-j`/`--json` (JSON) または `-c <fmt>`/`--custom <fmt>` (CSV、
`t`=title, `a`=app-id, `i`=unique identifier)。

## 結論

**approve** — build host で `bin/build-all lswt` により build + sign +
repo db 追加可。

## 更新方針

upstream の新 tag (v2.0.1 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認 (AUR が追随していれば流用)
2. 本 dir の PKGBUILD を差し替え
3. sha256 を独立再計算 (`curl -fsSL <url> | sha256sum` で照合)。sr.ht
   archive tarball の生成方式変化リスクがあるため必須
4. `nvchecker -c nvchecker.toml` の `[lswt]` section が新 tag を返すか確認
5. REVIEW.md に確認日 + 結論 update

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
