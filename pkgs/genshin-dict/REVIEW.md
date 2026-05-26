# genshin-dict — Security Review

## 状態

**review 済み、approve** (最新: 2026-05-26 / 6.6.0)

自家 PKGBUILD (AUR 不在)。 upstream: https://github.com/kotofurumiya/genshin-dict
data (TSV 辞書) + Python import helper スクリプトを同梱。

## 用途

原神 (Genshin Impact) の固有名詞 (キャラクター名・地名・アイテム名等) を
Mozc ユーザ辞書に登録するための辞書データ + import helper。
ayaka (Arch desktop) で日本語入力時に原神固有名詞を変換候補に出せるようにする。

## Source

- Upstream: https://github.com/kotofurumiya/genshin-dict
  - kotofurumiya (Kouto Furumiya)、MIT License
  - tag `v6.3.1` (GitHub Release、commit `6bffdf3da2d386117f941df2f483d48aafb6ace3`)
  - distfile: `github.com/kotofurumiya/genshin-dict/archive/refs/tags/v6.3.1.tar.gz`
- AUR: 不在 (= 自家 PKGBUILD)

## 検証結果

- [x] `source` URL: `github.com/kotofurumiya/genshin-dict` の公式 Release tarball。
  typosquat / domain spoof なし
- [x] `sha256sums` 独立検証
  - tarball 実測: `3759b9e5d575c17ac287ee079cfcfaf79e69805994f7c6eb35e4d89b6a7e79f0`
  - script 実測: `9a513e893d2334d6ba13812b49932a317a16c4d01973d2d6e37089fe1f2d47c9`
  - proto 実測: `b38529be56a93c215d58b931d367a822e26d88c70949e4fd870d4de4521fb373`
  - 全て PKGBUILD 記載値と一致
- [x] `build()`: `protoc --python_out` のみ。ネットワーク取得なし
- [x] `package()`: `install` のみ。curl / exec / 外部 fetch なし
- [x] `depends`: `python` / `python-protobuf` — 全て Arch 公式 extra
- [x] `makedepends`: `protobuf` (protoc) — Arch 公式 extra
- [x] `secrets` 混入なし
- [x] 同梱 `genshin-dict-mozc-import` (Python): ローカルファイルのみ操作。
  `sys.path.insert` で `/usr/lib/genshin-dict` を追加、`user_dictionary_storage_pb2` を
  import。injection なし、ネットワーク取得なし
- [x] 同梱 `user_dictionary_storage.proto`:
  `google/mozc` の `src/protocol/user_dictionary_storage.proto` を verbatim vendor。
  BSD-3-Clause (Mozc 本体ライセンス)、コードなし (schema 定義のみ)

## 設計判断

### mozc_user_dictionary_tool 非同梱問題と代替実装

`fcitx5-mozc` (Arch 公式) は `mozc_user_dictionary_tool` CLI を含まない
(`pacman -Ql fcitx5-mozc | grep user_dict` で確認済み、2026-05-21)。

upstream `google/mozc` のソースを調査した結果、`mozc_user_dictionary_tool`
は現行バージョンのビルドターゲットに存在しない。`mozc_tool` の全 `--mode` は
Qt GUI launch のみでヘッドレス実行不可 (`mozc_tool_libmain.cc` 参照)。

代替として **protobuf 直接操作** を採用:
- `user_dictionary_storage.proto` (Mozc BSD-3-Clause) を vendor し、
  `protoc --python_out` でビルド時に `user_dictionary_storage_pb2.py` を生成
- `genshin-dict-mozc-import` が `/usr/lib/genshin-dict/user_dictionary_storage_pb2.py`
  を import し `~/.config/mozc/user_dictionary.db` を read/write
- 書き込み前に `pgrep/pkill mozc_server` で file lock 競合を回避
  (mozc_server は fcitx5 が次回使用時に自動再起動)
- 冪等: `$XDG_DATA_HOME/mozc/.genshin-dict-imported` flag file で制御

### user_dictionary.db のパス

Mozc は以下の優先順位で profile dir を決定 (`system_util.cc` 参照):
1. `~/.mozc/` が存在する場合 (後方互換)
2. `$XDG_CONFIG_HOME/mozc/`
3. `~/.config/mozc/` (XDG default)

スクリプトも同じロジックで解決する。

### 辞書 TSV の encoding

`原神辞書_Windows.txt` は UTF-16 LE + BOM (Windows IME 形式)。
Python `decode('utf-16')` で BOM 自動検出・strip して UTF-8 str に変換。
`/usr/share/genshin-dict/genshin-dict.txt` にそのまま配置 (バイナリ保存)。
スクリプト内でのみ decode しているため、辞書ファイル自体は encoding に関係なく動作する。

### POS マッピング

genshin-dict v6.3.1 で実際に使用されている POS は 3 種類:
`名詞` → NOUN(1)、`人名` → PERSONAL_NAME(5)、`地名その他` → PLACE_NAME(9)。
スクリプトは Mozc proto の全 44 POS type をマッピングし、将来の追加にも対応。

### arch=any

辞書データ (TSV) + Python スクリプト + proto schema のみ。バイナリ不在。

## 結論

**approve** — build host で `bin/build-all genshin-dict` で build + sign + repo db 追加可。

## 更新方針

upstream で新 release が出たら:
1. PKGBUILD の `pkgver` / `sha256sums[0]` を更新
2. `makepkg --verifysource` で sha256 確認
3. REVIEW.md 「更新履歴」に 1 行追記

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | upstream tag commit | findings |
|---|---|---|---|---|
| 2026-05-21 | 6.3.1 | 4916106 (49161063616031108f3e175baaf8c3890ada8189) | v6.3.1 (6bffdf3da2d386117f941df2f483d48aafb6ace3) | 初回 add (pkgrel=1)。mozc_user_dictionary_tool 非同梱確認済み |
| 2026-05-21 | 6.3.1 | 96374fc | v6.3.1 (6bffdf3da2d386117f941df2f483d48aafb6ace3) | pkgrel=2: protobuf 直接操作に移行。user_dictionary_storage.proto vendor + genshin-dict-mozc-import を Python 実装に書き替え |
| 2026-05-26 | 6.6.0 | (this PR) | v6.6.0 (`4188a9d624083c992cab9ea9a0234a9f3a8b0f0d`) | safe-to-bump: 辞書エントリ追加のみ (アファールの鳥、アムリタ学院、ヴァフマナ等)。PKGBUILD 構造変化なし、sha256 独立検証済み |
| 2026-05-21 | 6.3.1 | (本 commit 後に確定) | v6.3.1 (6bffdf3da2d386117f941df2f483d48aafb6ace3) | pkgrel=3: バージョン stamp 付き flag file に変更 (`.genshin-dict-{pkgver}-imported`)。upgrade 時に自動再 import、古い flag を cleanup |
