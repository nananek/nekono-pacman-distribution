# docker-rootless-extras review

## 状態

**review 済み、approve** (最新: 2026-05-21 / 29.5.2)

AUR の `docker-rootless-extras` PKGBUILD を **純 fork**。diff は
`# Maintainer:` → `# Contributor:` の置換 + fork 説明コメント追加、および
v29.5.0 bump 時に `optdepends` に `gvisor-tap-vsock` を追加 (upstream で
slirp4netns の代替 driver として導入されたため)。関数本体・install hook
は無改変。各 release の review 履歴は本ファイル末尾の「更新履歴」 section
参照。

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
  - 2 つの shell script を `docker-v29.5.0` annotated tag から取得
  - tag commit: `9cfd86615a90644b86d33978ee0df3702e5021b0`
  - release author: `vvoland` (既知 moby maintainer)、prerelease: false

## 検証結果

- [x] `source` URL = `raw.githubusercontent.com/moby/moby/docker-v29.5.0/...`
  - Docker 公式 upstream repo、typosquat / なりすましリスクなし
- [x] `sha256sums` 4 件すべて独立検証 (Issue #24 事前調査 + PR #33 の
      claude-review.yml で再計算済み、両者一致)
  - `dockerd-rootless.sh` (29.5.0): `904c9b9e35f6927c0a5e65afb4d35b6bc9eb1278c878044501281fc728c9be46`
  - `dockerd-rootless-setuptool.sh` (29.5.0): `1c9f0dc93ebb3d75255254ec760d26a912affba7f329ab8abffe8e25eb0b3f94`
  - `docker.socket` (ローカル AUR snapshot、変更なし): `d8695293e5d4a814763f13e1d36ed37273040666b4b91363d6c33171df8934c7`
  - `99-docker-rootless.conf` (ローカル AUR snapshot、変更なし): `d0d790d4c3d887b10b2b155b83a58a44980b9fa638f8c0f1faec0739dc0ef473`
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
- [x] `optdepends`:
  - `fuse-overlayfs: overlayfs support` — overlayfs バックエンド利用時のみ
  - `gvisor-tap-vsock: alternative network driver (used when slirp4netns
    is unavailable)` — upstream v29.5.0 で追加。**Arch 公式 repo に無く
    AUR のみ存在、現状 orphaned (maintainer: None) なので、optdepends として
    install したいユーザは AUR / foreign repo 対応が必要**。本 pkg の機能
    上は optdepends なので非必須、build / 配布には影響なし。
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

## 依存方針 (AUR との意図的 diff)

AUR の現行 PKGBUILD (v29.5.0) と本 repo の PKGBUILD で `depends` /
`optdepends` の振り方に意図的な差異がある:

| | AUR v29.5.0 | 本 repo |
|---|---|---|
| `slirp4netns` | `optdepends` (recommended) | **`depends` (必須)** |
| `gvisor-tap-vsock` | (記載なし) | `optdepends` |

**本 repo が `slirp4netns` を depends に keep する理由**:

AUR は v29.5.0 で `gvisor-tap-vsock` を主推奨 network driver に切替え、
`slirp4netns` を optional に降格した。しかし `gvisor-tap-vsock` は
**Arch 公式 repo に無く、AUR でも現状 orphaned (maintainer: None)**
であり、安定供給が保証されていない。本 repo はビルド成果物を Tailscale
経由で配信する self-host 運用なので、依存先の安定性が AUR より重要。

そこで、本 repo は:
- `slirp4netns` を **`depends` のまま維持** (= 既存 ayaka 環境で動作保証)
- `gvisor-tap-vsock` を **新規 `optdepends` として追加** (= 使いたい
  ユーザは AUR / foreign repo で対応)

将来 `gvisor-tap-vsock` が Arch 公式 repo に入った時点で再評価する。

(claude-review.yml で「AUR との差異」として指摘された際は、本 section を
参照すること。)

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
5. REVIEW.md の「検証結果」 section を新 sha256 で上書き + 「更新履歴」に
   1 行追記 (review 日付 + PKGBUILD repo SHA + upstream tag commit SHA)

## 更新履歴

