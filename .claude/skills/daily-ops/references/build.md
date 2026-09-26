# Phase 3: build (`bin/build-all --pending`)

前提: このマシンが build host (`SKILL.md` の preflight で `BUILD_HOST`)。違えば user にコマンドを提示して終了。
`bin/build-all` は pkgs/* を **[nekono] 内 cross-dep の topological order** で makepkg → Nekono GPG 署名 →
`repo-add --sign` → `bin/prune` まで一気に行う。

## 1. 計画を見る

```sh
git pull --ff-only
bin/build-all --pending --dry-run
```

| 出力 | 意味 |
|---|---|
| `[pending] <pkg> <ver>` | build 対象 (repo/x86_64/ に `<pkg>-<ver>-*.pkg.tar.zst` が無い) |
| `[skip] ... already built` | build 済み |
| `[cascade-warn] <pkg> ...` | 上流の [nekono] pkg が今 rebuild されるので、この pkg も pkgrel +1 の PR が要る (自動 rebuild はしない方針) |
| `[*] Will retire N package(s)` | PKGBUILD が消えた pkg を db から外す (撤廃。`--pending` と無関係に常に走る) |
| `[OK] nothing to build` | 何もすることが無い (ただし「build 済みだが未 publish」の可能性は `publish.md` で確認) |

## 2. 実行前チェック

- 署名の readiness: `signing.md` (rogue agent 無し / forward socket あり / `gpg --card-status` で card が見える)。
- `df -h .` — 大型 build (moonlight-qt-nekono / sunshine-* / voicevox-bin 等) は作業領域を食う。空きが少ないなら先に user に伝える
  (ENOSPC で artifact が 0600 のまま残り client が 403 になった前例があり、build-all は chmod 0644 で補正している)。
- makepkg `-s` は makedepends を `pacman -S` する (sudo)。sudo / pacman の prompt で止まったら user に伝える。

## 3. 実行 — background + **本当の exit code を捕捉**

```sh
LOG=$(mktemp -t nekono-build-XXXXXX.log)     # repo の外。scratchpad があればそこ
( bin/build-all --pending > "$LOG" 2>&1; rc=$?; echo "build-all exit=$rc" >> "$LOG"; exit $rc )
```

これを `run_in_background` で実行する。`; echo ...` を subshell の外に置くと通知の exit code が
`echo` のもの (常に 0) になり、build-all が失敗していても「exit 0」と通知される
(実際に「完了」と誤報告した前例あり)。build は pkg によって長い (数十分以上)。sleep で poll せず、完了通知を待つ。
待っている間に `gpg` を叩かない。

## 4. 完了検証 — exit code を信じず、log と artifact で確認

```sh
tail -n 30 "$LOG"                        # `[OK] build-all completed. N signed packages` と `build-all exit=0`
```

`[pending]` だった各 pkg について:

```sh
ls -l --time-style=long-iso repo/x86_64/<pkg>-<ver>-*.pkg.tar.zst{,.sig}     # 両方あり、mtime が今回
LC_ALL=C gpg --verify repo/x86_64/<file>.pkg.tar.zst.sig repo/x86_64/<file>.pkg.tar.zst 2>&1 | grep -E 'Good signature|BAD'
LC_ALL=C gpg --verify repo/x86_64/nekono.db.tar.gz.sig repo/x86_64/nekono.db.tar.gz 2>&1 | grep -E 'Good signature|BAD'
tar -xzO -f repo/x86_64/nekono.db.tar.gz --wildcards '*/desc' \
  | awk '/^%FILENAME%$/{getline f; print f}' | grep '^<pkg>-'                  # db に新 version が載っている
bin/build-all --pending --dry-run                                               # 最後に `nothing to build` になる
```

全部通って初めて「build 成功」と報告し、Phase 4 (publish) へ進む。

## 5. 失敗時

| 症状 | 対応 |
|---|---|
| exit 16 / `署名に失敗しました` / signing failed | `signing.md` の復旧手順。直ったら `bin/build-all --pending` を再実行 (冪等) |
| makepkg の checksum / PGP 検証失敗 | upstream が tarball を差し替えた可能性。**sums を機械的に上書きせず**、supply-chain 案件として user に報告 |
| makedepends 不足 / upstream の API 変更 | CLAUDE.md「PKGBUILD でよく踏む落とし穴」#1–7 を確認。`( cd pkgs/<pkg> && makepkg --verifysource )` で切り分け。修正するなら fix PR (REVIEW.md 更新履歴に `X.Y.Z (build fix)`)。needs-attention の範囲を超えるなら停止 |
| 1 pkg だけ落ちる | `bin/build-all <他 pkg>` で他は進められる。ただし cross-dep の下流に注意 |
| CPU だけ食って object が出来ない | 自前 tarball の `--mtime` が未来 (JST で UTC 固定値) の qmake/autotools 再生成ループ。CLAUDE.md #7 |
| 途中で `expected package file missing` | makepkg が期待 file を作れていない。log の直前を読む (`-debug` pkg の欠落は許容される) |

## 6. cascade-warn の扱い

`[cascade-warn] <pkg>` が出たら、その pkg は上流 [nekono] pkg の rebuild に合わせた pkgrel +1 と rebuild が要る。
**自動では PR を作らず**最終報告で user に提示する (依頼があれば bot PR と同じ形で pkgrel bump PR を pkg ごとに 1 本)。
