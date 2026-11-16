# nekonolock review

## 状態

**review 済み、 approve** (2026-11-16、 v0.3.3 patch)

自家 source PKGBUILD (= AUR fork ではない、 nananek 自身の upstream を
GitHub release tag archive から直接 vendor)。 own-package フローは
nekono-btkeycast / nekono-pipewire-mcp と同一。

## 用途

Sway (wlroots) 用のスクリーンロッカー。 `ext-session-lock-v1` で画面を
ロックし、 YubiKey の OpenPGP applet によるチャレンジ署名 (`nekonolock-auth`
経由) またはパスワード (PAM) で解除する。 UI (`nekonolock`) と認証
(`nekonolock-auth`) を別プロセスに分け、 `SOCK_SEQPACKET` の Unix socket で
接続する。 scdaemon と pcscd のカード取り合いを構造的に避けるため、
ロック画面と sudo は OpenPGP applet、SDDM / login / TTY は PIV と使い分ける
(upstream `docs/decisions.md` ADR-0003 / ADR-0004)。

## Source

- Upstream: https://github.com/nananek/nekonolock
  - 主開発者 nananek (= 自家 upstream)
  - tag `v0.3.3` (= prefix "v")、 commit `5c99476f281f07830f68764725477d4be7bb18fe`
  - tag GPG 署名: **verified** (GitHub API `verification.verified=true`、
    signer `483DC691DF9F29327EA106BD030130E2F156CD74`)
  - license: MIT
- distfile: GitHub Release archive (= `archive/refs/tags/v0.3.3.tar.gz`)

## 検証結果

- [x] `source` URL = `github.com/nananek/nekonolock/archive/refs/tags/v0.3.3.tar.gz`
  - 自家 upstream、 typosquat 検討対象外
- [x] `sha256sums` 独立検証
  - 実測 (review 時に curl 直取得 + `sha256sum`):
    `f15510f901ddec9985e4c9d85e2684e3091738f703c3b1cc6de2ba8d46af4210`
  - 同梱 file 5 点も `packaging/` から再計算し、 PKGBUILD の値と一致
    - `nekonolock.pam` `f4ca0637...` (password-only chain)
    - `nekonolock-auth.socket` `79d6f29f...`
    - `nekonolock-auth.service` `3576102a...`
    - `nekonolock-sudo.toml` `ef8eed8f...`
    - `LICENSE` `e8a8dae9...`
- [x] tag `v0.3.3` GPG 署名状態: verified (上記 Source 参照)
- [x] `prepare()`: `cargo fetch --locked --target <host>` のみ。
  `Cargo.lock` を commit 済みで、 build 時の動的取得は無い
- [x] `build()`: `cargo build --release --frozen -p nekonolock -p nekonolock-auth`
  - v0.3.0 archive は `crates/pam` を含むため `-p pam-nekonolock` も追加される
  - `--frozen` で lockfile / registry の変更を拒否する
  - shell injection / eval / curl / wget / pipe-to-shell なし
- [x] `check()`: `cargo test --frozen --workspace` — upstream CI と同じ
  (`.github/workflows/ci.yml`)。 追加のテスト依存取得なし
- [x] `package()`: `install` のみ。 install script / post-install hook なし
  - `pam_nekonolock.so` を `/usr/lib/security/` に、 `sudo.toml` の例を
    `/usr/share/doc/nekonolock/` に置く
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
  - user unit は sandbox を入れず `UMask=0077` だけ。`CapabilityBoundingSet=` は
    `status=218/CAPABILITIES`、`RestrictAddressFamilies` などは PAM の setuid
    経路を壊すため入れない
- [x] PAM: `/etc/pam.d/nekonolock` の `auth` は `pam_faillock` + `pam_unix`
  の password-only chain。 `account` は `login` include で `pam_acct_mgmt` を
  通す。 `system-auth` の `pam_pkcs11` を先に通すと、 ロック画面の password が
  PIV PIN としてカードに送られるため。 `backup=` で pacman の管理対象にする。
  sudo 用は `/usr/lib/security/pam_nekonolock.so` を同梱し、 既定では PAM 設定を
  変えない (ansible の `nekonolock_sudo_auth` で有効化する)
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
| **sudo 用 PAM モジュールを同梱（既定無効）** | root への認証はユーザーが差し替えられるヘルパの応答を信頼しない。 root 所有の設定と公開鍵で検証し、 有効化は ansible で明示する |

## 結論

**approve** — build host で `bin/build-all nekonolock` で build + sign + repo db
追加可。v0.3.0 から `pam_nekonolock.so` も同梱される。 PAM file はパッケージ同梱の既定を使い、 ホスト固有の変更は
ansible-nekonodesk の `nekonolock` role で行う。

publish 後の配備:
1. `pacman -Sy nekonolock` で install
2. `systemctl --user enable --now nekonolock-auth.socket`
3. swayidle (またはロック用の呼び出し元) を `nekonolock` に差し替える

## 更新方針

upstream で新 release (= v0.4.0 等) が出たら:
1. user が `git tag -s v0.4.0 -m "Release 0.4.0"` + `git push origin v0.4.0` +
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
| 2026-11-15 | 0.2.1-1 | (this PR) | `b84c76f31eb0d0ae7dd4eeb9b93f359330963a79` | 0.2.1 bump (共通ベースライン / 全大文字コピー)。 depends / systemd unit / PAM は不変。 tag GPG 署名 verified、 tag archive sha256 独立再計算 |
| 2026-11-15 | 0.3.0-1 | (this PR) | `e77e7bf38df31bfce1db63f72156b65e0d3e839d` | 0.3.0 bump (sudo 用 PAM モジュール / PIN キャッシュ消去 / カード抜去時の password 切替)。 systemd unit を安全な sandbox に更新。 tag GPG 署名 verified、 tag archive sha256 独立再計算 |
| 2026-11-15 | 0.3.1-1 | (this PR) | `89e7285ed893ca562ffd182bc41d2c15f9591315` | 0.3.1 patch (ロック画面の PAM auth を password-only chain / カード抜去検知 1 秒)。 PAM file の hash を更新。 tag GPG 署名 verified、 tag archive sha256 独立再計算 |
| 2026-11-15 | 0.3.2-1 | (this PR) | `af5b78cd0d2602f3ae069a268356096b6043c83c` | 0.3.2 patch (認証ヘルパの systemd sandbox を撤回し UMask のみに戻す)。 unit の hash を更新。 tag GPG 署名 verified、 tag archive sha256 独立再計算 |
| 2026-11-16 | 0.3.3-1 | (this PR) | `5c99476f281f07830f68764725477d4be7bb18fe` | 0.3.3 patch (左上ラベルを既定でマシン名 / CLEAR_PASSPHRASE の ERR 判定 / PAM 署名の一時ファイルを fd 基準に)。 depends / systemd unit / PAM は不変。 tag GPG 署名 verified、 tag archive sha256 独立再計算 |
