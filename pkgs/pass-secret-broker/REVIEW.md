# pass-secret-broker review

## 状態

**review 済み、approve** (2026-11-28)

AUR には存在しない自家 package。upstream は `nananek/pass-secret-broker`。
`pass-secret-service` を置き換え、sandbox 内アプリごとに Secret Service の
store を分離する broker と、その systemd user unit を配布する。

## Source

- Upstream: https://github.com/nananek/pass-secret-broker
  - license: MIT
  - release: `v0.1.0` (signed annotated tag)
  - tag commit: `daff4e71e1a4d434afbdc7de900b54392afe118a`
- source URL: `https://github.com/nananek/pass-secret-broker/archive/refs/tags/v0.1.0.tar.gz`
- b2sum: `ed6d08f6c5944fa07a274d396b905fb04ac8b2651e58879e843f3c48bf27d0ea6924f79e42380d1e279093cbfd9bb1252e713688062d5323591975bc59a8692d`
- sha256 (参考): `d567cd8c81c77a6370a0f72eb947ba942dc6ea07c47a7db9c4addd5ead92df43`

## 検証結果

- [x] source URL は upstream の tag archive。tag は Nekono GPG で署名済み
- [x] b2sums がダウンロードした tarball と一致
- [x] `prepare()` は `cargo fetch --locked` のみ。build は `cargo build --frozen`
- [x] `build()` / `package()` に curl / wget / eval / pipe-to-shell 無し
- [x] `check()` で `cargo test --frozen --features test-support` を実行
  - 24 integration tests + 43 unit tests が成功
  - 監査ログに secret / label / attribute の値が出ないこともテスト済み
- [x] depends: `gcc-libs` / `glibc` / `dbus` / `gnupg`
  - `dbus`: provider の private bus を起動する `dbus-daemon`
  - `gnupg`: 内蔵 provider が secret を暗号化・復号する `gpg`
- [x] makedepends: `rust` (`cargo` を含む)
- [x] package 内容: binary / systemd user unit 2 本 / license / docs
- [x] `pass-secret-service` を `provides` / `replaces` / `conflicts` で置換え
- [x] デスクトップ通知は実装しない。監査は broker の journal ログで行う

## check() の TMPDIR

provider-run の private bus socket は D-Bus の path 長制限 (100 bytes 未満)
を受ける。`$srcdir` 配下を TMPDIR にすると path が長すぎてテストが落ちるため、
`check()` では `mktemp -d /var/tmp/pass-secret-broker-check.XXXXXX` を使う。
`/var/tmp` は exec 可能で、テストが一時 dir に置く fake `gpg` も実行できる。

## 意図的な差分 (AUR との比較ではなく設計判断)

| 項目 | 内容 | 理由 |
|---|---|---|
| package 名 | `pass-secret-broker` | upstream の repo 名に合わせる |
| `provides` / `replaces` | `pass-secret-service` を置換え | 同じ `org.freedesktop.secrets` を所有するため共存できない |
| unit の配置 | `/usr/lib/systemd/user/` | package 提供 unit。ansible は enable のみ |
| 通知 | 実装しない | 読み出しごとの通知は日常操作で絶えず、監査の役に立たない |
| 監査 | broker の journal ログ | app ID / 呼び出し元 PID / sender / method / object path を記録 |

## 結論

**approve** — build host で `makepkg -s --sign --key 483D...` 可。
`pass-secret-service` からの移行は ansible の `pass_secret_broker` role と
dotfiles の wrapper 登録で行う。旧 `Default/` store は自動移行しない。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-11-28 | 0.1.0-1 | (this PR) | `daff4e7` (v0.1.0) | 初回 add。自家 source、`pass-secret-service` 置換え、通知なし・監査ログあり |
