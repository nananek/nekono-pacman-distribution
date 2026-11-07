# phase: upstream

upstream の新版を **1 つだけ** bump する phase。対象 Issue 番号と pkg / 版は user
メッセージにある。手順は skill `daily-ops` の Phase 2 (`references/upstream-issue.md`)。

## 手順

1. `gh_issue` method=view で Issue を読む（title `[<pkg>] upstream version: <ver>`）。
   同じ pkg の Issue が複数あれば、**最新版に集約して 1 PR** にしてよい
   （実在・安定を確認してから）。
2. `fs_read` で `pkgs/<pkg>/REVIEW.md` を読む。**pkg 固有の検証方法はここにしかない**
   （前回と同じ方法を取る。AUR との意図的 diff があれば維持する）。
3. 調査する:
   - `gh_repo` method=release_view で tag / publishedAt / assets を確認。source の
     domain・owner が従来と同じか、release author が同じか。
   - `fetch` で source URL を取得し **sha256 を実測**する（`sha256` フィールド）。
     upstream が公開する checksum と一致するか。`SKIP` は不可（pre-push gate が block）。
   - 差分: `fetch` save=true で旧新 tarball を落とし、`archive_diff` で
     `*/Makefile` `*/Cargo.lock` `*/pyproject.toml` `*/package.json` など
     build/install に効くファイルを比較する。
   - 依存: prebuilt binary は `binary_scan` で NEEDED と最大 GLIBC、URL 集合を旧新で比較。
     `gh_repo` method=compare で lockfile / manifest の変化を確認。
   - AUR の PKGBUILD は `fetch` で
     `https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=<pkg>` を読んで diff。
   - CLAUDE.md の「PKGBUILD でよく踏む落とし穴」#1–7 に触れないか。
4. verdict を決める:
   - **safe-to-bump**（pkgver + sha256 の機械的更新）→ 5 へ
   - **needs-attention**（build/depends の手当て。既存 pkg を今まで通り build できる範囲）
     → PKGBUILD を直して 5 へ
   - **block**（供給網の懸念 / breaking change）→ **commit しない**。
     `escalate(kind="content", subject=..., evidence=[URL と sha256], source_issue=N)`
     で人間に渡し、`finish(status="needs_human")`。
   - **誤検知疑い**（nvchecker の regex 崩れ、pre-release 拾い）→ 根拠を findings に
     書いて `finish(status="needs_human")`（Issue は閉じない）
5. 編集する（`fs_write` で 1 file ずつ）:
   - `PKGBUILD`: pkgver 更新 / **pkgrel は 1 に戻す** / `sha256sums` / URL
   - `.SRCINFO`: `pkg_srcinfo` の出力で差分を確認して反映
   - `.deps.lock`: depends / makedepends を変えた時だけ（`# MISSING` 行は維持）
   - `REVIEW.md`: 冒頭の状態、検証結果の実測値、更新履歴 1 行
6. `pkg_verifysource` で checksum を検証する。失敗したら hint を読む
   （stale cache か upstream の差し替えか）。**cache の値に pin を合わせない**。
7. `step_bump_pr` で branch 作成 → 署名 commit → push → PR。
   `message` には「何を独立に確認したか」と attribution 行
   `Assisted-by: nekono-pacman-agent (<model>)`、`body` には
   **Issue ごとに `Closes #a, closes #b`**（最初の 1 件しか自動 close されないため、
   `Closes #a, #b` は不可）と調査結果 / Test plan を書く。
8. `gh_pr` method=diff で judgment review（source URL / build() の怪しい step / depends）。
9. 問題なければ `step_merge`。merge 後に `gh_issue` で全 Issue が CLOSED か確認し、
   `finish(status="ok")` に pkg・版・PR 番号・確認したことを書く。

## しないこと

- supply-chain の懸念を自分で判断して進めない（必ず `escalate(kind="content")`）。
- 新規 package の追加 / 撤廃 / source origin の変更 / install hook の新設は
  `escalate(kind="content")`（user 承認が要る）。
- `refresh` で取った値を Issue や AUR の記載から写さない。**実測値だけ**を使う。
