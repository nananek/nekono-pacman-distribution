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

- [x] `source` URL = `github.com/saivert/pwvucontrol/archive/refs/tags/0.5.2.tar.gz`
  - upstream tag `0.5.2` (saivert/pwvucontrol)、AUR と同じ正規 release
- [x] `b2sums` (BLAKE2b-512) 独立検証
  - 実測: `fb749511f886a0481edc5e6d8312241503d133724f316a76dfc1c3222c1b1874d4ed332fe847b358340cef8258ca04e56bd33ed2a72a713cacf8e06f992a7031`
    (= `curl -fsSL <url> | b2sum` で tarball から計算)
  - PKGBUILD 値と一致
  - BLAKE2 は暗号学的に強固で破綻なし (pam_pkcs11 #14 の md5 → sha256
    のような置換は不要)
- [⚠] Tag `0.5.2` の GPG 検証は未実施 (= saivert は commit signing を
      運用していない様子)。tarball b2sum pin + GitHub release CI で
      integrity 確保
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

**approve** — そのまま build host で `bin/build-all pwvucontrol` で build
+ sign + repo db 追加可。

これにより:
- `apt_packages/vars/Archlinux.yml` の Self-hosted block に `pwvucontrol`
  を追加し、Arch native apps block から再復活させる (= kirigiri / 将来の
  ayaka Arch 化版で `pacman -S pwvucontrol` 可能)

## 更新方針

upstream の新 release (0.5.3 等) が出たら:
1. AUR PKGBUILD の pkgver / b2sums を確認
2. 本 dir の PKGBUILD を差し替え
3. b2sum を独立再計算 (= `curl -fsSL <url> | b2sum` で照合)
4. tag commit author を GitHub API で確認 (= saivert が変わってないか)
5. REVIEW.md に確認日 + 結論 update
