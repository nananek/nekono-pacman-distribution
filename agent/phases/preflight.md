# phase: preflight

日次 run の前提を確認する phase。**判断はしない**。run を続けてよいか止めるべきかの
材料を `finish` に載せる。探索はしない — 使うツールは 3 つまで。

1. `preflight` ツールで状態を取る (branch / dirty / hooksPath / gh / 署名 / build host / publish ずれ)。
2. `dirty` が空でない場合だけ `repo_state` で中身を見る。**dirty な file を 1 つずつ読まない**
   (未コミット作業の調査はこの phase の仕事ではない)。file 名から「user の作業」か
   「実行時 cache」かを判断し、判断できなければ dirty の一覧をそのまま summary に書く。
3. 次のいずれかなら `finish(status="blocked")` で止める:
   - `signing_ok == false` (署名が無いと build も修正 commit も成立しない)
   - `gh_ok == false` (PR / Issue が読めないと triage が成立しない)
4. `build_host == false` は止まる理由ではない (driver が build/publish を skip する)。summary に明記する。
5. `publish` が `pending` なら「build 予定 0 件でも publish が要る」ことを summary に書く。

`finish` の summary には、branch / dirty 件数 / bot PR と upstream Issue の番号 / 署名 /
build host (と `build_host_detail`) / publish の 6 点を 1 行ずつ書く。`findings` は気づいた点だけ
(dirty の内容、hooksPath 未設定、古い branch 等)。

## escalate の使い分け (重要)

- **run を続行できるなら、改善要望は `finish(status="ok")` の findings に書く**。
  例: 「preflight の JSON に〜が無いと切り分けできない」は harness への要望であって、
  この phase の実行を止める理由ではない。
- `escalate(kind="harness")` は **この phase の仕事が物理的に実行できない** ときだけ
  (必要なツールが存在しない、応答が解釈できない、など)。使うと run 全体が止まる。
- `escalate(kind="content")` は世の中の判断 (供給網・想定外の配布形式など)。repo に Issue が立つ。
