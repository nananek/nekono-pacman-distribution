package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// The step_* tools are the only irreversible paths a phase can take. They wrap
// the single entrypoints in bin/ (signing, pushing, merging live there), so the
// model never assembles a git/gh command itself.

// StepBotPRTool re-signs and force-pushes a bot pkgrel PR branch.
func StepBotPRTool(env *Env) *Func {
	return &Func{
		Name:    "step_bot_pr",
		Desc:    "Run `bin/step-bot-pr <N>`: stage pkgs/<pkg>/{PKGBUILD,REVIEW.md,.SRCINFO,.deps.lock}, amend the bot commit with a Nekono GPG signature and force-push with lease (the pre-push gate runs). Requires the PR branch to be checked out (gh_pr checkout).",
		Params:  obj(`"number":{"type":"integer"}`, "number"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Number int `json:"number"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			res := run(ctx, env.RepoRoot, buildTimeout(env), "bin/step-bot-pr", strconv.Itoa(a.Number))
			out := res.Stdout + res.Stderr
			env.J.Emit("step", map[string]any{"step": "bot_pr", "number": a.Number, "rc": res.RC})
			return Result{
				Summary: fmt.Sprintf("step-bot-pr %d: rc=%d", a.Number, res.RC),
				Data:    map[string]any{"rc": res.RC, "output": truncate(out, 4000)},
			}, nil
		},
	}
}

// StepMergeTool merges an approved PR through bin/step-merge.
func StepMergeTool(env *Env) *Func {
	return &Func{
		Name:    "step_merge",
		Desc:    "Run `bin/step-merge <N>`: merge a PR whose head is deps/<pkg>-pkgrel-<N> or pkg/<pkg>-<ver>, then return to master. Judgement review must have happened already.",
		Params:  obj(`"number":{"type":"integer"}`, "number"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Number int `json:"number"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			res := run(ctx, env.RepoRoot, buildTimeout(env), "bin/step-merge", strconv.Itoa(a.Number))
			out := res.Stdout + res.Stderr
			env.J.Emit("step", map[string]any{"step": "merge", "number": a.Number, "rc": res.RC})
			return Result{
				Summary: fmt.Sprintf("step-merge %d: rc=%d", a.Number, res.RC),
				Data:    map[string]any{"rc": res.RC, "output": truncate(out, 4000)},
			}, nil
		},
	}
}

// StepBumpPRTool creates the signed bump branch + PR for an upstream update.
func StepBumpPRTool(env *Env) *Func {
	return &Func{
		Name: "step_bump_pr",
		Desc: "Run `bin/step-upstream-pr <pkg> <ver>`: create pkg/<pkg>-<ver> from HEAD carrying the working-tree edits, stage only pkgs/<pkg>/{PKGBUILD,REVIEW.md,.SRCINFO,.deps.lock}, commit with a Nekono GPG signature, push and open the PR.",
		Params: obj(`"pkg":{"type":"string"},"version":{"type":"string"},"title":{"type":"string"},"message":{"type":"string","description":"commit message (attribution line included)"},"body":{"type":"string","description":"PR body; put 'Closes #N' per issue, one keyword each"},"closes":{"type":"array","items":{"type":"integer"},"description":"issue numbers to close (only used when body is empty)"}`,
			"pkg", "version", "title", "message"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Pkg     string `json:"pkg"`
				Version string `json:"version"`
				Title   string `json:"title"`
				Message string `json:"message"`
				Body    string `json:"body"`
				Closes  []int  `json:"closes"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			if !validPkgName(a.Pkg) {
				return Result{}, fmt.Errorf("invalid package name: %q", a.Pkg)
			}
			msgFile := filepath.Join(env.Scratch, "bump-msg.txt")
			if err := os.WriteFile(msgFile, []byte(a.Message), 0o644); err != nil {
				return Result{}, err
			}
			args := []string{a.Pkg, a.Version, "--title", a.Title, "--message-file", msgFile}
			if a.Body != "" {
				bodyFile := filepath.Join(env.Scratch, "bump-body.txt")
				if err := os.WriteFile(bodyFile, []byte(a.Body), 0o644); err != nil {
					return Result{}, err
				}
				args = append(args, "--body-file", bodyFile)
			} else {
				for _, n := range a.Closes {
					args = append(args, "--closes", strconv.Itoa(n))
				}
			}
			res := run(ctx, env.RepoRoot, buildTimeout(env), "bin/step-upstream-pr", args...)
			out := res.Stdout + res.Stderr
			env.J.Emit("step", map[string]any{"step": "bump_pr", "pkg": a.Pkg, "version": a.Version, "rc": res.RC})
			return Result{
				Summary: fmt.Sprintf("step-upstream-pr %s %s: rc=%d", a.Pkg, a.Version, res.RC),
				Data:    map[string]any{"rc": res.RC, "output": truncate(out, 4000)},
			}, nil
		},
	}
}

// mergeAllowedBranch reports whether a PR head is one of the two pre-approved
// autonomous paths (used by the driver to decide what triage may touch).
func mergeAllowedBranch(head string) bool {
	if strings.HasPrefix(head, "deps/") && strings.Contains(head, "-pkgrel-") {
		return true
	}
	return strings.HasPrefix(head, "pkg/")
}
