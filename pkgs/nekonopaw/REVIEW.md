# nekonopaw review

## 状態

**review 済み、approve** (2026-05-16、in-house first release v0.1.0)

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
  - tag `v0.1.0` (= commit `bd7a3f6` time of cut)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekonopaw/archive/refs/tags/v0.1.0.tar.gz`
  - 自家 upstream、typosquat 検討不要
- [x] `sha256sums` 独立検証
  - 実測: `b51c3e48f23dbefd544a638a2f70d080261a0f62b1eba882286f5a3644b366a4`
    (= `curl -fsSL <url> | sha256sum` で tarball から計算)
  - PKGBUILD 値と一致
- [⚠] tag `v0.1.0` は **GPG verified: false** — annotated tag だが GPG sign
      無し (= author の commit signing policy では tag は未 sign 運用)。
      tarball sha256 pin + GitHub release CI で integrity 確保
- [x] `build()`: cgo で libpipewire を link する標準 Go build。
  - CGO_* / GOFLAGS は Arch 標準 (buildmode=pie + trimpath + ldflags=
    -linkmode=external + mod=readonly + modcacherw)
  - `go build -o nekonopaw .` のみ、network fetch / eval なし
- [x] `package()`: binary を /usr/bin/nekonopaw に install、README を
      /usr/share/doc/nekonopaw/ に。標準。
- [x] `depends`: `pipewire` — cgo で libpipewire-0.3.so を link、Arch では
      pipewire package に dev header も同梱
- [x] `makedepends`: `go` のみ (base-devel は makepkg 暗黙)
- [⚠] `license`: **unknown** — upstream に LICENSE ファイル無し。
      author が後日 LICENSE を repo に追加した上で、本 PKGBUILD の
      license= フィールドも併せて update する (= 別 PR)。
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
3. LICENSE が追加されていたら license= も更新
4. REVIEW.md に確認日 + 結論 update
