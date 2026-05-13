# claude-code review

## 状態

**未 review** (PKGBUILD 未取得)。

## 想定 source

- AUR: https://aur.archlinux.org/packages/claude-code (もしくは
  `claude-code-bin`、AUR で実際に登録されている方を採用)
- Upstream: npm registry の `@anthropic-ai/claude-code` package。
  Anthropic 公式 GPG key (`31DDDE24DDFAB679F42D7BD2BAA929FF1A7ECACE`) は
  Debian apt repo の signing key であって、npm tarball には signature が
  付かないので、tarball の sha256 と npm registry の published version で
  pin する。

## Review チェックリスト

- [ ] `source` が npm の official tarball
      (`registry.npmjs.org/@anthropic-ai/claude-code/-/...`) か、もしくは
      Anthropic 公式 `downloads.claude.ai` の path か (両方とも upstream)
- [ ] `sha256sums` が npm registry の published version の `dist.shasum`
      (SHA-1) や `dist.integrity` (SRI、SHA-512) と一致するか
      (`npm view @anthropic-ai/claude-code dist.tarball` 等で取得)
- [ ] AUR の `claude-code` と `claude-code-bin` の差: bin は prebuilt
      JS bundle を npm から取って install するだけ、無印は更にコンパイル
      step を入れる場合あり → bin 系を採用予定
- [ ] post_install hook で /usr/local/bin/claude-code への symlink が
      作られるか
- [ ] npm 依存 (= node_modules tree) の供給源: vendored か実行時 install か。
      bin 系は vendored 想定なのでそこで信頼境界が閉じる
- [ ] Anthropic 公式の GPG signing が無いので、tarball pin の sha が
      毎 release こちらで update する運用必須

## 結論

(review 完了後に記入)
