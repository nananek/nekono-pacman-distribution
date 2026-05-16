# vesktop-bin review

## 状態

**review 済み、approve** (2026-05-15)

AUR の `vesktop-bin` PKGBUILD (pkgver=1.6.5, pkgrel=1, _electronversion=40) を
fork。改変なし。

## Source

- AUR: https://aur.archlinux.org/packages/vesktop-bin
  - maintainer: zxp19821005
  - 同梱の `vesktop.sh` (= electron wrapper script、AUR 側で配布)
- Upstream: https://github.com/Vencord/Vesktop (= Vencord organization、
  GPL-3.0、8k stars、開発者 Vendicated)

## 検証結果

- [x] `source_x86_64` URL = `github.com/Vencord/Vesktop/releases/download/v1.6.5/vesktop-1.6.5.x86_64.rpm`
  - upstream 公式 release artifact、typosquatting / mirror spoof なし
- [x] `sha256sums_x86_64` が upstream .rpm と一致
  - 実測: `6fd5669fe2d5b1ad9bec10a8af1cf859bd911226528eae923ba70e6ef032c134`
  - PKGBUILD 値: `6fd5669fe2d5b1ad9bec10a8af1cf859bd911226528eae923ba70e6ef032c134`
  - 一致
- [x] `vesktop.sh` の sha256sums が現物と一致 (PR #5 review で末尾改行を
      追加した版、AUR original から sha 変化)
  - 実測: `4497d4c2cfb24ca0665cbeabf377a6bc850a8cfd6dd17469b0dc937a9ed6bf65`
  - PKGBUILD 値: `4497d4c2cfb24ca0665cbeabf377a6bc850a8cfd6dd17469b0dc937a9ed6bf65`
  - 一致
  - AUR original (末尾改行なし) の sha は `31ad33b6...` だった
- [x] Tag `v1.6.5` の git commit (`1f17c18362b12a9e8af21e571e844d05bda87ff5`) は
      **GPG verified** (= Vendicated = 主開発者署名)、tampering なし
- [x] `vesktop.sh` の内容 review:
  - electron@40 を `cd ${_APPDIR} && exec` で起動、`@runname@`/`@appname@`
    は prepare() で sed 置換 (=`app.asar`, `vesktop`)
  - `XDG_CONFIG_HOME` から `${_FLAGS_FILE}` を読んで flags 配列に展開
    (= `${XDG_CONFIG_HOME}/vesktop-flags.conf`、user が制御)、quote 適切で
    word splitting / injection なし
  - `--wayland` で Ozone Wayland flag を追加 (= cosmetic、optional)
  - root 起動時のみ `--no-sandbox` を追加 (= electron が root 拒否する
    一般的 hack、daemon-like 用途には許容)
  - curl / wget / eval / dynamic command 構造なし
- [x] `prepare()`: vesktop.sh の placeholder を sed 置換 + `.desktop` の
      `/opt/Vesktop/` prefix を削る。`_get_electron_version` で `strings`
      使って binary 中の Electron バージョンを表示 (= info 用、無害)
- [x] `package()`:
  - `/usr/bin/vesktop` ← vesktop.sh (= wrapper)
  - `/usr/lib/vesktop/app.asar` ← .rpm 内の `/opt/Vesktop/resources/app.asar`
  - `/usr/share/applications/vesktop.desktop`
  - `/usr/share/icons/hicolor/scalable/apps/vesktop.svg`
  - すべて install -Dm 標準コマンド
- [x] `depends`: `electron40`, `debugedit` — system-wide electron を使う
      (= 独自 electron 同梱しない、supply chain plus、Arch の electron40
      package の更新追従に依存)
- [x] license `GPL-3.0-only` — upstream 一致

## 結論

**approve** — そのまま build host で `makepkg -s --sign --key 483D...` 可。

`.rpm` を source にして Vencord の release CI prebuilt をそのまま使う path。
独自 `electron` を同梱せず Arch の `electron40` package を再利用する判断は
SBom 観点でプラス。

## 更新方針

upstream の新 release (v1.7.x 等) が出たら:
1. AUR で pkgver / `_electronversion` / sha256sums の値を確認
2. 本 dir の PKGBUILD と vesktop.sh を差し替え
3. 両方の sha256 を独立再計算 (= `curl -fsSL <url> | sha256sum` + `sha256sum
   vesktop.sh`) で照合
4. tag commit が GPG verified か gh api で確認
5. `_electronversion` が `electron${N}` Arch package とマッチするか
   `pacman -Ss '^electron[0-9]+$'` で確認 (= 古い electron が drop されたら
   切り替え)
6. REVIEW.md に確認日 + 結論 update
