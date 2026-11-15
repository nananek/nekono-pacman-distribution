# nekonolock review

## 状態

**review 済み、 approve** (2026-11-15、 v0.2.0 UI update)

自家 source PKGBUILD (= AUR fork ではない、 nananek 自身の upstream を
GitHub release tag archive から直接 vendor)。 own-package フローは
nekono-btkeycast / nekono-pipewire-mcp と同一。

## 用途

Sway (wlroots) 用のスクリーンロッカー。 `ext-session-lock-v1` で画面を
ロックし、 YubiKey の OpenPGP applet によるチャレンジ署名 (`nekonolock-auth`
経由) またはパスワード (PAM) で解除する。 UI (`nekonolock`) と認証
(`nekonolock-auth`) を別プロセスに分け、 `SOCK_SEQPACKET` の Unix socket で
接続する。 scdaemon と pcscd のカード取り合いを構造的に避けるため、
PIV / pam_pkcs11 ではなく OpenPGP applet 一本に寄せている
(upstream `docs/decisions.md` ADR-0003 / ADR-0004)。

## Source

- Upstream: https://github.com/nananek/nekonolock
  - 主開発者 nananek (= 自家 upstream)
  - tag `v0.2.0` (= prefix "v")、 commit `4cd874d301fbc5d043d30cf6a349340cf4b93149`
  - tag GPG 署名: **verified** (GitHub API `verification.verified=true`、
    signer `483DC691DF9F29327EA106BD030130E2F156CD74`)
  - license: MIT
