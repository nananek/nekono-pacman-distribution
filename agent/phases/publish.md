# phase: publish

署名済み `repo/x86_64/` を配信ホストへ送り、client と同じ経路で照合する phase。
手順は skill `daily-ops` の Phase 4 (`references/publish.md`)。

1. `preflight` ツールで `publish` の状態を見る。`up to date` なら配信は不要なので
   `finish(status="ok", summary="publish: up to date (skip)")`。
2. 前提を確認する:
   - `preflight.build_host == false` → ここでは publish できない。`finish(status="blocked")`。
   - `dist_publish` が `rclone not found` / `[ERROR] NEKONO_DAV_* is not set` を返した →
     **黙って skip しない**。何が足りないかを summary に書き `finish(status="needs_human")`。
3. `dist_publish` を実行する。2 pass 構成 (sync → copy --ignore-times) は不変。
   - `ok == false` は失敗。`error` と `pass` を findings に書いて `finish(status="failed")`。
   - `deleted > 0` は退役 pkg の掃除。件数を summary に書く。
4. `dist_verify` で照合する。`files` には必ず `nekono.db.tar.gz` と `nekono.db.tar.gz.sig` を
   含め、今回 build した pkg の `.pkg.tar.zst` と `.sig` も加える。退役分は `retired` に入れる。
   - `ok == true` が成功条件。**db と sig の両方が MATCH で初めて成功**と言える。
   - MISMATCH が残ったら `finish(status="failed")`。値や host 名は書かない。
5. `publish.env` の値は絶対に読まない・出力しない (`fs_read` でも deny される)。
6. `sudo pacman -Sy` による client 検証はしない。最後に「user が client で
   `sudo pacman -Sy && pacman -Si <pkg>` を確認する」ことを summary に残して `finish(status="ok")`。
