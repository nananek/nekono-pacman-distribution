# docker-rootless-extras review

## 状態

**review 済み、approve** (最新: 2026-11-25 / 29.10.1)

AUR の `docker-rootless-extras` 由来。package/install の主要処理は継承しつつ、
安定供給できる `slirp4netns` を hard dependency に維持し、利用可能な network
driver を optdepends に明示する依存方針を採用。AUR の shellcheck 用 no-op guards
はpackage動作に影響しないため未追従。各 release の review 履歴は本ファイル
末尾の「更新履歴」 section参照。

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
  - 2 つの shell script を `docker-v29.10.1` annotated tag から取得
  - tag object: `497e9c93a17abe303d417de5575e243dce5049db` (annotated + GPG 署名済み)
  - tag commit: `54cd494d0b39abfc7a958fcc876f706271f8b909`
  - tagger: Sebastiaan van Stijn `github@gone.nl` (= 既知 moby maintainer
    `thaJeztah`、Docker 社員)、GPG 署名 `verified: true` / `reason: valid`
  - release author: `thaJeztah`、prerelease: false
  - release publish: 2026-11-24T21:49:50Z
  - **注**: 29.10.0 までの tagger は `vvoland` (Paweł Gronowski) だったが、
    本 release は `thaJeztah` が tag を切っている。両者とも既知の moby
    maintainer であり、GPG 署名が `verified: true` で検証されているため
    supply-chain 上の懸念はなし (= domain / owner の変更なし、repo も
    `moby/moby` のまま)。

## 検証結果

- [x] `source` URL = `raw.githubusercontent.com/moby/moby/docker-v29.10.1/...`
  - Docker 公式 upstream repo、typosquat / なりすましリスクなし
  - 29.10.0 からの URL 変更は tag 部分 (`docker-v29.10.0` → `docker-v29.10.1`) のみ
- [x] `sha256sums` 4 件すべてを 2026-11-25 に独立検証 (`makepkg --verifysource`
      で 4/4 Passed、加えて raw URL / Contents API blob SHA を照合)
  - upstream 2 scripts は 29.10.0 タグの内容と **byte 単位で完全同一**
    (sha256 4 値すべて不変)。`makepkg --verifysource` が新 tag から
    download した実ファイルで 4/4 Passed
    - raw 取得の実測値と Contents API の
      `git/blobs/c5ea79a395729e5644d6b4c8635790c78609210f` が一致
  - `dockerd-rootless.sh` (29.10.1): `200203633806081a401e60aefdf68a8fa73fc7dc80aa854c52a69d47710a3488` (不変)
  - `dockerd-rootless-setuptool.sh` (29.10.1): `1c9f0dc93ebb3d75255254ec760d26a912affba7f329ab8abffe8e25eb0b3f94` (不変)
  - `docker.socket` (ローカル AUR snapshot、変更なし): `d8695293e5d4a814763f13e1d36ed37273040666b4b91363d6c33171df8934c7`
  - `99-docker-rootless.conf` (ローカル AUR snapshot、変更なし): `d0d790d4c3d887b10b2b155b83a58a44980b9fa638f8c0f1faec0739dc0ef473`
- [x] tag 範囲の差分 (`docker-v29.10.0...docker-v29.10.1`、281 commits / 30 files)
  - 変更 file は `daemon/` / `vendor/` / `integration/` / `go.mod` /
    `go.sum` / `cmd/dockerd/winresources/docker.ico` のみ。
    **`contrib/` の変更は 0 件** (= 本 pkg が ship する 2 script に無関係)
  - `vendor/github.com/moby/sys/user/*` の変更 = CVE-2026-135844
    (user/group/id-mapping ファイル parse 時の unbounded read) の fix
  - `daemon/containerfs_linux.go` の変更 = CVE-2026-135124
    (`docker cp` 中の read-only bind mount 書換) の fix
  - いずれも dockerd バイナリ側の修正で、Arch 公式 `docker` package が
    受け持つ。本 pkg (rootless script + systemd unit + sysctl) は影響なし