- **2026-07-17 / 29.6.2** — approve (Issue #410)。Docker Engine v29.6.2 (release 2026-07-16、by `vvoland`)
  への sync。upstream tag commit: `3d80467678f6e36325fa9ae3dd486fe91e5652e3`。
  **security release**: BuildKit の脆弱性 5 件を修正 (CVE-2026-15788/15789/15791/15792/15793)、
  containerd v2.2.6 / Go 1.26.5 / RootlessKit v3.0.2 (docker 内部 packaging note、Arch 公式
  `rootlesskit` package の追従とは別軸)。CVE 群は BuildKit 側の fix であり、本 pkg が取得する
  rootless スクリプト群 (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`) の中身は v29.6.1 と
  **完全同一** (= sha256 4 値すべて不変、byte 単位で一致確認)。PKGBUILD 改変は `pkgver=29.6.2` の 1 行のみ。
  sha256 独立再検証済み (moby/moby docker-v29.6.2 の raw fetch + 実測)。Closes #410。

- **2026-07-03 / 29.6.1** — approve (Issue #331)。Docker Engine v29.6.1 (release 2026-06-26、by `vvoland`)
  への sync (途中の v29.6.0 も含めまとめて追従)。upstream tag commit: `8ec5ab355a34b2a0e2b3238d67bdefe77fefa982`。
  **security release**: 悪意ある image による /etc/passwd・/etc/group パーサの OOM (GHSA-mjcv-p78q-w5fw ほか)、
  custom frontend による BuildKit Seccomp/AppArmor bypass (GHSA-7236-3392-c5c6) を修正、containerd v2.2.5 /
  BuildKit v0.31.1。ただし CVE 群は docker 本体 (daemon/buildkit) 側の fix であり、本 pkg が取得する
  rootless スクリプト群 (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`) の中身は v29.5.3 と
  **完全同一** (= sha256 4 値すべて不変、byte 単位で一致確認)。PKGBUILD 改変は `pkgver=29.6.1` の 1 行のみ。
  sha256 独立再検証済み (moby/moby docker-v29.6.1 の raw fetch + 実測)。Closes #331。

- **2026-06-04 / 29.5.3** — approve。Docker Engine v29.5.3 (release 2026-06-03、by `vvoland`)
  への sync。upstream tag commit: `285b47192d4b2f183aba5dd360a92cd52d723004`。
  パッチリリース (rootless UDP / plugin 修正、RootlessKit v3.0.1)。
  rootless スクリプト群 (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`)
  の中身は v29.5.2 と **完全同一** (= sha256 不変)。PKGBUILD 改変は `pkgver=29.5.3` の 1 行のみ。
  Issue #160 (upstream-version-issue.yml 調査) で同一性確認済み。Closes #160。

- **2026-05-21 / 29.5.2** — approve。Docker Engine v29.5.2 (release 2026-05-20、by `vvoland`)
  への sync。upstream tag commit: `b5c3467ddde0169d80aade06b47994dca8137fe6`。
  パッチリリース (docker cp バグ修正 1 件のみ)。rootless スクリプト群
  (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`) の中身は v29.5.1 と
  **完全同一** (= sha256 不変)。PKGBUILD 改変は `pkgver=29.5.2` の 1 行のみ。
  Issue #91 (upstream-version-issue.yml 調査) で同一性確認済み。Closes #91。

- **2026-05-19 / 29.5.1** — approve。Docker Engine v29.5.1 (release
  2026-05-18, by `vvoland`) への sync。 upstream tag commit:
  `dd24a3adc1db4c762fb1b26b35c08ffd936f2d8f`。 セキュリティパッチリリースで
  `docker cp` 系 CVE 3 件 (CVE-2026-41567 / 41568 / 42306) を修正、 ただし
  rootless スクリプト群 (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`)
  の中身は v29.5.0 と **完全同一** (= sha256 不変)。 PKGBUILD 改変は
  `pkgver=29.5.1` の 1 行のみ。 Issue #37 (upstream-version-issue.yml の
  事前調査) で同一性 + supply-chain 確認済み。 release author は v29.5.0 と
  同じ `vvoland` (Paweł Gronowski, Docker 社員)。

- **2026-05-17 / 29.5.0** — approve。Docker Engine v29.5.0 (release
  2026-05-14, by `vvoland`) への sync。upstream tag commit:
  `9cfd86615a90644b86d33978ee0df3702e5021b0`、本 repo PKGBUILD SHA:
  `2cdf394535847cca50deb63b42ff849ac2f62996`。
  主な upstream 変更:
  - rootlesskit `--detach-netns` モードのデフォルト有効化
  - `gvisor-tap-vsock` を新 network driver fallback として追加 (= 本
    PKGBUILD の optdepends にも追記、ただし AUR orphan なので注意)
  - sysctl 呼び出しが netns 内で実行されるよう変更 (detach-netns 有効時)

  PKGBUILD の `package()` 内 `awk '/Unit/,/EOT/'` 抽出範囲は setuptool.sh
  の変更箇所 (Unit セクション外) に非該当、引き続き動作する想定。**build
  host で makepkg 実行時に生成 `docker.service` の内容を目視確認推奨**。

  AUR v29.5.0 との意図的 diff (`slirp4netns` を depends に維持、
  `gvisor-tap-vsock` は新規 optdepends に追加) は「依存方針」 section
  参照。
- **2026-05-16 / 29.4.3** — approve (初回 fork + review、純 fork で改変
  なし)。upstream tag commit: `05b03cea61ee67d3787a27910c37681d04f2cbb0`。
  当時の sha256 実測値:
  - `dockerd-rootless.sh`: `5cdcd9512da29704c8615de33390cfe950b7b720f8b52b215fbbaa7646d693d3`
  - `dockerd-rootless-setuptool.sh`: `e2d0b86f145323f9597dd1d587bbc8a5b0524dd92f34d03b49fb455768ab65e6`
