# phase: build

`bin/build-all --pending` で署名付き build を行う phase。手順は skill `daily-ops` の
Phase 3 (`references/build.md`)。ただし**plan の解釈と実行だけ**があなたの仕事で、
成否の判定は harness の `build_verify` が行う。

1. `build_plan` で計画を取る。`pending` が空で `retire` も空なら、build せず
   `finish(status="ok", summary="nothing to build")`。`pending` が空で `retire` がある場合は
   退役反映のために `build_run` を実行する (db から外すため)。
2. `disk_free` を確認する。free が 10 GiB 未満なら `escalate(kind="content")` ではなく
   `finish(status="needs_human")` で止め、free/total を summary に書く (ENOSPC で
   artifact が 0600 のまま残った前例がある)。
3. `signing_probe` で署名経路を確認する。`ok == false` なら build を始めない
   (`finish(status="blocked")`)。署名は 1 pkg = 1 回 user の承認なので、承認が来ないまま
   走らせると `Failed to sign package file` で中途半端に終わる。
4. `build_run` を実行する (長い。進捗は harness が journal に流す)。
5. 結果 (`rc` / `built` / `error`) を読み、失敗があれば **原因の切り分けだけ**を行う:
   - `error` に `makepkg failed` → `fs_read` で該当 `pkgs/<pkg>/PKGBUILD` と REVIEW.md を読み、
     CLAUDE.md の落とし穴 (source の動的取得 / bundle の Python ABI / multi-volume 7z /
     自前 tarball の --mtime) のどれに当たりそうかを findings に書く。
   - `error` に `expected package file missing` → debug package の欠落は正常。それ以外は
     build 異常として findings に書く。
   - 署名失敗 (`署名に失敗` / rc=16) → 署名経路の問題。**PKGBUILD を直さない**。
   - checksum / PGP 検証失敗 → upstream が tarball を差し替えた可能性。
     `escalate(kind="content", subject="checksum mismatch")` で人間に渡す。
6. **あなたは file を編集しない** (`fs_write` はこの phase に無い)。PKGBUILD 修正が要るなら
   `finish(status="needs_human")` で「どの pkg の何を直すべきか」を findings に書く。
7. 最後に `finish(status="ok")`。summary には pending 件数・built の pkg と版・
   `cascade` の pkg を書く (cascade は自動 rebuild しない方針なので、PR が必要な旨を残す)。
