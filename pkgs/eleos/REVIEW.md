# eleos review

## 状態

**review 済み、approve** (2026-11-14、0.8.7 へ bump。初回 review は 2026-11-11 の
自家 PKGBUILD 新規作成 + Arch VM での実機検証)

AUR に `eleos` は存在しない (`aur.archlinux.org/rpc/v5/info?arg[]=eleos` が 2026-11-11 時点で
`resultcount: 0`。2026-11-13 の bump 時は AUR が TLS で到達できず再確認できなかったが、
release notes にも AUR 追加の記載は無い)。upstream の GitHub Releases / COPR / APT / snap の
いずれにも Arch 向け package は無かったため、本 repo で PKGBUILD を新規作成した (AUR fork ではない)。

> 2026-11-12 の v0.8.6 で upstream 自身が `packaging/arch/` に PKGBUILD を追加し、
> GitHub Pages の pacman repository (`[eleos]`) で `eleos` を配布し始めた (経緯は下記
> 「upstream の Arch packaging との関係」)。本 repo の PKGBUILD は従来どおり自家管理で、
> upstream の release asset は使わない。

## 用途

nftables を **netlink 経由で直接** 操作する Go 製ファイアウォール。`nft(8)` は
使わず、外部 API も叩かない (= 完全オフラインで rule を生成・適用できる)。
設定は YAML (`/etc/eleos.yml`) 1 枚で、`elctl enable|apply|disable` で適用する。

- `elctl enable` — 現在の設定を適用 (systemd unit の `ExecStart`)
- `elctl apply` — 設定変更の再適用 (`ExecReload`)
- `elctl disable` — `table inet eleos` を削除 (`ExecStop`)

既存 firewall (ufw / firewalld) とは併用しない前提。`default.allowAllOut: true` なら
outbound は素通し、inbound は `ports:` に列挙したものだけ許可、`allowPing` で
ICMP echo をレートリミット付き許可、port scan 対策 (XMAS / NULL / 小さすぎる MSS
/ IP fragment の drop) と SYN-flood のレートリミットが既定で入る。

## Source

- AUR: **なし** (2026-11-11 時点で RPC `info` が 0 件)
- Upstream: https://github.com/nexryai/eleos
  - owner: nexryai (作成 2023-06-17、stars 13、forks 0、MIT)
  - `url` / README / `eleos.spec` の URL と一致 (typosquat / なりすまし無し)
- Tag: `v0.8.7` (annotated tag `586b2301ff479d4cd9594d933a78feb0b485b11b`、
  SSH 署名 verified) → commit `276ce1ff0bcc90f81d26c4e6ceec2afc4c0d5584`
  (2026-11-14T07:59:44Z "Deploy documentation alongside signed package repositories
  on release"、tag 自体は署名付き)
- 取得元: `https://github.com/nexryai/eleos/archive/refs/tags/v0.8.7.tar.gz`
  (GitHub の release tag tarball。upstream の公式 release 手順と同じ)
- release: 2026-11-14T08:03:54Z 公開 (draft / prerelease ではない)、author は
  従来と同じ `github-actions[bot]` (release workflow による自動公開)

## 検証結果

- [x] `source` URL = upstream の公式 release tag path。domain / owner は upstream と一致
      (v0.8.6 から変化なし)
- [x] `sha256sums` を独立実測: `81561d10b71dc0acdf3a5185cfcea425d08bbc314596475d867d8acae973ed37`
      (= `fetch` した raw bytes の sha256。`makepkg --verifysource` も同じ値で Passed)。
      upstream が `main` の `packaging/arch/{PKGBUILD,.SRCINFO}` に公開している
      `sha256sums` とも一致 (2026-11-14 取得)。tag `v0.8.7` は annotated tag で
      SSH 署名 verified
- [x] 旧新 tarball (`v0.8.6` → `v0.8.7`) の diff: 変更 43 file の内訳は
      `.github/workflows/{release,snap}.yml` / `AGENTS.md` / `README.md` /
      `eleos.spec` / `packaging/arch/{PKGBUILD,.SRCINFO}` と、**新規 `docs/` 一式**
      (Astro + pnpm による多言語ドキュメント site。`pnpm-lock.yaml` を含む)。
      **Go source (`main.go` / `internal/` / `integration/` / `go.mod` / `go.sum` /
      `config.default.yml` / `systemd/eleos.service`) は byte 一致**で、build 対象と
      依存に変化なし (`go.sum` は両 tag で sha256 が同一であることを確認)
- [x] `go.mod`: `go 1.26.10` で 0.8.6 と同一。Arch の `go` は `.deps.lock` の
      `2:1.27.4-1` でより新しいため `GOTOOLCHAIN=auto` の toolchain download は走らない
- [x] `build()` / `check()` / `package()` に curl / wget / exec / eval / pipe-to-shell /
      動的 fetch / `pip install` は無い。`go build` + `go test ./...` + `install` のみ。
      diff が無いため今回は build script を変更していない
- [x] Go の依存は `go.sum` で checksum pin 済みで、0.8.6 と完全一致
      (`github.com/google/nftables` v0.3.0, `mdlayher/netlink`, `lorenzosaino/go-sysctl`,
      `golang.org/x/sys` v0.45.0, `gopkg.in/yaml.v2` v2.4.0)。新規依存なし
- [x] `makedepends=('go')` のみ、`depends=()` は不変。`CGO_ENABLED=0` の完全 static
      binary で、実行時に必要なのは kernel の nf_tables だけ
- [x] `arch=('x86_64')` に限定 ([nekono] は `repo/x86_64/` のみ配信)。upstream の CI は
      arm / riscv64 / s390x 等も build し、pacman repo では aarch64 も配布するが、
      本 repo では x86_64 のみ
- [x] `pkgver` は literal (`0.8.7`)。`$` 展開を含まないので upstream 検知と比較可能
- [x] `prepare()` は無し (patch 無し)
- [x] `*.install` (post_install 等) は追加していない。`packaging/deb/build-deb.sh` や
      `snap/` hook は本 PKGBUILD では使わない
- [x] `systemd/eleos.service` は 0.8.6 と byte 一致 (hardening 設定・`ExecStart` /
      `ExecReload` / `ExecStop`・`/etc/eleos.yml` 参照先すべて不変)
- [x] `/etc/eleos.yml` は upstream `config.default.yml` を byte 一致で install し、
      `backup=('etc/eleos.yml')` で upgrade 時の設定保持を保証。`config.default.yml` も
      0.8.6 と byte 一致
- [x] 設定 schema (`default:` / `ports:` / `router:` / `nat:`) の破壊的変更なし
      (Go source と default config が無変更のため)
- [x] secret 混入なし。`.git` / download cache を package に含めない
- [x] 落とし穴 #1–7 非該当 (prebuilt binary 無し / Python 無し / 7z 無し /
      環境依存 source 分岐無し / prepare 専用 tool 無し / 自前 tarball 無し)
