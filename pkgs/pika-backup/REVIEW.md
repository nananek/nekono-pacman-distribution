# pika-backup review

## 状態

**未 review** (PKGBUILD 未取得)。

## 想定 source

- AUR: https://aur.archlinux.org/packages/pika-backup
- Upstream: https://gitlab.gnome.org/World/pika-backup

## Review チェックリスト

- [ ] `source` URL が `gitlab.gnome.org/World/pika-backup/-/archive/v<ver>/...`
- [ ] `sha256sums` 一致
- [ ] `makedepends`: rust / cargo / gtk4 / libadwaita / meson / 等
- [ ] `build()` の meson / ninja / cargo step が標準的
- [ ] cargo lockfile を尊重しているか (vendor crate を offline build できる
      schema か)
- [ ] borg を内部呼び出しするので CVE 追跡対象 (borg 本体は extra から
      入る、pika 側は GUI フロント)

## 結論

(review 完了後に記入)
