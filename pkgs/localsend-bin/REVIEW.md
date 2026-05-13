# localsend-bin review

## 状態

**未 review** (PKGBUILD 未取得)。

## 想定 source

- AUR: https://aur.archlinux.org/packages/localsend-bin
- Upstream: https://github.com/localsend/localsend (GitHub releases に
  Linux tarball あり)

## Review チェックリスト

- [ ] `source` URL が `github.com/localsend/localsend/releases/download/v<ver>/...`
- [ ] `sha256sums` が release artifact と一致 (`curl -sL <url> | sha256sum`)
- [ ] `package()` での install path / .desktop / icon 配置が標準的
- [ ] depends に flutter runtime 系 (libapp.so 用 GTK / GLib / etc.) が
      列挙されているか

## 結論

(review 完了後に記入)
