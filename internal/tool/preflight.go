package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// execTimeout is the default budget for one short command.
func execTimeout(env *Env) time.Duration {
	s := env.Cfg.ExecTimeoutSec
	if s <= 0 {
		s = 300
	}
	return time.Duration(s) * time.Second
}

// buildTimeout is the budget for build/publish (they run for tens of minutes).
func buildTimeout(env *Env) time.Duration {
	s := env.Cfg.BuildTimeoutSec
	if s <= 0 {
		s = 6 * 60 * 60
	}
	return time.Duration(s) * time.Second
}

// PreflightTool runs bin/preflight --json and hands back its structured state.
// It is read-only except for `git fetch origin` inside the script.
func PreflightTool(env *Env) *Func {
	return &Func{
		Name:   "preflight",
		Desc:   "Read the daily-ops preflight state (branch, dirty files, hook config, gh auth, open bot PRs / upstream Issues, signing readiness, build-host flag, publish drift). Read-only.",
		Params: obj(``),
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			res := run(ctx, env.RepoRoot, execTimeout(env), "bin/preflight", "--json")
			var data map[string]any
			if err := json.Unmarshal([]byte(res.Stdout), &data); err != nil {
				return Result{}, &HarnessGap{
					Tool:   "preflight",
					Detail: fmt.Sprintf("bin/preflight --json returned non-JSON (rc=%d): %s", res.RC, errText(res)),
				}
			}
			env.J.Emit("preflight", map[string]any{
				"branch":      data["branch"],
				"build_host":  data["build_host"],
				"signing_ok":  data["signing_ok"],
				"publish":     data["publish"],
				"bot_prs":     countOf(data["bot_prs"]),
				"upstream":    countOf(data["upstream_issues"]),
				"hooks_ok":    data["hooks_ok"],
				"dirty_files": countOf(data["dirty"]),
			})
			return Result{Summary: fmt.Sprintf("preflight ok (branch=%v build_host=%v publish=%v)", data["branch"], data["build_host"], data["publish"]), Data: data}, nil
		},
	}
}

func countOf(v any) int {
	if a, ok := v.([]any); ok {
		return len(a)
	}
	return 0
}