- [x] `package()` (29.10.0 から変更なし):
  - `install -Dm755 dockerd-rootless.sh → /usr/bin/dockerd-rootless.sh`
  - `install -Dm644 docker.socket → /usr/lib/systemd/user/docker.socket`
  - `install -Dm644 99-docker-rootless.conf → /usr/lib/sysctl.d/`
  - `awk '/Unit/,/EOT/' setuptool.sh | head + sed` で systemd unit を抽出 →
    `/usr/lib/systemd/user/docker.service`
  - すべて静的、network fetch / eval / 動的コマンド構築なし
  - `awk` / `sed` / `head` は coreutils + base 同梱、makedepends 不要
  - setuptool.sh が byte-identical のため `awk` 抽出結果
    (`[Unit]` / `[Service]` / `[Install]` の 3 section) も従来と同一
- [x] `depends`: `bash docker rootlesskit slirp4netns` — rootless 動作に
      必須の依存すべて揃っている
- [x] `optdepends`:
  - `fuse-overlayfs: overlayfs support` — overlayfs バックエンド利用時のみ
  - `gvisor-tap-vsock: alternative network driver (used when slirp4netns
    is unavailable)` — upstream v29.5.0 で追加。**Arch 公式 repo に無く
    AUR のみ存在、現状 orphaned (maintainer: None) なので、optdepends として
    install したいユーザは AUR / foreign repo 対応が必要**。本 pkg の機能
    上は optdepends なので非必須、build / 配布には影響なし。
  - `lxc`: experimental `lxc-user-nic` network driver
  - `passt`: experimental `pasta` network / `pesto` port driver。pesto は
    RootlessKit 3.1.0 + passt 2026_05_07.1afd4ed 以降、IPv4 only
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
- [x] AUR との diff: 29.10.1 時点の AUR PKGBUILD (pkgver=29.10.1、
      sha256 4 値は本 repo と一致) を取得して照合。package/install の主要処理は
      同一。依存方針とlint用 no-op guardsのみ意図的差分
      (`# Maintainer:` → `# Contributor:` + コメント追記含む)

## セキュリティ所見

- `awk '/Unit/,/EOT/'` による service file 抽出は AUR オリジナル実装。
  source は sha256 pin されているため supply-chain リスクなし。ただし
  `setuptool.sh` の upstream 実装が変わった場合にサイレントに壊れる
  可能性あり (PKGBUILD 内 TODO コメント済み)。
- `kernel.unprivileged_userns_clone=1` sysctl は rootless Docker 必須。
  Linux ≥5.11 default 有効のため実質的に変化なし。意図した設定、問題なし。
- 29.10.1 は **security release** (CVE-2026-135844 / CVE-2026-135124) だが、
  両 CVE とも dockerd バイナリ側 (= Arch 公式 `docker` package) の修正で、
  本 pkg が ship する成果物には含まれない。本 pkg 自身の供給内容に
  脆弱性の持ち込みはない。

## 依存方針 (AUR との意図的 diff)

AUR の現行 PKGBUILD (v29.10.1) と本 repo の PKGBUILD で `depends` /
`optdepends` の振り方に意図的な差異がある:

| | AUR v29.10.1 | 本 repo |
|---|---|---|
| `slirp4netns` | `optdepends` (recommended) | **`depends` (必須)** |
| `gvisor-tap-vsock` | (記載なし) | `optdepends` |
| `lxc` / `passt` | `optdepends` | `optdepends` |
| shell / version floor | `sh`、`docker>=1:29.5.0`、`rootlesskit>=3.0.0` | `bash`、`docker`、`rootlesskit` (Arch rolling の現行版を `.deps.lock` で監視) |
| shellcheck用 no-op guards | あり | なし (package動作に影響しないため) |

**本 repo が `slirp4netns` を depends に keep する理由**:

AUR は複数の network driver を optional とし、`slirp4netns` を recommended
としている。本 repo はビルド成果物を Tailscale 経由で配信する self-host
運用なので、既定経路が必ず解決することを優先する。

そこで、本 repo は:
- `slirp4netns` を **`depends` のまま維持** (= 既存 ayaka 環境で動作保証)
- `gvisor-tap-vsock` を **`optdepends` として維持** (= 使いたい
  ユーザは AUR / foreign repo で対応)
- AUR と同じ `lxc` / `passt` alternative も `optdepends` として明示

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