- [ ] **実機検証は今回未実施** (0.8.6 で行った Arch VM での 50 check は、Go source・
      systemd unit・default config が byte 一致のため再実行していない。binary の中身は
      同一の source から生成されるが、`bin/build-all` と client 検証は人間の実行に委ねる)

### 実機検証 (Arch VM / KVM)

`bin/build-all` 前の段階で、makepkg が生成した `eleos-0.8.5-1-x86_64.pkg.tar.zst` を
**素の Arch Linux (kernel 7.2.15-arch1-1, systemd, nftables) の VM** に
`pacman -U` で install し、以下を確認した (計 50 check、全 PASS)。

- package 内容: `/usr/bin/elctl` (static ELF 3.0MB), `usr/lib/systemd/system/eleos.service`,
  `/etc/eleos.yml` (upstream `config.default.yml` と byte 一致), LICENSE, README
- 同梱 default config で `elctl enable` → `table inet eleos` + `udp dport 5353 accept`
  (mDNS) が生成され、`elctl disable` で table が消える
- `systemctl start/reload/stop eleos` (= `elctl enable/apply/disable`) が期待通り動作。
  `systemctl enable` で `multi-user.target.wants` に symlink が張られる
- 生成 ruleset: `input` = `policy drop` + `ct established,related accept` +
  `iifname "lo" accept` + ICMP echo rate limit + IPv6 ND accept + `syn-flood` jump +
  `tcp dport <allowed> accept`、`output` = `policy accept` (allowAllOut 時)
- veth + netns で実通信: 許可 port (2222/tcp) は接続成功、未列挙 port (8081/tcp) は
  **RST ではなく無音 DROP** (timeout)、ICMP echo は許可
- 設定に port を追記 → `elctl apply` (systemd reload) で新 port が通る。
  `apply` の二重実行で table が重複しない (冪等)
- 壊れた YAML は non-zero exit で拒否
- **2 回目の起動** で `eleos.service` が enabled のまま自動適用され、
  login 前の時点で `table inet eleos` と `policy drop` が存在し、未列挙 port が DROP される

検証は使い捨ての Arch VM (KVM) 上で行った。ログは repo には含めない
(再現する場合は同じ条件で Arch を起動し、`pacman -U` の後に `elctl enable` と
veth + netns での接続確認を行えばよい)。

## 依存方針 (upstream の build 手順との diff)

upstream README の "Install from source" は
`CGO_ENABLED=0 go build -o elctl main.go` だが、PKGBUILD では次の 3 点だけ変えた。
いずれも **build 対象と binary の性質を変えない** 範囲の変更:

1. `main.go` 単体指定 → `go build .` (module 全体。`internal/` を確実に取り込む)
2. `-trimpath` を追加 (build 環境の path を binary に残さない)
3. `-ldflags '-s -w'` を追加 (size 削減。upstream の `eleos.spec` と同じ flag)

