# ff2mpv-rust review

## 状態

**review 済み、approve** (2026-05-25、純 fork、改変なし)

AUR の `ff2mpv-rust` PKGBUILD (pkgver=1.1.7, pkgrel=1) を **純 fork**。
唯一の diff は `# Maintainer:` → `# Contributor:` の置換 + fork 説明
コメント追加のみ (= 関数本体・依存・ハッシュは無改変)。

## 用途

Firefox の "ff2mpv" 拡張 (= ページ / リンクの URL を mpv に渡して再生) の
**native messaging host** (Rust 実装)。Firefox からは
`/usr/lib/mozilla/native-messaging-hosts/ff2mpv.json` を読んで
`/usr/bin/ff2mpv-rust` を起動 → host は stdin から JSON message (URL +
options) を受け取り → `mpv <args> <url>` を spawn する。

mpv は yt-dlp 連携済み (= `os_packages` で `mpv` `yt-dlp` 配備済み) のため、
YouTube / ニコニコ動画等の動画 URL を browser から右クリック →
「Play in mpv」 で直接 mpv に流せる。

Python upstream (`woodruffw/ff2mpv`) には GitHub Releases が無く、AUR にも
stable tag pin の Python 版は存在しない (`-git` 版のみ = nekono-pacman-
distribution の「`<name>-git` を避ける」規約に違反)。Rust 版 (`ryze312/
ff2mpv-rust`) は release tag を切る運用 + AUR の `ff2mpv-rust` (pkgver=1.1.7
pin) が利用可能なため、こちらを採用。

`provides=("ff2mpv-native-messaging-host-git" ...)` で AUR base pkg を
互換的に満たすため、Firefox extension 側 (= `ff2mpv@yossarian.net`) からは
manifest path 経由で透過に動く。

## Source

- AUR: https://aur.archlinux.org/packages/ff2mpv-rust
  - maintainer: eNV25 <env252525@gmail.com>
- Upstream: https://github.com/ryze312/ff2mpv-rust
  - owner: ryze312 (作成 2023-02-13、stars 87、forks 6)
  - AUR `url` フィールドと一致 (typosquat / なりすましリスクなし)
- Firefox extension 側 upstream (=本 host が受ける native message の発信側):
  https://github.com/woodruffw/ff2mpv (= extension ID `ff2mpv@yossarian.net`)

## 検証結果

- [x] `source` URL = `github.com/ryze312/ff2mpv-rust/archive/refs/tags/1.1.7.tar.gz`
  - upstream tag `1.1.7` → commit `14d663e14b5ca5a9106e9da5c171ae8044ef280c`
    (= author "Ryze <ryze@noreply.code.thishorsie.rocks>"、commit msg
    "Bump version: 1.1.6 -> 1.1.7"、date 2025-05-31)、AUR と同じ正規 release
- [x] `sha256sums` 独立検証
  - 実測: `0d6e3f36d585c463b9ef5b18042f9e47a867a119135072f5afd68b40ad3723bb`
    (= `curl -fsSL <url> | sha256sum` で tarball から計算)
  - PKGBUILD 値と一致
- [⚠] Tag `1.1.7` の GPG 検証は未実施 (= upstream は commit signing 運用なし)。
      tarball sha256 pin + GitHub release URL で integrity 確保
- [x] `prepare()`: `cargo fetch --locked --target ...` — `Cargo.lock` に従って
      依存を取得 (= reproducible)。`--locked` で lock file の改変防止
- [x] `build()`: `cargo build --frozen --release --all-features` — `--frozen`
      で build 中の Cargo.lock 改変 + network fetch を禁止 (= prepare で
      fetch 済みの依存だけで build 完結)
- [x] `build()` 内 perl 置換:
      `"$exe" manifest | exe="$exe" perl -pe 's|\Q$ENV{exe}\E|/usr/bin/ff2mpv-rust|g' >manifest.json`
      — Rust binary 自身が出力する manifest JSON 内の `"path"` 値を、
      build host の絶対 path から install 後の `/usr/bin/ff2mpv-rust` に
      書き換えるだけ。`\Q...\E` で metachar escape 済み、injection なし
