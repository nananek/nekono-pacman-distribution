# python-httpx-sse review

## 状態

**review 済み、 approve** (2026-05-19、 初回 add)

AUR の `python-httpx-sse` PKGBUILD (pkgver=0.4.3, pkgrel=1) を fork。 純 fork
(= 機能変化なし、 改変は Maintainer 行 + nekono 説明コメント + check() 削除のみ)。

## 用途

`python-fastmcp` 依存 chain (= `python-fastmcp` → `python-mcp` → `python-httpx-sse`)
で必要な leaf 依存。 Arch 公式 repo に不在で AUR にしか無いため [nekono] に
fork 投入する。

ayaka 上の自家 MCP server を pacman 経由 install できるようにする取り組みの一環。

## Source

- AUR: https://aur.archlinux.org/packages/python-httpx-sse
  - Maintainer: envolution
  - Contributor: Carl Smedstad <carl.smedstad at protonmail dot com>
- Upstream: https://github.com/florimondmanca/httpx-sse
  - florimondmanca (Florimond Manca)、 MIT
  - tag `0.4.3` (= prefix なし)、 commit `e8fcd9e159066185963ffb9fa29efb8ba2ca84bf`

## 検証結果

- [x] `source` URL = `github.com/florimondmanca/httpx-sse/archive/refs/tags/0.4.3.tar.gz`
  - upstream 公式 GitHub Release archive、 typosquat なし
- [x] `sha256sums` 独立検証
  - 実測 (build host 上 `makepkg --verifysource`): `926e88187fa45117882f287955dbcb53c1c4ec36148893f0c514f021b7f3b000`
  - AUR PKGBUILD と一致
- [x] `build()`: `python -m build --wheel --no-isolation`、 ネットワーク取得なし
- [x] `package()`: `python -m installer --destdir` + `install -Dm644 LICENSE` のみ、 shell injection / curl / wget なし
- [x] `depends`: python / python-httpx — 全て Arch 公式 extra
- [x] `makedepends`: python-build / python-installer / python-setuptools-scm / python-wheel — 全て Arch 公式 extra
- [x] `secrets` 混入なし

## AUR との意図的差分

| 変更 | 理由 |
|---|---|
| `# Maintainer:` 行を Nekono に書換 + AUR の Maintainer / Contributor は `# Upstream AUR Maintainer:` / `# Upstream AUR Contributor:` として保持 | 当 PKGBUILD は Nekono が責任を持つため。 AUR の信頼の起点を記録するため上流情報を残す |
| `check()` 関数を削除 | AUR の checkdepends に `python-sse-starlette` が含まれるが、 これは AUR 限定で [nekono] では PR #3 で別 fork 投入予定。 [nekono] build host は `bin/build-all` で `makepkg` を `--check` なしで走らせる (= bin/build-all は default で check skip) ので、 check() があっても呼ばれないが、 PKGBUILD の completeness 観点で削除 + checkdepends も連動削除 (= 派生して `.deps.lock` の MISSING 行も減る) |

## 結論

**approve** — build host で `bin/build-all python-httpx-sse` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release (= 0.4.4 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認 (= `git pull` of AUR fork or `curl https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=python-httpx-sse`)
2. 本 dir の PKGBUILD + .SRCINFO を差し替え
3. sha256 を独立再計算 (= `curl + sha256sum`)
4. depends / makedepends が変化していれば `.deps.lock` を更新
5. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-19 | 0.4.3 | `1eb2fc83d93358a76287d87076586a411910dd53` | `e8fcd9e159066185963ffb9fa29efb8ba2ca84bf` | 初回 add、 純 fork (= check() 削除のみ) |
| 2026-06-21 | 0.4.3-2 | `(this PR)` | — (pkgrel bump のみ) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
