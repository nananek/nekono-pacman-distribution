# genshin-dict — Security Review

## 状態

**review 済み、approve** (2026-05-21、初回 add)

自家 PKGBUILD (AUR 不在)。 upstream: https://github.com/kotofurumiya/genshin-dict
data-only package (tarball 展開 + ファイルコピーのみ)。

## 用途

原神 (Genshin Impact) の固有名詞 (キャラクター名・地名・アイテム名等) を
Mozc ユーザ辞書に登録するための辞書データ + import helper。
ayaka (Arch desktop) で日本語入力時に原神固有名詞を変換候補に出せるようにする。

## Source

- Upstream: https://github.com/kotofurumiya/genshin-dict
  - kotofurumiya (Kouto Furumiya)、MIT License
  - tag `v6.3.1` (GitHub Release)
  - distfile: `github.com/kotofurumiya/genshin-dict/archive/refs/tags/v6.3.1.tar.gz`
- AUR: 不在 (= 自家 PKGBUILD)

## 検証結果

- [x] `source` URL: `github.com/kotofurumiya/genshin-dict` の公式 Release tarball。
  typosquat / domain spoof なし
- [x] `sha256sums` 独立検証
  - 実測 (`curl -fsSL <url> | sha256sum`): `3759b9e5d575c17ac287ee079cfcfaf79e69805994f7c6eb35e4d89b6a7e79f0`
  - PKGBUILD に記録済み
- [x] `build()`: 不要 (data-only)。`package()` は `install` + `cp` のみ。
  curl / exec / 外部 fetch なし
- [x] `depends`: なし。`optdepends`: fcitx5-mozc のみ (import helper 用、任意)
- [x] `makedepends`: なし
- [x] `secrets` 混入なし
- [x] 同梱 `genshin-dict-mozc-import`: shell script、ローカルファイルのみ操作、
  ネットワーク取得なし、shell injection なし

## 設計判断

### mozc_user_dictionary_tool 非同梱問題

`fcitx5-mozc` (Arch 公式 extra) は `/usr/lib/mozc/mozc_tool` (GUI) のみを含み、
コマンドライン `mozc_user_dictionary_tool` を含まない (`pacman -Ql fcitx5-mozc | grep user_dict` で確認済み、2026-05-21)。

`genshin-dict-mozc-import` スクリプトはこれを graceful に処理する:
- tool が存在する場合: `--mode=import_data --strategy=merge` で自動 import + flag file 作成 (冪等)
- tool が存在しない場合: exit 1 + 手動 import 手順を stderr に出力

Ansible task は `failed_when: false` で呼ぶため、tool 不在でも playbook は通る。
手動 import: fcitx5-configtool → Mozc → ユーザ辞書 → ファイルから辞書をインポート
→ `/usr/share/genshin-dict/genshin-dict.txt`

将来 `mozc` (AUR 等) に `mozc_user_dictionary_tool` が含まれる場合は自動 import が機能する。

### 辞書ファイルの encoding

`原神辞書_Windows.txt` は UTF-16 LE + BOM (Windows IME 形式)。
`/usr/share/genshin-dict/genshin-dict.txt` にそのまま配置する
(Mozc GUI import は UTF-16 LE を受け付ける実績あり。mozc_user_dictionary_tool での
encoding サポートは tool 入手後に要確認)。

### arch=any

辞書データ (TSV) + shell script のみ。バイナリ不在。

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
| 2026-05-21 | 6.3.1 | (commit 後に記入) | v6.3.1 | 初回 add、自家 PKGBUILD (AUR 不在)。mozc_user_dictionary_tool 非同梱 確認済み |
