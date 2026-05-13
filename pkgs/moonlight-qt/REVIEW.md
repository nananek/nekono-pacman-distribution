# moonlight-qt review

## 状態

**未 review** (PKGBUILD 未取得)。

## 想定 source

- AUR: https://aur.archlinux.org/packages/moonlight-qt
  ```
  git clone https://aur.archlinux.org/moonlight-qt.git pkgs/moonlight-qt
  rm -rf pkgs/moonlight-qt/.git
  ```
- Upstream: https://github.com/moonlight-stream/moonlight-qt

## Review チェックリスト

- [ ] `source` URL が `github.com/moonlight-stream/moonlight-qt/archive/v<ver>.tar.gz`
- [ ] `sha256sums` が upstream tag tarball と一致 (`curl -sL <url> | sha256sum`)
- [ ] `makedepends`: qt6-base / qt6-declarative / ffmpeg / sdl2 / opus / 等
- [ ] `build()` の qmake / make コマンドが標準的、追加 patch があれば中身確認
- [ ] PR #1648 (microphone pass-through) を別 patch として当てる予定があれば
      review 必須

## 結論

(review 完了後に記入)
