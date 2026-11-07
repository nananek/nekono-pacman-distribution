package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SigningProbeTool reports whether the signing path is reachable. It reuses
// bin/preflight so the VM/host difference (gpg shim -> ssh -> host rga) is
// decided in exactly one place.
func SigningProbeTool(env *Env) *Func {
	return &Func{
		Name:   "signing_probe",
		Desc:   "Check signing readiness (rga reachable, directly or through the VM->host bridge).",
		Params: obj(``),
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			res := run(ctx, env.RepoRoot, execTimeout(env), "bin/preflight", "--json")
			var data struct {
				SigningOK     bool   `json:"signing_ok"`
				SigningDetail string `json:"signing_detail"`
			}
			if err := json.Unmarshal([]byte(res.Stdout), &data); err != nil {
				return Result{}, &HarnessGap{Tool: "signing_probe", Detail: "bin/preflight --json returned non-JSON"}
			}
			env.J.Emit("signing_probe", map[string]any{"ok": data.SigningOK, "detail": data.SigningDetail})
			return Result{Summary: fmt.Sprintf("signing ok=%v (%s)", data.SigningOK, data.SigningDetail), Data: data}, nil
		},
	}
}

// SigningVerifyTool verifies detached signatures.
func SigningVerifyTool(env *Env) *Func {
	return &Func{
		Name:   "signing_verify",
		Desc:   "Verify detached GPG signatures (pass the .sig paths; the target is derived by stripping .sig).",
		Params: obj(`"sig_paths":{"type":"array","items":{"type":"string"},"description":"repo-relative or absolute .sig paths"}`, "sig_paths"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				SigPaths []string `json:"sig_paths"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			type check struct {
				Sig    string `json:"sig"`
				Target string `json:"target"`
				Result string `json:"result"`
				Detail string `json:"detail,omitempty"`
			}
			checks := make([]check, 0, len(a.SigPaths))
			for _, p := range a.SigPaths {
				abs, err := resolveInRepo(env, p)
				if err != nil {
					return Result{}, err
				}
				target := strings.TrimSuffix(abs, ".sig")
				res := run(ctx, env.RepoRoot, execTimeout(env), "gpg", "--verify", abs, target)
				out := res.Stdout + res.Stderr
				c := check{Sig: p, Target: strings.TrimSuffix(p, ".sig")}
				switch {
				case strings.Contains(out, "Good signature"):
					c.Result = "GOOD"
				case strings.Contains(out, "BAD signature"):
					c.Result = "BAD"
				default:
					c.Result = "ERROR"
					c.Detail = truncate(out, 400)
				}
				checks = append(checks, c)
				env.J.Emit("signing_verify", map[string]any{"sig": p, "result": c.Result})
			}
			allGood := true
			for _, c := range checks {
				if c.Result != "GOOD" {
					allGood = false
				}
			}
			return Result{Summary: fmt.Sprintf("verified %d signature(s), all_good=%v", len(checks), allGood), Data: map[string]any{"checks": checks, "all_good": allGood}}, nil
		},
	}
}
