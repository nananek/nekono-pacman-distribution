package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/nananek/nekono-pacman-distribution/internal/phase"
	"github.com/nananek/nekono-pacman-distribution/internal/tool"
)

// runState is what the report is built from. Everything here is a structured
// fact gathered by Go, not prose produced by the model.
type runState struct {
	Preflight   map[string]any `json:"preflight"`
	Detect      map[string]any `json:"detect,omitempty"`
	Deps        map[string]any `json:"deps,omitempty"`
	BuildPlan   tool.BuildPlan `json:"build_plan"`
	BuildPhase  *phaseOutcome  `json:"build_phase,omitempty"`
	BuildVerify map[string]any `json:"build_verify,omitempty"`
	PubPhase    *phaseOutcome  `json:"publish_phase,omitempty"`
	DistVerify  map[string]any `json:"dist_verify,omitempty"`
	Status      string         `json:"status"`
	Notes       []string       `json:"notes,omitempty"`
}

type phaseOutcome struct {
	Phase   string `json:"phase"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Steps   int    `json:"steps"`
	Tokens  int    `json:"tokens"`
}

func cmdRun(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	cfgPath := fs.String("config", "agent/agent.json", "path to agent config")
	repo := fs.String("repo", ".", "repository root")
	skipDetect := fs.Bool("skip-detect", false, "skip bin/check-upstream and bin/check-deps")
	skipBuild := fs.Bool("skip-build", false, "do not run the build phase")
	skipPublish := fs.Bool("skip-publish", false, "do not run the publish phase")
	dryRun := fs.Bool("dry-run", false, "plan only: no build, no publish")
	if err := fs.Parse(args); err != nil {
		return err
	}
	a, _, err := setup(ctx, []string{"-config", *cfgPath, "-repo", *repo})
	if err != nil {
		return err
	}

	st := &runState{}

	// Phase 0 (deterministic): preflight state. No model call: every decision
	// the driver takes from it is a rule, not a judgement.
	pf, err := tool.PreflightTool(a.env).Run(ctx, json.RawMessage(`{}`))
	if err != nil {
		return err
	}
	st.Preflight, _ = pf.Data.(map[string]any)
	buildHost, _ := st.Preflight["build_host"].(bool)
	publishState, _ := st.Preflight["publish"].(string)
	a.j.Human("\n== preflight: branch=%v build_host=%v publish=%v", st.Preflight["branch"], buildHost, publishState)

	// Phase 0.5: detection (deterministic, idempotent entrypoints).
	if !*skipDetect {
		if data, res, derr := tool.ExecJSON(ctx, a.root, time.Hour, "bin/check-upstream", "--json"); derr != nil {
			st.Notes = append(st.Notes, fmt.Sprintf("check-upstream failed (rc=%d)", res.RC))
			a.j.Emit("warn", map[string]any{"stage": "detect", "msg": "check-upstream failed", "rc": res.RC})
		} else {
			st.Detect = data
			a.j.Emit("detect", map[string]any{"stage": "upstream", "rc": res.RC, "created": lenList(data["created"]), "skipped": lenList(data["skipped"])})
		}
		if data, res, derr := tool.ExecJSON(ctx, a.root, 2*time.Hour, "bin/check-deps", "--json"); derr != nil {
			st.Notes = append(st.Notes, fmt.Sprintf("check-deps failed (rc=%d)", res.RC))
			a.j.Emit("warn", map[string]any{"stage": "detect", "msg": "check-deps failed", "rc": res.RC})
		} else {
			st.Deps = data
			a.j.Emit("detect", map[string]any{"stage": "deps", "rc": res.RC, "entries": lenList(data["entries"])})
		}
	}

	// Phase 3: build plan (deterministic), then the build phase if there is
	// anything to build and we are on the build host.
	planRes, perr := tool.BuildPlanTool(a.env).Run(ctx, json.RawMessage(`{}`))
	if perr != nil {
		return perr
	}
	if p, ok := planRes.Data.(tool.BuildPlan); ok {
		st.BuildPlan = p
	}

	pending := len(st.BuildPlan.Pending)
	built := 0
	if pending > 0 && buildHost && !*skipBuild && !*dryRun {
		out, rerr := a.runPhase(ctx, "build")
		st.BuildPhase = toOutcome(out)
		if rerr != nil || out.Status != "ok" {
			st.Status = "failed_at_build"
			st.Notes = append(st.Notes, "build phase did not finish ok: "+out.Summary)
			return finishRun(a, st)
		}
		// Deterministic verification: the model's claim is not the criterion.
		args, _ := json.Marshal(map[string]any{"pkgs": st.BuildPlan.Pending})
		vr, verr := tool.BuildVerifyTool(a.env).Run(ctx, args)
		if verr != nil {
			return verr
		}
		st.BuildVerify, _ = vr.Data.(map[string]any)
		if ok, _ := st.BuildVerify["ok"].(bool); !ok {
			st.Status = "build_verify_failed"
			st.Notes = append(st.Notes, "artifact verification failed (see build_verify.checks)")
			return finishRun(a, st)
		}
		built = pending
	} else if pending > 0 && !buildHost {
		st.Notes = append(st.Notes, "not a build host: build/publish must run on the build host")
	}

	// Phase 4: publish. Run it when we built something, or when preflight
	// found local db/sig differing from the published copy (retire / repo-add).
	needPublish := (built > 0 || publishState == "pending") && !*skipPublish && !*dryRun
	if needPublish {
		out, rerr := a.runPhase(ctx, "publish")
		st.PubPhase = toOutcome(out)
		if rerr != nil || out.Status != "ok" {
			st.Status = "failed_at_publish"
			st.Notes = append(st.Notes, "publish phase did not finish ok: "+out.Summary)
			return finishRun(a, st)
		}
		files := append([]string{"nekono.db.tar.gz", "nekono.db.tar.gz.sig"}, builtFiles(a.root, st.BuildPlan)...)
		args, _ := json.Marshal(map[string]any{"files": files})
		vr, verr := tool.DistVerifyTool(a.env).Run(ctx, args)
		if verr != nil {
			return verr
		}
		st.DistVerify, _ = vr.Data.(map[string]any)
		if ok, _ := st.DistVerify["ok"].(bool); !ok {
			st.Status = "dist_verify_failed"
			st.Notes = append(st.Notes, "published files do not match the local repo (see dist_verify.checks)")
			return finishRun(a, st)
		}
	}

	if st.Status == "" {
		st.Status = "ok"
	}
	return finishRun(a, st)
}

func toOutcome(o phase.Outcome) *phaseOutcome {
	return &phaseOutcome{Phase: o.Phase, Status: o.Status, Summary: o.Summary, Steps: o.Steps, Tokens: o.Usage.TotalTokens}
}

func lenList(v any) int {
	if a, ok := v.([]any); ok {
		return len(a)
	}
	return 0
}

// builtFiles lists the package files that exist for the planned packages.
func builtFiles(root string, plan tool.BuildPlan) []string {
	dir := filepath.Join(root, "repo", "x86_64")
	var out []string
	for _, p := range plan.Pending {
		matches, _ := filepath.Glob(filepath.Join(dir, p.Pkg+"-"+p.Version+"-*.pkg.tar.zst"))
		for _, m := range matches {
			base := filepath.Base(m)
			out = append(out, base, base+".sig")
		}
	}
	sort.Strings(out)
	return out
}

// finishRun prints the deterministic report (Phase 5) and journals it.
func finishRun(a *app, st *runState) error {
	a.j.Emit("report", map[string]any{
		"status": st.Status, "notes": st.Notes,
		"pending": len(st.BuildPlan.Pending), "cascade": len(st.BuildPlan.Cascade),
		"retire": st.BuildPlan.Retire,
	})
	b, _ := json.Marshal(st)
	fmt.Fprintln(os.Stdout, string(b))

	fmt.Fprintf(os.Stderr, "\n== run %s\n", st.Status)
	fmt.Fprintf(os.Stderr, "   build: pending=%d cascade=%d retire=%v\n", len(st.BuildPlan.Pending), len(st.BuildPlan.Cascade), st.BuildPlan.Retire)
	if st.BuildVerify != nil {
		fmt.Fprintf(os.Stderr, "   build_verify: ok=%v\n", st.BuildVerify["ok"])
	}
	if st.DistVerify != nil {
		fmt.Fprintf(os.Stderr, "   dist_verify: ok=%v\n", st.DistVerify["ok"])
	}
	for _, n := range st.Notes {
		fmt.Fprintf(os.Stderr, "   note: %s\n", n)
	}
	if st.Status != "ok" {
		return fmt.Errorf("run finished with status %s", st.Status)
	}
	return nil
}
