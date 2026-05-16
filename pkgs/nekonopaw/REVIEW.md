# nekonopaw review

## 状態

**review 済み、approve** (2026-05-16、v0.1.2 bump: SPA_POD_OBJECT_FOREACH implicit declaration fix + GitHub Actions CI 追加)

直前の review:
- 2026-05-16, v0.1.1 cgo flag fix (本 repo PKGBUILD SHA `7d0202de007c11c4f686444962b2f1b00068f258`) — `CGO_*_ALLOW='.*'` 追加
- 2026-05-16, v0.1.1: LICENSE 追加 + license=MIT に bump
- 2026-05-16, v0.1.0 (本 repo PKGBUILD SHA `2a09fb89252b8acf1d95ae53f6296de37658e7cf`)
  — initial release、license=unknown で approve (Claude bot 軽微指摘あり)

## 2026-05-16 v0.1.2 bump (build error fix + CI)

build host (Arch、libpipewire 最新) で v0.1.1 build 中に:

```
pw_glue.c:113:9: error: implicit declaration of function 'SPA_POD_OBJECT_FOREACH'
```

`SPA_POD_OBJECT_FOREACH` macro は `spa/pod/iter.h` で定義されているが、
ayaka (Debian、libpipewire-0.3-dev 1.4.2) では他 header からの transitive
include で間接的に解決されていた。新 version では transitive が外れた様子。

upstream v0.1.2 (= PR #2 merge) で:
- `pw_glue.c` に `#include <spa/pod/iter.h>` を明示追加
- `.github/workflows/build.yml` で GitHub Actions CI を追加 (ubuntu-latest
  + libpipewire-0.3-dev で毎 push & pull_request に go vet/build/test を
  実行、再発防止)

source 内容変更のため新 release v0.1.2 を切って bump:
- pkgver: 0.1.1 → 0.1.2
- sha256: 6b94b017... → 3de929b8... (= 新 tarball、curl + sha256sum で実測一致)
- pkgrel=1 のまま (= 過去の v0.1.1 もまだ一度も配布されていないため bump 不要)
- license=MIT は維持 (LICENSE は v0.1.1 で追加済 → v0.1.2 も継承)

## 2026-05-16 cgo flag fix (v0.1.1 pkgrel 据置)

build host で v0.1.1 build を回したら以下で fail:

```
github.com/nananek/nekonopaw/internal/pw: invalid flag in pkg-config --cflags: -fno-strict-overflow
```

Go cgo は CGO_CFLAGS 経由で渡された flag を security policy (allowlist)
で validate し、Arch makepkg.conf default の `-fno-strict-overflow` を
unsafe と判定して reject していた。PKGBUILD の build() に
`CGO_{CFLAGS,CPPFLAGS,CXXFLAGS,LDFLAGS}_ALLOW='.*'` を追加して全許可。

評価:
- 本 package は自家 upstream (= 第三者 AUR fork ではない)
- CGO_* は makepkg.conf 由来の信頼できるシステム変数のみ、外部入力由来の
  unsafe flag 注入経路なし
- cgo 部分は `libpipewire-0.3.so` binding 限定、攻撃面が狭い
- `GOFLAGS` の `buildmode=pie -trimpath -ldflags=-linkmode=external
  -mod=readonly` は維持、hardened build に影響なし

source / sha256 / pkgver は変わらないので `pkgrel` は据置 (= まだ一度も
build 成功 / artifact 配布されていないため `pkgrel=1` のまま、初配布
artifact が `0.1.1-1` になる)。

ansible-nekonodesk owner (= nananek) 自身の upstream を pacman で配布する
ための PKGBUILD。AUR fork ではなく self-authored package。

## 用途

PipeWire のアプリ別 / デバイス別 出力音量ミキサーを web UI で操作する
小さい daemon。Tailscale 内 IP に bind して別端末 (iPhone/Mac/iPad 等)
から操作する想定。

旧実装は ansible-nekonodesk の roles/nekonopaw に submodule として埋め込み、
target machine 上で `go build` していたが、nekono-ansible-run の git
archive 経路で submodule の source が含まれず空 dir のまま fail していた。
pacman 配布に切替えることで `pacman -S nekonopaw` 1 発で binary が入る。
systemd-user service の deploy + Tailscale IP の bind は引き続き
ansible 側 (roles/nekonopaw) の責務。

## Source

- Upstream: https://github.com/nananek/nekonopaw
  - 主開発者 nananek (= ansible-nekonodesk owner)
  - tag `v0.1.2` (= merge commit `2365cd2`、`v0.1.1` + SPA include fix + GitHub Actions CI)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekonopaw/archive/refs/tags/v0.1.2.tar.gz`
  - 自家 upstream、typosquat 検討不要
- [x] `sha256sums` 独立検証
  - 実測: `3de929b8517dfc40705cc7dbc948ff506f8b282e480d3a3e00833eeca1b50788`
    (= `curl -fsSL <url> | sha256sum` で tarball から計算)
  - PKGBUILD 値と一致
- [⚠] tag `v0.1.2` は **GPG verified: false** — annotated tag だが GPG sign
      無し (= author の commit signing policy では tag は未 sign 運用)。
      tarball sha256 pin + GitHub release CI で integrity 確保
- [x] `build()`: cgo で libpipewire を link する標準 Go build。
  - CGO_* / GOFLAGS は Arch 標準 (buildmode=pie + trimpath + ldflags=
    -linkmode=external + mod=readonly + modcacherw)
  - `go build -o nekonopaw .` のみ、network fetch / eval なし
- [x] `package()`: binary を /usr/bin/nekonopaw に install、README を
      /usr/share/doc/nekonopaw/ に、LICENSE を /usr/share/licenses/nekonopaw/
      に (= Arch convention)。標準。
- [x] `depends`: `pipewire` — cgo で libpipewire-0.3.so を link、Arch では
      pipewire package に dev header も同梱
- [x] `makedepends`: `go` のみ (base-devel は makepkg 暗黙)
- [x] `license`: `MIT` — v0.1.1 で upstream に LICENSE 追加済 (= PR #1 merge)。
      SPDX `MIT`、`/usr/share/licenses/nekonopaw/LICENSE` に install。
- [x] `secrets` 混入なし

## 結論

**approve** — そのまま build host で `bin/build-all nekonopaw` で build +
sign + repo db 追加可。ansible 側 (roles/nekonopaw) は別 PR で `pacman -S
nekonopaw` 経路に refactor 予定。

## 更新方針

upstream で新 release tag (v0.2.0 等) が出たら:
1. PKGBUILD の pkgver を更新
2. tarball を curl + sha256sum で実検証、本 PKGBUILD の sha256sums と
   .SRCINFO を差し替え
3. license が SPDX 別値に変わったら license= も更新
4. REVIEW.md に確認日 + 結論 update