- distfile: GitHub Release archive (= `archive/refs/tags/v0.1.0.tar.gz`)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekonolock/archive/refs/tags/v0.2.0.tar.gz`
  - 自家 upstream、 typosquat 検討対象外
- [x] `sha256sums` 独立検証
  - 実測 (review 時に curl 直取得 + `sha256sum`):
    `2126593d4b115b588b69074d9362b6b67d403a741e6dba56dc05256abb50ae0e`
  - 同梱 file 4 点も `packaging/` から再計算し、 PKGBUILD の値と一致
    - `nekonolock.pam` `cf1dfa38...`
    - `nekonolock-auth.socket` `79d6f29f...`
    - `nekonolock-auth.service` `510be1b0...`
    - `LICENSE` `e8a8dae9...`
- [x] tag `v0.2.0` GPG 署名状態: verified (上記 Source 参照)
- [x] `prepare()`: `cargo fetch --locked --target <host>` のみ。
  `Cargo.lock` を commit 済みで、 build 時の動的取得は無い
- [x] `build()`: `cargo build --release --frozen -p nekonolock -p nekonolock-auth`
  - `--frozen` で lockfile / registry の変更を拒否する
  - shell injection / eval / curl / wget / pipe-to-shell なし
- [x] `check()`: `cargo test --frozen --workspace` — upstream CI と同じ
  (`.github/workflows/ci.yml`)。 追加のテスト依存取得なし
- [x] `package()`: `install` のみ。 install script / post-install hook なし
- [x] `depends`: upstream の実装と照合
  - `gnupg` — OpenPGP applet へのチャレンジ署名 (`gpg --detach-sign` /
    `--verify`) を subprocess で呼ぶ
  - `libxkbcommon` — `wl_keyboard` のキーマップ変換。 upstream README が
    build 依存として明記
  - `pam` — `libpam.so.0` を `libloading` で dlopen する
  - `ttf-dejavu` — `[ui] font` の既定フォント。 fontconfig 経由で解決する
- [x] `makedepends`: `rust` (`cargo` を含む)。 upstream の `rust-version =
  1.99` を満たす。 build host の rust 1.100.0-1 で確認
- [x] `optdepends`: `pcsclite` / `ccid` — pcscd 経由で YubiKey を使うホスト
  向け。 既定の scdaemon 直接経路では不要
- [x] `systemd` unit:
  - `nekonolock-auth.socket` は `ListenSequentialPacket=` (= `SOCK_SEQPACKET`)。
    `ListenStream=` にすると `SOCK_STREAM` になり、 ヘルパが起動時に種別を
    検証して拒否する (upstream 設計 §6)
  - `nekonolock-auth.service` は socket activation 専用で `[Install]` を
    持たない。 enable するのは `.socket` だけ
  - `SocketMode=0600` / `DirectoryMode=0700` / `UMask=0077`
- [x] PAM: `/etc/pam.d/nekonolock` は `auth` と `account` の両方で `login` を
  include する。 認証ヘルパが `pam_authenticate` の後に `pam_acct_mgmt` を
  呼ぶため `account` が要る。 `backup=` で pacman の管理対象にする
- [x] `secrets` 混入なし
- [x] `arch=('x86_64')` — `libxkbcommon` / `libpam` に link する native binary。
  upstream CI も x86_64 の Ubuntu で検証している

## 設計判断 (= AUR fork でないので「AUR との意図的差分」 ではなく §設計判断)

| 判断 | 理由 |
|---|---|
| **AUR ではなく自家 upstream を直接 vendor** | upstream が nananek/nekonolock 自身。 AUR に同名 package は無い |
| **`nekonolock` と `nekonolock-auth` を 1 package に同梱** | 2 プロセスは同じ protocol crate を共有し、 バージョンがずれると socket の互換性が壊れる。 同一 package で同時に更新する |
| **`.service` に `[Install]` を付けない** | socket activation 専用。 `.service` を enable すると socket を経由せず起動できてしまい、 `SOCK_SEQPACKET` 前提が崩れる |
| **PAM file を package に同梱し、 ansible 側を正にする** | パスワード認証の include 先はホストの PAM 方針 (pam_faillock 等) に依存する。 package は動く既定を置き、 ansible role が上書きする |
| **`ttf-dejavu` を depends に置く** | `[ui] font` の既定値が DejaVu。 フォントが無いとロック画面が描画できない |
| **`pcsclite` / `ccid` は optdepends** | scdaemon の `disable-ccid` 運用が既定。 pcscd 経由はホスト設定で選ぶ |

## 結論

**approve** — build host で `bin/build-all nekonolock` で build + sign + repo db
追加可。 PAM file はパッケージ同梱の既定を使い、 ホスト固有の変更は
ansible-nekonodesk の `nekonolock` role で行う。

publish 後の配備:
1. `pacman -Sy nekonolock` で install
2. `systemctl --user enable --now nekonolock-auth.socket`
3. swayidle (またはロック用の呼び出し元) を `nekonolock` に差し替える

## 更新方針

upstream で新 release (= v0.3.0 等) が出たら:
1. user が `git tag -s v0.3.0 -m "Release 0.3.0"` + `git push origin v0.3.0` +
   GitHub Release publish
2. nvchecker (GitHub Releases、 `nananek/nekonolock`、 prefix "v") が新 release
   を検知 → Issue が立つ
3. PKGBUILD の pkgver を更新、 tag archive の sha256 を独立再計算
4. depends / systemd unit / PAM file の変化を source diff で確認
5. `.SRCINFO` + `.deps.lock` を再生成
6. REVIEW.md「更新履歴」 に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-11-15 | 0.1.0-1 | (this PR) | `10ee9c85c92d6769b45bc3c2b20ffc89cde27c80` | 初回 add、自家 source v0.1.0。 tag GPG 署名 verified、 tag archive sha256 独立再計算、 依存と systemd unit を source と照合 |
| 2026-11-15 | 0.2.0-1 | (this PR) | `4cd874d301fbc5d043d30cf6a349340cf4b93149` | 0.2.0 bump (Ink UI / HiDPI / 無操作時の純黒)。 depends / systemd unit / PAM は不変。 tag GPG 署名 verified、 tag archive sha256 独立再計算、 `.deps.lock` の rust を 1.100.0-1 に更新 |
