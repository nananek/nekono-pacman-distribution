# vesktop-bin review

## 状態

**未 review** (PKGBUILD 未取得)。

## 想定 source

- AUR: https://aur.archlinux.org/packages/vesktop-bin
  ```
  git clone https://aur.archlinux.org/vesktop-bin.git pkgs/vesktop-bin
  rm -rf pkgs/vesktop-bin/.git
  ```
- Upstream: https://github.com/Vencord/Vesktop (GitHub releases に AppImage
  と AppImage.sig が出てる)

## Review チェックリスト (実施時に埋める)

- [ ] `source` URL が `github.com/Vencord/Vesktop/releases/...` を指している
- [ ] `sha256sums` (or `b2sums`) が upstream release artifact と一致 (= `pkgver`
      に対応する tag の AppImage hash)
- [ ] `build()` / `package()` で外部 curl / wget / git clone が発生していない
- [ ] `depends` が想定 (electron, libxxx 系) 通り、余計な物無し
- [ ] AppImage の中身を一度展開して `.desktop` ファイルや main binary path を
      確認 (任意、深く調べる時のみ)

## 結論

(review 完了後に記入: 受入 / 改変要 / 却下、改変ありなら diff)
