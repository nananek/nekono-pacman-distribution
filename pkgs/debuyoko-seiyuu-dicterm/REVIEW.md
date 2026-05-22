# debuyoko-seiyuu-dicterm review

## 状態

**review 済み、 approve** (2026-05-22、 初回 add)

自家 PKGBUILD (AUR 不在)。声優人名 MS-IME ユーザ辞書 (debuyoko.com) を
Mozc IME に import するパッケージ。

## 用途

fcitx5-mozc ユーザが声優名を IME で変換できるようにする。
`debuyoko-seiyuu-mozc-import` を postinstall / postupgrade で自動実行。

## Source

- 配布元: https://debuyoko.com/1388
  - 作者不詳 (味噌煮込み氏の Perl スクリプトを参考に作成と記載)
  - エントリ数: 9238 件 (2026-04-05 時点)
  - ライセンス明示なし → `LicenseRef-debuyoko-proprietary` で処理
  - Wikipedia 男性/女性声優カテゴリから抽出
- ダウンロード URL: `https://debuyoko.com/?sdm_process_download=1&download_id=1398`
  - WordPress Download Manager 経由、認証不要で直 DL 可
- AUR: 該当なし

## 検証結果

- [x] `source` URL = `debuyoko.com/?sdm_process_download=1&download_id=1398`
  - 作者直営サイト
- [x] `sha256sums` 独立検証
  - 実測: `6a1bf397309a84eed84111b6faf82a89904f87c7a13d4afac9d18f8e62723519`
- [x] ファイル形式確認
  - UTF-16 LE + BOM、CRLF 改行、TSV 3 列 (よみ / 表記 / 人名)
  - コメント行なし、全 9238 行が有効エントリ
  - POS は `人名` のみ → Mozc `PERSONAL_NAME` (enum 5)
- [x] `build()`: `protoc --python_out` のみ (genshin-dict / okaguchi 同パターン)
- [x] `package()`: install -Dm のみ、curl / wget / eval なし
- [x] `debuyoko-seiyuu-mozc-import`:
  - genshin-dict-mozc-import を template に作成
  - 差分: DICT_FILE / DICT_NAME / pkgname / flag prefix のみ
  - encoding: `utf-16` (genshin-dict 同様、BOM 自動 strip)
  - コメントスキップ: `#` のみ (ファイルにコメント行なし、実害なし)
  - fallback POS: `人名` (5) (全エントリが人名のため)
  - `secrets` 混入なし
- [x] `user_dictionary_storage.proto`: genshin-dict と同一ファイル (BSD-3-Clause / google/mozc)
- [x] `arch=('any')` — pure Python + protobuf 生成

## ライセンス方針

ライセンス明示なし。Wikipedia データ派生 (CC BY-SA 4.0)。PKGBUILD を書くだけ
(= インストーラを提供) であり、辞書データ自体は配布元から直接取得するため
[nekono] としての再配布には当たらない (genshin-dict / okaguchi-lawyer-dicterm
と同じ方針)。

## AUR との意図的差分

AUR に該当 pkg なし。完全自家 PKGBUILD。

## 結論

**approve** — build host で `bin/build-all debuyoko-seiyuu-dicterm` で build +
sign + repo db 追加可。

## 更新方針

debuyoko.com のページ (https://debuyoko.com/1388) に「※N件　YYYY/M/D」の
形式で更新日が記載される。変化を検知したら:
1. `download_id=1398` の URL から新ファイルを取得
2. sha256 を再計算
3. pkgver を `YYYYMMDD` 形式で更新、pkgrel=1 にリセット
4. REVIEW.md 更新履歴に 1 行追記

nvchecker では URL からの date 抽出が難しいため、手動確認運用とする。

## 更新履歴

| 日付 | release | review した PKGBUILD repo SHA | findings |
|---|---|---|---|
| 2026-05-22 | 20260405-1 | (本 PR commit SHA) | 初回 add、自家 PKGBUILD (AUR 不在)、声優人名 9238 件、UTF-16 LE |
