package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiskFreeTool reports free space on the repo filesystem (large builds need it).
func DiskFreeTool(env *Env) *Func {
	return &Func{
		Name:   "disk_free",
		Desc:   "Report free/total bytes of the repository filesystem.",
		Params: obj(``),
		Run: func(_ context.Context, _ json.RawMessage) (Result, error) {
			free, total, err := statfs(env.RepoRoot)
			if err != nil {
				return Result{}, err
			}
			gib := func(b uint64) float64 { return float64(b) / (1 << 30) }
			return Result{
				Summary: fmt.Sprintf("free %.1f GiB of %.1f GiB", gib(free), gib(total)),
				Data:    map[string]any{"free_bytes": free, "total_bytes": total},
			}, nil
		},
	}
}

// BuildPlan is the parsed `bin/build-all --pending --dry-run --json` document.
type BuildPlan struct {
	DryRun  bool `json:"dry_run"`
	Pending []struct {
		Pkg     string `json:"pkg"`
		Version string `json:"version"`
	} `json:"pending"`
	Skipped []struct {
		Pkg     string `json:"pkg"`
		Version string `json:"version"`
	} `json:"skipped"`
	Cascade []struct {
		Pkg     string `json:"pkg"`
		Version string `json:"version"`
	} `json:"cascade"`
	Retire []string `json:"retire"`
	Built  []string `json:"built"`
	Error  string   `json:"error"`
	RC     int      `json:"rc"`
}

// BuildPlanTool asks build-all what it would do.
func BuildPlanTool(env *Env) *Func {
	return &Func{
		Name:   "build_plan",
		Desc:   "Dry-run `bin/build-all --pending` and return the structured plan: packages to build (with versions), already-built skips, cascade-warn packages, packages to retire.",
		Params: obj(``),
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			res := run(ctx, env.RepoRoot, buildTimeout(env), "bin/build-all", "--pending", "--dry-run", "--json")
			plan, err := parsePlan(res)
			if err != nil {
				return Result{}, err
			}
			env.J.Emit("build_plan", map[string]any{
				"pending": len(plan.Pending), "skipped": len(plan.Skipped),
				"cascade": len(plan.Cascade), "retire": len(plan.Retire),
			})
			summary := fmt.Sprintf("pending=%d skipped=%d cascade=%d retire=%d", len(plan.Pending), len(plan.Skipped), len(plan.Cascade), len(plan.Retire))
			return Result{Summary: summary, Data: plan}, nil
		},
	}
}

// BuildRunTool runs the real build and returns the same structured document.
func BuildRunTool(env *Env) *Func {
	return &Func{
		Name:    "build_run",
		Desc:    "Run `bin/build-all --pending` (makepkg + GPG sign + repo-add + prune). Long-running: progress lines are journaled while it runs. Returns the structured result (built/failed/error/rc).",
		Params:  obj(`"dry_run":{"type":"boolean","description":"optional: only plan (same as build_plan)"}`),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				DryRun bool `json:"dry_run"`
			}
			_ = json.Unmarshal(raw, &a)
			args := []string{"--pending", "--json"}
			if a.DryRun {
				args = append(args, "--dry-run")
			}
			res := runStreaming(ctx, env, "build", env.RepoRoot, buildTimeout(env), "bin/build-all", args...)
			plan, err := parsePlan(res)
			if err != nil {
				return Result{}, err
			}
			env.J.Emit("build_run", map[string]any{
				"rc": plan.RC, "built": plan.Built, "error": plan.Error, "dry_run": plan.DryRun,
			})
			summary := fmt.Sprintf("rc=%d built=%v", plan.RC, plan.Built)
			if plan.Error != "" {
				summary += " error=" + plan.Error
			}
			return Result{Summary: summary, Data: plan}, nil
		},
	}
}

func parsePlan(res CmdResult) (BuildPlan, error) {
	var plan BuildPlan
	line := strings.TrimSpace(res.Stdout)
	if line == "" {
		// A crashed script may still have written human output to stderr.
		return plan, &HarnessGap{Tool: "build_plan", Detail: fmt.Sprintf("build-all --json produced no JSON (rc=%d): %s", res.RC, errText(res))}
	}
	// The document is one line; take the last non-empty line in case a child
	// process printed something on stdout before the trap fired.
	if i := strings.LastIndexByte(line, '\n'); i >= 0 {
		line = line[i+1:]
	}
	if err := json.Unmarshal([]byte(line), &plan); err != nil {
		return plan, &HarnessGap{Tool: "build_plan", Detail: "build-all --json output is not valid JSON: " + err.Error()}
	}
	return plan, nil
}

