# nekono-pacman-agent

`daily-ops` を回すための harness。OpenCode の TUI/CLI を VM の中で動かす代わりに、
**閉じたツール集合 + 決定的な driver** で日次 run を実行する。

## なぜ OpenCode を置き換えるか

- `opencode run` は終了するまで自分の進捗を出さない。run の途中で何が起きているか
  見えず、「rc=0 なのに署名で落ちていた」を検出できなかった。
- 権限は `shell:<pattern>` の文字列マッチで、1 segment でも allow 外だと**呼び出し全体が
  拒否**される (2026-10-29 の publish 未実行)。sandbox ではない。
- phase 分離を subagent でやっていたが、進行は LLM の判断に委ねられていた。

この harness では、進行・検証・照合はすべて Go。LLM は**判断が要る箇所だけ**呼ぶ。
ツールは typed function のみで、argv を組み立てるのは Go。モデルは JSON しか渡せない。

## 構造

```
cmd/nekono-pacman-agent/   main (phase / run / doctor)
internal/llm/              OpenAI 互換 client (SSE streaming, tool calls, retry)
internal/tool/             ツール実装 + 許可 (closed set)
internal/phase/            phase loop (prompt + tool subset + step budget)
internal/journal/          ndjson イベント + 秘匿値の redaction
internal/cfg/              agent/agent.json
agent/rules.md             全 phase 共通の規約
agent/phases/*.md          phase ごとの prompt
```

## phase とツール

| phase | ツール | 判断 / 決定論 |
|---|---|---|
| preflight | preflight, repo_state, repo_diff, repo_log, fs_read | 状態の読み取り。blocked 条件の判定 |
| triage | gh_pr, gh_issue, pkg_3points, pkg_srcinfo, step_bot_pr, step_merge, fs_read/write, repo_* | bot PR 1 本の 3 点検証と merge 可否 |
| upstream | gh_repo, gh_issue, fetch, archive_*, binary_scan, pkg_srcinfo/verifysource, step_bump_pr, step_merge | 新版の調査・verdict・PKGBUILD 改修 |
| build | build_plan, build_run, signing_probe, signing_verify, disk_free | 失敗の切り分け。成否は `build_verify` が決める |
| publish | dist_publish, dist_verify, preflight | 失敗時の報告。一致は `dist_verify` が決める |

`run` の順序 (コードで固定):

```
preflight(tool) → check-upstream / check-deps
  → triage  × open deps/* PR（1 本 = 1 phase run）
  → upstream × open upstream Issue（1 本 = 1 phase run）
  → build_plan → build phase → build_verify
  → publish phase → dist_verify
  → report
```

triage / upstream は対象ごとに新しい会話で走る（phase ごとに `finish` で結論）。
`content` の escalate が出た対象は飛ばして次へ進み、run の status は
`ok_with_escalations` になる。`harness` の gap は run 全体を止める。

build plan は triage / upstream の後で計算し直す（merge 後の状態を見るため）。

検証用に、merge 系ツールを除いた config を渡せば副作用なしで phase を試せる:

```sh
nekono-pacman-agent phase triage -config agent/agent.test.json -task "対象 PR: #783 ..."
```

`fetch` はホスト allowlist 制（既定: github / aur / npm / pypi / crates）+ 対象 pkg の
`source=` に出てくるホスト。redirect は各 hop で再検査する。

## escalate の契約

- `escalate(kind="content")` — 世の中 (upstream / supply chain / 想定外の配布形式) の判断。
  **repo に Issue を 1 本だけ**立てる。fingerprint (`sha256(phase|pkg|subject)`) を本文に埋め、
  同じ fingerprint の Issue が既にあれば作らない (一度だけ報告)。
  `source_issue` があれば、その Issue に保留理由をコメントする。
- `escalate(kind="harness")` — ツール集合で表現できなかった (実装の穴)。
  **Issue は立てない**。journal に `capability_gap` を残し、run 全体を止める。
  レポートと Discord に載る。

## 設定

`agent/agent.json`:

- `model` / `base_url` / `api_key_file` — `base_url` はリポジトリに tailnet 名を置かないため
  空にし、`NEKONO_AGENT_BASE_URL` から取る。鍵はファイルパスのみ。
- `phases.<name>.tools` — その phase でモデルが呼べるツール。ここに無いものは呼べない
  (phase 分離はコードで強制)。
- `exec_timeout_sec` / `build_timeout_sec` / `max_steps`。

## 使い方

```sh
go build -o ~/.local/bin/nekono-pacman-agent ./cmd/nekono-pacman-agent
nekono-pacman-agent doctor                  # 前提チェック (値は出さない)
nekono-pacman-agent phase preflight         # 1 phase だけ
nekono-pacman-agent run --dry-run           # 計画だけ
nekono-pacman-agent run                     # preflight → detect → build → publish
```

出力は 2 本: **stdout に ndjson イベント** (host が tail する)、**stderr に人間向けログ**。

## bin/* との関係

`bin/*` は機械層としてそのまま。agent は次の `--json` を受け取る:

```
bin/preflight --json
bin/build-all --pending [--dry-run] --json
bin/publish --json
bin/check-upstream [--dry-run] --json
bin/check-deps [--dry-run] --json
```

`--json` のとき stdout は 1 行の JSON だけになり、人間向けの出力と子プロセスの log は
stderr に回る (`bin/lib/json.sh`)。不可逆操作 (commit / push / PR / merge) は
引き続き `bin/step-*` だけが行う。
