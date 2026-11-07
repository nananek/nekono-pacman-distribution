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
| build | build_plan, build_run, signing_probe, signing_verify, disk_free | 失敗の切り分け。成否は `build_verify` が決める |
| publish | dist_publish, dist_verify, preflight | 失敗時の報告。一致は `dist_verify` が決める |

`run` は `preflight`(tool) → `check-upstream`/`check-deps` → `build_plan` → build phase →
`build_verify` → publish phase → `dist_verify` の順に回す。build するものが無く
`publish: PENDING` のときは build を飛ばして publish する。

**未実装 (次のコミット)**: triage (bot PR) / upstream (Issue 起点 bump) の phase。
現在 `run` はこの 2 つを skip し、レポートの notes に残す。

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