// BuildVerifyTool is the deterministic post-build check. The model is not asked
// to trust `rc=0`: presence, signature, db entry and "nothing left to build"
// are all verified here.
func BuildVerifyTool(env *Env) *Func {
	return &Func{
		Name:   "build_verify",
		Desc:   "Verify built artifacts for the given packages: .pkg.tar.zst + .sig exist and are signed by a good key, the repo db lists the new version, and `build-all --pending --dry-run` reports nothing left. Returns named checks.",
		Params: obj(`"pkgs":{"type":"array","items":{"type":"object","properties":{"pkg":{"type":"string"},"version":{"type":"string"}},"required":["pkg","version"]},"description":"packages that were just built"}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Pkgs []struct {
					Pkg     string `json:"pkg"`
					Version string `json:"version"`
				} `json:"pkgs"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			type check struct {
				Name   string `json:"name"`
				Pkg    string `json:"pkg,omitempty"`
				OK     bool   `json:"ok"`
				Detail string `json:"detail,omitempty"`
			}
			var checks []check
			repodir := filepath.Join(env.RepoRoot, "repo", "x86_64")
			db := filepath.Join(repodir, "nekono.db.tar.gz")

			dbFiles := map[string]bool{}
			if res := run(ctx, env.RepoRoot, execTimeout(env), "tar", "-xzO", "-f", db, "--wildcards", "*/desc"); res.RC == 0 {
				for _, line := range strings.Split(res.Stdout, "\n") {
					if strings.HasPrefix(line, "%FILENAME%") {
						continue
					}
					if strings.Contains(line, ".pkg.tar.zst") {
						dbFiles[line] = true
					}
				}
			} else {
				checks = append(checks, check{Name: "db_readable", OK: false, Detail: errText(res)})
			}

			for _, p := range a.Pkgs {
				pattern := filepath.Join(repodir, p.Pkg+"-"+p.Version+"-*.pkg.tar.zst")
				matches, _ := filepath.Glob(pattern)
				if len(matches) == 0 {
					checks = append(checks, check{Name: "artifact_exists", Pkg: p.Pkg, OK: false, Detail: "no match for " + filepath.Base(pattern)})
					continue
				}
				art := matches[0]
				_, sigErr := os.Stat(art + ".sig")
				checks = append(checks, check{Name: "artifact_exists", Pkg: p.Pkg, OK: true, Detail: filepath.Base(art)})
				checks = append(checks, check{Name: "sig_file", Pkg: p.Pkg, OK: sigErr == nil, Detail: filepath.Base(art) + ".sig"})

				vr := run(ctx, env.RepoRoot, execTimeout(env), "gpg", "--verify", art+".sig", art)
				good := strings.Contains(vr.Stdout+vr.Stderr, "Good signature")
				detail := ""
				if !good {
					detail = truncate(vr.Stdout+vr.Stderr, 300)
				}
				checks = append(checks, check{Name: "artifact_signed", Pkg: p.Pkg, OK: good, Detail: detail})
				checks = append(checks, check{Name: "db_lists_artifact", Pkg: p.Pkg, OK: dbFiles[filepath.Base(art)], Detail: filepath.Base(art)})
			}

			vr := run(ctx, env.RepoRoot, execTimeout(env), "gpg", "--verify", db+".sig", db)
			checks = append(checks, check{Name: "db_signed", OK: strings.Contains(vr.Stdout+vr.Stderr, "Good signature")})

			if res := run(ctx, env.RepoRoot, buildTimeout(env), "bin/build-all", "--pending", "--dry-run", "--json"); res.RC == 0 {
				if plan, err := parsePlan(res); err == nil {
					checks = append(checks, check{Name: "nothing_left_to_build", OK: len(plan.Pending) == 0, Detail: fmt.Sprintf("pending=%d", len(plan.Pending))})
				}
			}

			ok := true
			for _, c := range checks {
				if !c.OK {
					ok = false
				}
				env.J.Emit("verify", map[string]any{"check": c.Name, "pkg": c.Pkg, "ok": c.OK})
			}
			return Result{Summary: fmt.Sprintf("build verification ok=%v (%d checks)", ok, len(checks)), Data: map[string]any{"ok": ok, "checks": checks}}, nil
		},
	}
}
