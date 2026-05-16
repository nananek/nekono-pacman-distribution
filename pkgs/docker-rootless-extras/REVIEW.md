# docker-rootless-extras review

## 状態

**review 済み、approve** (2026-05-16、純 fork、改変なし)

AUR の `docker-rootless-extras` PKGBUILD (pkgver=29.4.3, pkgrel=1) を
**純 fork**。唯一の diff は `# Maintainer:` → `# Contributor:` の置換 +
fork 説明コメント追加のみ (= 関数本体・依存・ハッシュ・install hook は
無改変)。

## 用途

Docker rootless mode 用 extras。systemd-user 経路で非 root daemon を立てる
ための前提部品 (`dockerd-rootless.sh` + systemd-user `docker.service`
+ `docker.socket` + sysctl `99-docker-rootless.conf`) を ship。

Arch 公式 (core/extra) には無く AUR のみ存在する。`ansible-nekonodesk`
の `roles/docker` が kts_sz uid で rootless daemon を運用するために必要。

**注意**: 本 PKGBUILD は `dockerd-rootless-setuptool.sh` を /usr/bin/ に
install **しない** (= awk で systemd unit のみ抽出して捨てる、AUR
maintainer の意図的設計)。Debian apt の同名 package とは flow が異なる
ため、`roles/docker` は OS-dispatch で扱う必要あり (= 別 issue)。

## Source

- AUR: https://aur.archlinux.org/packages/docker-rootless-extras
  - maintainer: Ľubomír 'the-k' Kučera
  - contributors: Hugo Osvaldo Barrera / PastLeo / koba1t
- Upstream: https://github.com/moby/moby (= Docker / Moby project)
  - 2 つの shell script を `docker-v29.4.3` annotated tag から取得
  - tag commit: `05b03cea61ee67d3787a27910c37681d04f2cbb0`

## 検証結果

- [x] `source` URL = `raw.githubusercontent.com/moby/moby/docker-v29.4.3/...`
  - Docker 公式 upstream repo、typosquat / なりすましリスクなし
- [x] `sha256sums` 4 件すべて独立検証 (= curl + sha256sum / sha256sum)
  - `dockerd-rootless.sh`
    - 実測: `5cdcd9512da29704c8615de33390cfe950b7b720f8b52b215fbbaa7646d693d3`
    - PKGBUILD 値と一致
  - `dockerd-rootless-setuptool.sh`
    - 実測: `e2d0b86f145323f9597dd1d587bbc8a5b0524dd92f34d03b49fb455768ab65e6`
    - PKGBUILD 値と一致
  - `docker.socket` (ローカル AUR snapshot)
    - 実測: `d8695293e5d4a814763f13e1d36ed37273040666b4b91363d6c33171df8934c7`
    - PKGBUILD 値と一致
  - `99-docker-rootless.conf` (ローカル AUR snapshot)
    - 実測: `d0d790d4c3d887b10b2b155b83a58a44980b9fa638f8c0f1faec0739dc0ef473`
    - PKGBUILD 値と一致
- [x] `package()`:
  - `install -Dm755 dockerd-rootless.sh → /usr/bin/dockerd-rootless.sh`
  - `install -Dm644 docker.socket → /usr/lib/systemd/user/docker.socket`
  - `install -Dm644 99-docker-rootless.conf → /usr/lib/sysctl.d/`
  - `awk '/Unit/,/EOT/' setuptool.sh | head + sed` で systemd unit を抽出 →
    `/usr/lib/systemd/user/docker.service`
  - すべて静的、network fetch / eval / 動的コマンド構築なし
  - `awk` / `sed` / `head` は coreutils + base 同梱、makedepends 不要
- [x] `depends`: `bash docker rootlesskit slirp4netns` — rootless 動作に
      必須の依存すべて揃っている
- [x] `optdepends`: `fuse-overlayfs: overlayfs support` — overlayfs
      バックエンド利用時のみ。妥当
- [x] `install` hook (`docker-rootless-extras.install`):
  - `sysctl --system` で `99-docker-rootless.conf` の `kernel.
    unprivileged_userns_clone=1` を即時反映
    - Linux ≥5.11 は default 有効、実質的に変化なし
  - subuid/subgid / `systemctl --user enable docker.socket` 等の
    案内 echo (= ansible 側で実 setup を行う)
- [x] `provides` / `conflicts`: `docker-rootless` / `docker-rootless-extras`
      / `docker-rootless-extras-bin` を network 互換に list
- [x] `secrets` 混入なし (`.git` 削除済、秘密鍵 / token / `.gpg` ファイル
      等なし)
- [x] AUR との diff: 純 fork (= 機能部分完全一致、`# Maintainer:` →
      `# Contributor:` + コメント追記のみ)

## セキュリティ所見

- `awk '/Unit/,/EOT/'` による service file 抽出は AUR オリジナル実装。
  source は sha256 pin されているため supply-chain リスクなし。ただし
  `setuptool.sh` の upstream 実装が変わった場合にサイレントに壊れる
  可能性あり (PKGBUILD 内 TODO コメント済み)。
- `kernel.unprivileged_userns_clone=1` sysctl は rootless Docker 必須。
  Linux ≥5.11 default 有効のため実質的に変化なし。意図した設定、問題なし。

## 結論

**approve** — そのまま build host で `bin/build-all docker-rootless-extras`
で build + sign + repo db 追加可。

build 完了後の予定:
- `roles/docker/tasks/Archlinux.yml` を Arch-native rootless flow に refactor
  (= subuid/subgid 配備 + `systemctl --user enable --now docker.socket`、
  setuptool.sh 起動経路は使わない)
- `apt_packages/vars/Archlinux.yml` の Self-hosted block (もしくは
  docker role 内の install list) に `docker-rootless-extras` を追加

## 更新方針

upstream の新 release (docker 29.5.x 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. 本 dir の PKGBUILD + .SRCINFO を差し替え
3. 4 sources の sha256 を独立再計算 (= curl + sha256sum で照合)
4. setuptool.sh の `awk '/Unit/,/EOT/'` 抽出が依然成立するか確認 (= 抽出
   結果の docker.service 内容を目視確認)
5. REVIEW.md に確認日 + 結論 update