`depends=()` の根拠: `file` で `statically linked` を確認済み。nftables は kernel 側の
機能で、`nft(8)` にも依存しない。`optdepends` も付けていない (`nft list ruleset` で
中身を見たい場合だけ `nftables` を入れればよい)。

`.deps.lock` は `makedepends=go` の 1 行のみ。`go` は Arch 公式 `extra`
(`2:1.27.4-1` 時点) なので監視対象。`depends=()` / `makedepends=('go')` は 0.8.7 でも不変。

### upstream の Arch packaging との関係

v0.8.6 から upstream は `packaging/arch/{PKGBUILD,.SRCINFO,build.sh}` を持ち、
release workflow が x86_64 / aarch64 の `.pkg.tar.zst` を release asset に上げ、
GitHub Pages の pacman repository でも配布する。upstream の PKGBUILD は本 repo の
PKGBUILD を元にしている (同じ `install` 行・同じ build flag) が、次の点で**意図的に違う**:

- `arch=('x86_64')` のみ ([nekono] は `repo/x86_64/` のみ配信。upstream は aarch64 も配信)
- cross-compile 用の `GOARCH` / `GOOS` / `GOTOOLCHAIN=auto` export と
  `check()` の cross 判定を持たない (native build のみなので不要)
- `options=('!strip' '!debug')` を付けない (cross build しないため strip で問題ない)

v0.8.7 で upstream は release workflow に **tag から `packaging/arch/PKGBUILD` /
`.SRCINFO` / `eleos.spec` を自動更新して main に signed commit で戻す** job を追加した。
そのため tag 時点の `packaging/arch/PKGBUILD` は `pkgver=0.8.6` のまま (main 側が 0.8.7) で、
これは正常。`sha256sums` は workflow が tag tarball を実測して埋めるので、本 repo の
実測値と一致することを独立確認に使える (今回一致)。

release notes / README は upstream の pacman repo を案内するが、本 repo は
release asset を使わず tag tarball から自前 build する。同じ pkgname `eleos` が
upstream repo と [nekono] の両方に存在し得るため、client 側で両 repo を有効にすると
同一 version では repo 優先順で選ばれる (version が上がれば pacman は新しい方を取る)。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-11-11 | 0.8.5-1 | (this PR) | `50ffcf8b` (tag `v0.8.5`) | 受入 (自家 PKGBUILD 新規作成) |
| 2026-11-13 | 0.8.6-1 | (this PR) | `3b227778` (tag `v0.8.6`) | 受入。Go source 変更なし、差分は systemd unit の hardening と upstream の Arch packaging (`packaging/arch/`, release workflow) のみ。`depends` / `makedepends` 不変 |
| 2026-11-14 | 0.8.7-1 | (this PR) | `276ce1ff` (tag `v0.8.7`) | 受入。Go source / `go.mod` / `go.sum` / systemd unit / `config.default.yml` すべて byte 一致。差分は新規 `docs/` (Astro site) と release / snap workflow、`eleos.spec`、upstream `packaging/arch/` のみ。`depends` / `makedepends` 不変。実機検証は未実施 (source 同一のため) |

### 次回 bump 時の確認点

- `source` の tag と sha256 を再実測 (upstream は tag を打ち直す運用はしていないが、
  GitHub tarball は tag が動くと変わるので必ず実測する)。upstream の
  `packaging/arch/.SRCINFO` (main 側) の `sha256sums` と突き合わせると独立確認になる
- `go.mod` の `go` directive が Arch の `go` より新しくないか
  (`go 1.26.10` のように patch まで要求するので、Arch 側が古いと
  `GOTOOLCHAIN=auto` による toolchain download が走る。build 失敗時はここを疑う)
- systemd unit / `config.default.yml` の変更有無 (`diff` で確認し、変わっていれば
  `backup` 対象や rule の意味が変わっていないか review する)
- 0.8.6 で unit に `SystemCallFilter=@system-service` / `RestrictAddressFamilies=AF_NETLINK AF_UNIX`
  / `IPAddressDeny=any` が入った。unit がさらに厳しくなった場合、`elctl enable|apply|disable`
  が netlink 以外 (file write / DNS / exec 等) を必要としていないか実機で確認する
- upstream の `packaging/arch/PKGBUILD` が変わったら、(a) 本 repo の build 手順に取り込むべき
  変更か (b) 本 repo 固有の意図的 diff かの切り分けを更新する。
  なお tag 時点の PKGBUILD は release workflow が main に遅れて更新するため、
  tag 内の `pkgver` が 1 つ古いことがある (異常ではない)
- `docs/` のような build に不要な tree が tarball に増えていないか
  (0.8.7 で +90KB 程度。package には含めないので実害は無い)
- `ports:` / `router:` / `nat:` など設定 schema の破壊的変更が無いか