- [x] `package()`: `install -Dm0755 target/release/...` (= binary) と
      `install -Dm644 manifest.json` (= manifest を `usr/lib/mozilla/
      native-messaging-hosts/ff2mpv.json` 等に配置)。標準 install のみ
- [x] `depends`: 無し (= Rust binary は static + system libc/gcc-libs のみ。
      runtime の mpv は `optdepends=mpv` で表現)
- [x] `makedepends`: `cargo` のみ (= Arch では `rust` パッケージが
      `provides=cargo`、virtual provide 経由)
- [x] `optdepends`: `mpv` + 各 browser (firefox / chromium / vivaldi 等)。
      Firefox 限定運用なので不要な browser entry も並ぶが、optdepends は
      実 install に影響しない (= 情報表示のみ)
- [x] Rust source 監査 (= `src/{main,lib,command,browser,config,error}.rs`):
  - main.rs: arg dispatcher (help / manifest / manifest_chromium / validate /
    default = FF2Mpv)
  - browser.rs: Mozilla Native Messaging Protocol (LE u32 length prefix +
    JSON body) の単純実装。stdin read、stdout write のみ
  - command.rs `ff2mpv()`: stdin から `{url, options}` JSON を読む →
    `process::Command::new(config.player_command).args(args).arg(url).spawn()`
    で mpv を起動。`process::Command` 経由 (= execve 直接呼び、shell 不介在)
    なので URL に special char が混じっても shell injection 不能
  - `detach_mpv` (Unix): `process_group(0)` で mpv を別 process group へ
    分離 (= Firefox が host を kill しても mpv は生き続ける、MDN の
    Native Messaging 仕様準拠)
  - `Cargo.toml` 依存: `serde 1.0.219` / `serde_json 1.0.140` / Windows-only
    `windows 0.61.1` の 3 つのみ、いずれも well-known crate
- [x] `secrets` 混入なし (`.git` + `.nvchecker.toml` (AUR ローカルの監視
      設定、本 repo では root `nvchecker.toml` で管理) を削除済み、秘密鍵 /
      token / `.gpg` ファイル等なし)
- [x] AUR との diff: 純 fork (= 機能部分は完全一致、`# Maintainer:` →
      `# Contributor:` + コメント追記のみ)

## 結論

**approve** — そのまま build host で `bin/build-all ff2mpv-rust` で build
+ sign + repo db 追加可。

これにより:
- `ansible-nekonodesk` の `roles/os_packages/vars/Archlinux.yml` の
  Self-hosted (nekono repo) block に `ff2mpv-rust` を追加できる
- `roles/firefox/templates/policies.json.j2` の `ExtensionSettings` に
  `ff2mpv@yossarian.net` を `force_installed` で auto-install させる
  (= AMO `https://addons.mozilla.org/firefox/downloads/latest/ff2mpv/latest.xpi`)

## 更新方針

upstream の新 release (1.1.8 等) が出たら:
1. AUR PKGBUILD の pkgver / sha256sums を確認
2. 本 dir の PKGBUILD を差し替え (`# Contributor:` 行 + nekono 改変
   コメントは維持)
3. sha256 を独立再計算 (= `curl -fsSL <url> | sha256sum` で照合)
4. tag commit author を GitHub API で確認 (= ryze312 が変わってないか)
5. `Cargo.toml` の依存 crate 一覧を確認 (= 新規依存があれば supply-chain
   review を追加)
6. REVIEW.md に確認日 + 結論 update + 更新履歴 row 追加

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | findings |
|---|---|---|---|
| 2026-05-25 | 1.1.7-1 | (初回 review) | AUR 純 fork、approve |

<!-- claude-review fast-path selftest (throwaway, PR は close する) -->
