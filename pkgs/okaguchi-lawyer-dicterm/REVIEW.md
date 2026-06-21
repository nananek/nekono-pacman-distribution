# okaguchi-lawyer-dicterm — Security Review

## 状態

**review 済み、approve** (2026-05-21、初回 add)

自家 PKGBUILD (AUR 不在)。 upstream: https://togilab.com/dicterm/

## 用途

岡口基一裁判官が配布していた法律用語辞書を Ryosuke Togi (Ease Up Ltd.) が
現代 IME 向けにリメイクしたデータを Mozc ユーザ辞書に登録するための辞書データ + import helper。
ayaka (Arch desktop) で日本語入力時に法律用語を変換候補に出せるようにする。

## Source

- Upstream: https://togilab.com/dicterm/
  - Ryosuke Togi / Ease Up Ltd.
  - バージョン管理なし。ページの「最終更新日」2020-11-22 を `pkgver=20201122` として採用
  - download URL: `https://togilab.com/download/dictermfile/?wpdmdl=150`
    - WordPress Download Manager (WPDM) plugin。`wpdmdl=150` = post ID、認証・refresh token 不要で直接 GET 可能
- AUR: 不在 (= 自家 PKGBUILD)

## ライセンス方針

ページには「© 2020 Ryosuke Togi / Ease Up Ltd.」のみ記載。MIT/CC 等の明示的 OSS ライセンスなし。
以下の理由により私的使用の範囲と判断し、[nekono] Tailscale 内配信に限定する:

- 配信先は自家 Tailscale network 内 (= 所有デバイスのみ)。著作権法上の「公衆送信」に当たらない
- 転売・再配布・公開 repo への push は行わない
- 将来作者から削除要請があれば即座に対応する

## 検証結果

- [x] `source` URL: `togilab.com/download/dictermfile/?wpdmdl=150` — 公式配布 URL、typosquat なし
- [x] `sha256sums` 独立検証
  - 辞書 TXT 実測: `1ddd61bdd62f1bc6b229efb2c957f47216872cac69e1660b08b109d62fcf4eef`
  - script 実測: `cf0506c0400a0f7d8f48562e6067e1ea988b5e6629bb2906305356c694162c70`
  - proto 実測: `b38529be56a93c215d58b931d367a822e26d88c70949e4fd870d4de4521fb373` (genshin-dict と同一)
  - 全て PKGBUILD 記載値と一致
- [x] `build()`: `protoc --python_out` のみ。ネットワーク取得なし
- [x] `package()`: `install` のみ。curl / exec / 外部 fetch なし
- [x] `depends`: `python` / `python-protobuf` — 全て Arch 公式 extra
- [x] `makedepends`: `protobuf` (protoc) — Arch 公式 extra
- [x] `secrets` 混入なし
- [x] 同梱 `okaguchi-lawyer-mozc-import` (Python):
  - ローカルファイルのみ操作
  - `sys.path.insert` で `/usr/lib/okaguchi-lawyer-dicterm` を追加、`user_dictionary_storage_pb2` を import
  - injection なし、ネットワーク取得なし
- [x] 同梱 `user_dictionary_storage.proto`:
  - genshin-dict と同一ファイル (sha256 一致)。`google/mozc` の BSD-3-Clause schema 定義のみ

## 辞書データ詳細

- ファイル形式: Microsoft IME Dictionary Tool 98 形式 (Shift-JIS / cp932、BOM なし、CRLF)
- コメント行: `!` で始まる行 (例: `!Microsoft IME Dictionary Tool 98`)
- データ構造: `reading\tword\tpos` の 3 列 TSV
- エントリ数: 1,510 件
- POS 種別: `さ変名詞` のみ → Mozc PosType enum 10 (`名詞サ変`)
  - 法律用語の多くはサ変名詞 (訴訟する・上告する 等) であり適切なマッピング

## 設計判断

### genshin-dict との共通部分

protobuf 直接操作パターンは genshin-dict と同一:
- `user_dictionary_storage.proto` (Mozc BSD-3-Clause) を vendor し `protoc --python_out` でビルド時 compile
- `user_dictionary_storage_pb2.py` を `/usr/lib/okaguchi-lawyer-dicterm/` に install
- import script が DB を read/write、`pgrep/pkill mozc_server` で file lock 競合を回避
- `$XDG_DATA_HOME/mozc/.okaguchi-lawyer-dicterm-{pkgver-pkgrel}-imported` flag file で冪等制御

### genshin-dict との差分

| 項目 | genshin-dict | okaguchi-lawyer-dicterm |
|---|---|---|
| エンコーディング | UTF-16 LE + BOM (`decode('utf-16')`) | Shift-JIS (`decode('cp932')`) |
| コメント行スキップ | `line.startswith('#')` | `line.startswith('!')` |
| POS map | 44 種類フルマップ | `{'さ変名詞': 10}` のみ |
| source URL | GitHub Release tarball (stable) | WordPress WPDM `?wpdmdl=150` |

### nvchecker 非登録

ページに version がなく、URL にも version が含まれない。
更新時は手動で sha256 を再計算し pkgrel を +1 する。

## 結論

**approve** — build host で `bin/build-all okaguchi-lawyer-dicterm` で build + sign + repo db 追加可。

## 更新方針

upstream で更新があれば:
1. `curl -fsSL "https://togilab.com/download/dictermfile/?wpdmdl=150" | sha256sum` で新 hash 確認
2. PKGBUILD の `pkgver` (新 YYYYMMDD)、`sha256sums[0]` を更新
3. `pkgrel=1` にリセット
4. REVIEW.md 「更新履歴」に 1 行追記

## 更新履歴

| 日付 | pkgver | review した PKGBUILD repo SHA | findings |
|---|---|---|---|
| 2026-05-21 | 20201122 | daac685 | 初回 add (pkgrel=1)。辞書 sha256 実測確認済み |
| 2026-06-03 | 20201122-2 | (this PR) | `pkgrel` +1 (deps changed): protobuf / python-protobuf 34.1-1 → 35.0-1 |
| 2026-06-15 | 20201122-3 | (this PR) | `pkgrel` +1 (deps changed): protobuf / python-protobuf 35.0-1 → 35.0-2 |
| 2026-06-21 | 20201122-4 | (this PR) | `pkgrel` +1 (deps changed): python 3.14.5-1 → 3.14.6-1 |
