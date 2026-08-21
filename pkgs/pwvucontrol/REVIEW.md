# pwvucontrol review

## 状態

**review 済み、approve** (2026-05-16、純 fork、改変なし)

AUR の `pwvucontrol` PKGBUILD (pkgver=0.5.2, pkgrel=1) を **純 fork**。
唯一の diff は `# Maintainer:` → `# Contributor:` の置換 + fork 説明
コメント追加のみ (= 関数本体・依存・ハッシュは無改変)。

## 用途

PipeWire 用 GUI ミキサー (`pavucontrol` の PipeWire ネイティブ版、Rust
+ GTK4 + libadwaita 製)。

Arch 公式 (core/extra) には無く AUR のみ存在するため nekono-pacman-
distribution に取り込む。当面 `ansible-nekonodesk` の Arch native apps
block から drop 済み (= build 完了後に Self-hosted block へ移行)。

## Source

- AUR: https://aur.archlinux.org/packages/pwvucontrol
  - maintainer: Moabeat <moabeat@berlin.de>
- Upstream: https://github.com/saivert/pwvucontrol
  - 主開発者 saivert、Rust + GTK4 + libadwaita 製
  - AUR `url` フィールドと一致 (typosquat / なりすましリスクなし)

## 検証結果

- [x] `source` URL = `github.com/saivert/pwvucontrol/archive/fd5c9d1e5458b625d45e4159634b5b072f66cde1/...`
  - mutable な upstream tag `0.5.3` ではなく、再リリース後の commit を固定
  - commit `fd5c9d1e5458b625d45e4159634b5b072f66cde1` は upstream maintainer
    saivert による `Update cargo lock`。実際の SHA は下記履歴に記録
- [x] `b2sums` (BLAKE2b-512) 独立検証
  - 実測: `a00b69ae226be9d4103ebf0de0a2566e3a47bb46b750b11d5a24057078be005876edc4095121393a5e4c4c6d47966a6c491803c4e08bf3d6b4c21314805d9492`
    (= commit 固定 tarball を `b2sum` で計算)
  - PKGBUILD 値と一致
  - BLAKE2 は暗号学的に強固で破綻なし (pam_pkcs11 #14 の md5 → sha256
    のような置換は不要)
- [⚠] upstream は release 後に `0.5.3` tag を別 commit へ移動しており、
      旧 tag の `Cargo.lock` / `build-aux/cargo-sources.json` と現在 tag の
      lockfile が異なる。現在の commit は同一 maintainer による crate lock 更新で、
      build script や runtime code の追加は確認できないが、commit signature は未検証。
- [x] `build()`: `arch-meson --reconfigure && meson compile` — meson 標準、
      network fetch / eval / 動的コマンド構築なし
- [x] `check()`: `meson test --print-errorlogs` — checkdepends に
      `appstream-glib` 明示済み、妥当
- [x] `package()`: `meson install --no-rebuild --destdir` — 標準
- [x] `depends`: `glib2 gtk4 libadwaita wireplumber` — Rust + GTK4 +
      libadwaita 製 PipeWire mixer として妥当
- [x] `makedepends`: `rust clang meson blueprint-compiler` — Blueprint UI
      記述言語 + meson + Rust の組合せ、upstream の `meson.build` と整合
- [x] `secrets` 混入なし (`.git` 削除済、秘密鍵 / token / `.gpg` ファイル
      等なし)
- [x] AUR との diff: 純 fork (= 機能部分は完全一致、`# Maintainer:` →
      `# Contributor:` + コメント追記のみ)

## 結論

**approve with source pin** — upstream tag の差し替えを受け、commit 固定 source と
再計算した `b2sums` を採用する。build host で `bin/build-all pwvucontrol` を実行し、
build + sign + repo db 追加可。

これにより:
- `apt_packages/vars/Archlinux.yml` の Self-hosted block に `pwvucontrol`
  を追加し、Arch native apps block から再復活させる (= kirigiri / 将来の
  ayaka Arch 化版で `pacman -S pwvucontrol` 可能)

## 更新方針

upstream の新 release (0.5.3 等) が出たら:
1. AUR PKGBUILD の pkgver / b2sums を確認
2. 本 dir の PKGBUILD を差し替え
3. b2sum を独立再計算 (= commit 固定 tarball を `b2sum` で照合)
4. upstream commit author を GitHub API で確認 (= saivert が変わってないか)
5. REVIEW.md に確認日 + 結論 update

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-23 | 0.5.2-2 | `f574a11` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): meson 1.11.1-1 → 1.11.1-3 |
| 2026-05-25 | 0.5.2-3 | `745841a` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libadwaita 1:1.9.0-1 → 1:1.9.1-1 |
| 2026-05-31 | 0.5.2-4 | `ed39ad1` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): clang 22.1.5-1 → 22.1.6-1 |
| 2026-06-03 | 0.5.2-5 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): rust 1:1.95.0-1 → 1:1.96.0-1 |
| 2026-06-21 | 0.5.2-6 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): wireplumber 0.5.14-1 → 0.5.15-1 |
| 2026-06-27 | 0.5.2-7 | (this PR) | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): glib2 2.88.1-1 → 2.88.2-1 |
| 2026-06-29 | 0.5.2-8 | bot PR #307 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): libadwaita 1:1.9.1-1 → 1:1.9.2-1 |
| 2026-07-01 | 0.5.2-9 | bot PR #316 | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): rust 1:1.96.0-1 → 1:1.96.1-1 |
| 2026-07-10 | 0.5.3-1 | upstream tag `0.5.3` | upstream tag `0.5.3` | pkgver bump 0.5.2 → 0.5.3 (Issue #375, safe-to-bump)。system color scheme 追従 fix (`gtk::init()` → `adw::init()`)。build()/package()・depends 無変更、b2sums 差し替えのみ。makepkg --verifysource で b2 検証済み |
| 2026-08-21 | 0.5.3-2 | (this PR) | `fd5c9d1e5458b625d45e4159634b5b072f66cde1` | upstream が `0.5.3` tag を再配置。commit 固定 source、b2sums 更新、pkgrel 1 → 2。Cargo lock の crate 更新のみ、build()/package()/depends 無変更 |