upstream の新 release (docker 29.10.x 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. 本 dir の PKGBUILD + .SRCINFO を差し替え
3. 4 sources の sha256 を独立再計算 (= curl + sha256sum で照合)
4. setuptool.sh の `awk '/Unit/,/EOT/'` 抽出が依然成立するか確認 (= 抽出
   結果の docker.service 内容を目視確認)
5. REVIEW.md の「検証結果」 section を新 sha256 で上書き + 「更新履歴」に
   1 行追記 (review 日付 + PKGBUILD repo SHA + upstream tag commit SHA)

## 更新履歴

- **2026-11-25 / 29.10.1** — approve (Issue #842)。Moby v29.10.1 (release 2026-11-24、by `thaJeztah`) へ sync。annotated tag object `497e9c93a17abe303d417de5575e243dce5049db` の target commit `54cd494d0b39abfc7a958fcc876f706271f8b909`、tagger `Sebastiaan van Stijn <github@gone.nl>` (= `thaJeztah`, Docker 社員、既知 moby maintainer、GPG 署名 `verified: true` / `reason: valid`)。29.10.0 までの tagger は `vvoland` だったが本 release は `thaJeztah` が担当。両者とも既知 maintainer かつ署名検証成立、repo (`moby/moby`) / domain の変更もないため supply-chain 懸念なし。release note は **security release** で、CVE-2026-135844 (moby/sys の user/group/id-mapping ファイル parse 時 unbounded read、GHSA-ffh6-8jcg-cjmw) / CVE-2026-135124 (実行中 container 内 process が `docker cp` 中に read-only bind mount を writable 化できる、GHSA-vfmq-r373-x89f) の 2 件に加え、`Expect: 100-continue` 付き build context upload の "http: invalid Read on closed Body" 修正 / `kernel.apparmor_restrict_unprivileged_unconfined=1` 環境 (最近の Ubuntu) で同 container 内 process が互いを signal/trace できない AppArmor default profile の修正 / `docker network rm` の "has active endpoints" 誤検知修正 / Swarm ネットワーク接続と endpoint 削除の race 修正 / BuildKit v0.35.1 / Windows icon 更新。tag 範囲の差分 (`docker-v29.10.0...docker-v29.10.1`、281 commits / 30 files) の変更 file は `daemon/` / `vendor/` / `integration/` / `go.mod` / `go.sum` / `cmd/dockerd/winresources/docker.ico` のみで **`contrib/` の変更は 0 件**。本 pkg が ship する 2 script (`dockerd-rootless.sh` / `dockerd-rootless-setuptool.sh`) は raw URL 実測 + Contents API blob SHA 照合で 29.10.0 と **byte-identical** (sha256 4 値すべて不変)、`makepkg --verifysource` も 4/4 Passed。setuptool.sh が同一のため `awk '/Unit/,/EOT/'` 抽出入力 (`docker.service`) も従来と同一。PKGBUILD 改変は `pkgver` 1 行のみ (`pkgrel` は 1 のまま)。AUR も 29.10.1-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。`.deps.lock` は変更なし (Arch 公式 `extra` の `docker` は本 issue 処理時点でまだ `1:29.10.0-1`、2026-11-25 に out-of-date flag が立った段階。rootlesskit 3.3.0-1 / slirp4netns 1.3.7-1 / bash 5.3.20-1 も不変)。Closes #842。

- **2026-11-20 / 29.10.0** — approve (Issue #822)。Moby v29.10.0 (release 2026-11-16、by `vvoland`) へ sync (`pkgrel` は 2 → 1 に reset)。annotated tag object `f411721f04b2d1f8e1ab6aff34037acac7189520` の target commit `06db1734267bf7770914b417bff6556d835bf8cc`、tagger `vvoland` (Paweł Gronowski, Docker 社員、GPG 署名 `verified: true`)。release note は **minor release** で、Engine API 1.57 化 / API spec の OpenAPI 3.2 移行 / `insecure-registries` 明示時に loopback を自動 insecure 扱いしない **挙動変更** (⚠️ potential breaking change) / jobs extension (experimental) / embedded containerd の `/etc/docker/containerd.d/` drop-in / libnetwork (nftables backend / overlay / DNS) 修正群 / rootless 修正 2 件 (SELinux 有効時の overlay2 許可、`nosuid`/`nodev` bind mount 下での `docker cp` EPERM 修正) / BuildKit v0.35.0 / Go 1.27.4 / runc v1.5.3 を含むが、本 pkg が ship する 2 script (`dockerd-rootless.sh` / `dockerd-rootless-setuptool.sh`) は raw URL 実測で 29.9.3 と **byte-identical** (sha256 4 値すべて不変)。tag 範囲の差分 (`docker-v29.9.3...docker-v29.10.0`、1084 commits / 300 files) で `contrib/` の変更は `check-config.sh` のみ (本 pkg は未 ship)。PKGBUILD 改変は `pkgver` と `pkgrel` の 2 行のみ。AUR も 29.10.0-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。`.deps.lock` は Arch 公式 `extra` の `docker` が `1:29.10.0-1` に追従したため更新 (rootlesskit 3.3.0-1 / slirp4netns 1.3.7-1 / bash 5.3.20-1 は不変)。Closes #822。

- **2026-11-05 / 29.9.3** — approve (Issue #772)。Moby v29.9.3 (release 2026-11-03、by `vvoland`) へ sync。annotated tag object `320259ace9cacaf345dd04207b28a26554ae95c2` の target commit `70efc0e31da3a115fc87574d1b786fa2b867c346`、tagger `vvoland` (Paweł Gronowski, Docker 社員、GPG 署名 `verified: true`)。release note は security release (CVE-2026-116587 RSA key parse の CPU 消費 / CVE-2026-116591 zstd 展開 panic / CVE-2026-115538 Swarm overlay の暗号化・spoofing 保護 / CVE-2026-115539 SwarmKit manager 証明書 / BuildKit v0.34.2 の GHSA 2 件) + bug fix 3 件だが、いずれも daemon / CLI / BuildKit / SwarmKit 側で、本 pkg が ship する 2 script は 29.9.2 と **byte-identical** (git blob SHA も一致) で sha256 4 値すべて不変。tag 範囲の差分 (`docker-v29.9.2...docker-v29.9.3`、29 commits / 125 files) に `contrib/` の変更なし。PKGBUILD 改変は `pkgver` 1 行のみ。AUR も 29.9.3-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。Arch 公式 `extra` の `docker` は本 issue 処理時点でまだ `1:29.9.2-1` (`extra-testing` に `1:29.9.3-1` が入った段階) のため `.deps.lock` は変更なし (rootlesskit 3.3.0-1 / bash 5.3.20-1 / slirp4netns 1.3.6-1 も不変)。Closes #772。

- **2026-10-22 / 29.9.2** — approve (Issue #745)。Moby v29.9.2 (release 2026-10-20、by `vvoland`) へ sync。annotated tag object `d7dfd6629b22aed77600050ab53695149b8c550d` の target commit `0367cbeb49b9d98cb0b4f51e95f7b54b4803e76c`、tagger `vvoland` (Paweł Gronowski, Docker 社員、GPG 署名 `verified: true`)。release note は Go runtime 1.26.10 更新による CVE-2026-111943 修正 (eMudhra 中間 CA 8 件の失効を証明書チェーン検証で拒否する Go 側の変更) + packaging update のみ。tag 範囲の差分 (`docker-v29.9.1...docker-v29.9.2`、232 commits) は Dockerfile / CI workflow / `.golangci.yml` / `releases/versions.yaml` のみで `contrib/` に変更なし。本 pkg が ship する 2 script は 29.9.1 と **byte-identical** (git blob SHA も一致) で sha256 4 値すべて不変、PKGBUILD 改変は `pkgver` 1 行のみ。AUR も 29.9.2-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。`.deps.lock` を現行 Arch repo へ更新 (docker 1:29.9.1-1 → 1:29.9.2-1、rootlesskit 3.2.0-1 → 3.3.0-1、bash / slirp4netns は不変)。Closes #745。

- **2026-10-15 / 29.9.1** — approve (Issue #718)。Moby v29.9.1 (release 2026-10-14、by `vvoland`) へ sync。annotated tag object `512aee916afea72074ed1c2fe751a3223e494c2b` の target commit `b26fb77d3e7b01f9d21b89539c1ab428f5ca431f`、tagger `vvoland` (Paweł Gronowski, Docker 社員、GPG 署名 valid)。release note は containerd v2.4.2 static binaries 更新による CVE (CVE-2026-95839 / GHSA-8647-2jxj-p2rc / GHSA-hc94-p972-ch66、いずれも containerd 側) + rootless `--domainname` fix + daemon bug fix 群 + RootlessKit v3.3.0 の packaging 更新だが、本 pkg が ship する 2 script は 29.9.0 と **byte-identical** で sha256 4 値すべて不変、PKGBUILD 改変は `pkgver` 1 行のみ (`sh -n` 成功、setuptool.sh の `awk` 抽出入力も同一)。AUR も 29.9.1-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。`.deps.lock` を現行 Arch repo へ refresh (bash 5.3.20-1 / docker 1:29.9.0-1 / rootlesskit 3.2.0-1 / slirp4netns 1.3.6-1)。Closes #718。

- **2026-10-10 / 29.9.0** — approve (Issue #701)。Moby v29.9.0 (release 2026-10-08、by `vvoland`) へ sync。annotated tag object `007c041bb4308ac5fdf552e03f92d0d9ff5dd41b` の target commit `a5b58c94d334674eae3a3207d4f05417643a90d0`、tagger `vvoland` (Paweł Gronowski, Docker 社員、GPG 署名 valid)。release note は Go runtime 更新による CVE 5 件修正 (CVE-2026-97032 / 78659 / 78663 / 78669 / 56857、いずれも daemon 側 HTTP/2・Windows 経路) + BuildKit v0.34.0 / Go 1.26.9 / containerd v2.4.1 / RootlessKit 3.2.0 の packaging 更新だが、本 pkg が ship する 2 script は 29.8.2 と **byte-identical** で sha256 4 値すべて不変、PKGBUILD 改変は `pkgver` 1 行のみ (`sh -n` 成功、setuptool.sh の `awk` 抽出入力も同一)。AUR も 29.9.0-1 で sha256 4 値一致、depends 構成に変更なし (意図的 diff 維持)。Closes #701。

- **2026-10-02 / 29.8.2** — approve (Issue #678)。Moby v29.8.2 (release 2026-09-30、by `vvoland`) へ sync。annotated tag object `a8816ccac86853a18a9903da4dfaaebf01cb1774` の target commit `8af9fe3a36bab3e039862a2ab1cef1880c9b4d03`、tagger `vvoland` (Paweł Gronowski, Docker 社員)。release は security release (Docker Engine / BuildKit の CVE 複数) だが、本 pkg が ship する 2 script は 29.8.1 と **byte-identical** で sha256 4 値すべて不変、PKGBUILD 改変は `pkgver` 1 行のみ。setuptool.sh の `awk` 抽出入力も同一のため docker.service 出力に変化なし。

- **2026-09-21 / 29.8.1** — approve (Issue #639)。Moby v29.8.1
  (release 2026-09-15、by `vvoland`) へ sync。annotated tag object
  `b2d20c90a74af78b3f0f967db92292e4a603c03d` の target commit
  `464cd50c3d9e92877d56940ea160de6fca7bea23`、tagger `vvoland`
  (Paweł Gronowski, Docker 社員)。release note は bug fix (containerd
  image store / CRI pod-sandbox 警告 / OpenVZ user-namespace 検出 /
  Windows hard link) + containerd static binaries を v2.3.5 に更新の
  packaging update のみ、CVE 番号なし。`dockerd-rootless.sh` /
  `dockerd-rootless-setuptool.sh` は v29.8.0 と **完全同一** (= sha256
  4 値すべて不変、byte 単位で一致確認)。PKGBUILD 改変は `pkgver=29.8.1`
  の 1 行のみ。Arch 公式 `docker` package は本 issue 処理時点でまだ
  `1:29.8.0-1` (= `depends.docker` の追従は `dep-version-pr.yml` 側で
  別途検知される想定、CLAUDE.md「PR review 時の個別事情」表参照)。
  Closes #639。

- **2026-09-06 / 29.8.0** — approve (Issue #596)。Moby v29.8.0
  (release 2026-09-03、by `vvoland`) へsync。annotated tag object
  `dc4db3d292c317ca216dae9301dffc935c8d7680` と target commit
  `3ce5872b7950c63ba2ffbc5123101019ff3e6682` はともに署名検証済み。
  `dockerd-rootless.sh` の実効差分は pesto port driver のコメント2行のみ
  (+2/-1、実行行変更なし)、setuptoolはbyte-identical。raw GitHubとContents
  API blobを照合し sha256 を独立実測: rootless script `2002036338…`、
  setuptool `1c9f0dc9…` (不変)。両scriptは `sh -n` 成功、systemd unit抽出結果も
  従来と同一。dependsの必須変更なし。AUR現行との依存方針差を再確認し、
  `lxc` / `passt` をoptdependsへ追加、slirp4netns hard dependencyは維持。
  `.deps.lock` を現行Arch repoへrefresh。Closes #596。

- **2026-08-09 / 29.7.2** — approve (Issue #502)。Docker Engine v29.7.2
  (release 2026-08-06、by `vvoland`) への sync。upstream tag commit:
  `6a43e3d5afddf4111da0f864bbc7cae5d7e95001`。bug fix のみのリリース (CVE 番号
  記載無し、docker service create/update の環境変数重複指定時 panic 修正、
  v29.7.0 で混入した image pull リグレッション 2 件の修正)。rootless スクリプト群
  (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`) の中身は v29.7.1 と
  **完全同一** (= sha256 4 値すべて不変、byte 単位で一致確認)。PKGBUILD 改変は
  `pkgver=29.7.2` の 1 行のみ。tag は annotated + GPG 署名済み (`verified: true`)、
  tagger は既知の Docker Inc. maintainer `vvoland` (Paweł Gronowski)。sha256
  独立再検証済み (moby/moby docker-v29.7.2 の raw fetch + 実測)。Closes #502。

- **2026-08-01 / 29.7.1** — approve (Issue #470)。Docker Engine v29.7.1 (release
  2026-07-31、by `vvoland`) への sync。upstream tag commit:
  `c5b8ce9274b5c00cb1f8287c8e258edc1f01176d`。regression fix 2 件
  (image pull で親ディレクトリ entry が無い layer が pull できない不具合、
  `CopyToContainer` が `/var/run` -> `/run` のような絶対 symlink 経路を誤って
  拒否する不具合) のみで CVE 番号の記載無し。rootless スクリプト群
  (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`) の中身は v29.7.0 と
  **完全同一** (= sha256 4 値すべて不変、byte 単位で一致確認)。PKGBUILD 改変は
  `pkgver=29.7.1` の 1 行のみ。tag は annotated + GPG 署名済み (`verified: true`)、
  tagger は既知の Docker Inc. maintainer `vvoland` (Paweł Gronowski)。sha256
  独立再検証済み (moby/moby docker-v29.7.1 の raw fetch + 実測)。Closes #470。

- **2026-07-31 / 29.7.0** — approve (Issue #463)。Docker Engine v29.7.0 (release 2026-07-30、by `vvoland`)
  への sync。upstream tag commit: `4b5cb715735a1b7000093bb2ab54295dfedfebfe`。
  **security release**: `github.com/moby/go-archive` を v0.3.0 に更新し CVE-2026-17106
  (GHSA-hfg8-hc9c-6c3h、細工された tar archive 内の link を使った path traversal) を修正。
  fix commit は go-archive maintainer `thaJeztah` (Sebastiaan van Stijn) による正当な
  upstream security fix。ただし CVE 群は go-archive (tar 展開) 側の fix であり、本 pkg が
  取得する rootless スクリプト群 (`dockerd-rootless.sh` + `dockerd-rootless-setuptool.sh`)
  の中身は v29.6.2 と **完全同一** (= sha256 4 値すべて不変、byte 単位で一致確認)。
  PKGBUILD 改変は `pkgver=29.7.0` の 1 行のみ。sha256 独立再検証済み
  (moby/moby docker-v29.7.0 の raw fetch + 実測)。Closes #463。

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
- 29.9.3-2 (2026-11-12, pkgrel bump のみ): `pkgrel` +1 (deps changed): docker 1:29.9.2-1 → 1:29.9.3-1, slirp4netns 1.3.6-1 → 1.3.7-1
