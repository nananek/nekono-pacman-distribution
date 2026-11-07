package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RepoState is the git view of the working tree.
type RepoState struct {
	Branch   string   `json:"branch"`
	Head     string   `json:"head"`
	Dirty    []string `json:"dirty"`
	Upstream string   `json:"upstream,omitempty"`
	Ahead    int      `json:"ahead"`
	Behind   int      `json:"behind"`
}

// RepoStateTool reports branch / dirty paths / ahead-behind.
func RepoStateTool(env *Env) *Func {
	return &Func{
		Name:   "repo_state",
		Desc:   "Read git state: current branch, HEAD sha, dirty paths, upstream and ahead/behind counts.",
		Params: obj(``),
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			st := RepoState{}
			res := run(ctx, env.RepoRoot, execTimeout(env), "git", "status", "--porcelain=v1", "--branch")
			if res.RC != 0 {
				return Result{}, fmt.Errorf("git status failed: %s", errText(res))
			}
			for _, line := range strings.Split(res.Stdout, "\n") {
				if line == "" {
					continue
				}
				if strings.HasPrefix(line, "## ") {
					rest := strings.TrimPrefix(line, "## ")
					if i := strings.Index(rest, "..."); i >= 0 {
						st.Branch = rest[:i]
						up := rest[i+3:]
						if j := strings.Index(up, " ["); j >= 0 {
							st.Upstream = up[:j]
							ab := strings.TrimSuffix(up[j+2:], "]")
							for _, part := range strings.Split(ab, ", ") {
								if strings.HasPrefix(part, "ahead ") {
									st.Ahead, _ = strconv.Atoi(strings.TrimPrefix(part, "ahead "))
								}
								if strings.HasPrefix(part, "behind ") {
									st.Behind, _ = strconv.Atoi(strings.TrimPrefix(part, "behind "))
								}
							}
						} else {
							st.Upstream = up
						}
					} else {
						st.Branch = rest
					}
					continue
				}
				st.Dirty = append(st.Dirty, line)
			}
			if head := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "HEAD"); head.RC == 0 {
				st.Head = strings.TrimSpace(head.Stdout)
			}
			env.J.Emit("repo_state", map[string]any{"branch": st.Branch, "dirty": len(st.Dirty), "ahead": st.Ahead, "behind": st.Behind})
			return Result{Summary: fmt.Sprintf("branch=%s dirty=%d ahead=%d behind=%d", st.Branch, len(st.Dirty), st.Ahead, st.Behind), Data: st}, nil
		},
	}
}

// RepoDiffTool shows a diff, optionally limited to paths and a ref.
func RepoDiffTool(env *Env) *Func {
	return &Func{
		Name:   "repo_diff",
		Desc:   "Read a git diff (working tree vs ref, default HEAD). Optional paths narrow it. Output is truncated for very large diffs.",
		Params: obj(`"ref":{"type":"string","description":"commit/ref to diff against (default HEAD)"},"paths":{"type":"array","items":{"type":"string"},"description":"paths to limit the diff to"}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Ref   string   `json:"ref"`
				Paths []string `json:"paths"`
			}
			_ = json.Unmarshal(raw, &a)
			ref := a.Ref
			if ref == "" {
				ref = "HEAD"
			}
			args := []string{"diff", "--no-color", "--no-ext-diff", ref, "--"}
			args = append(args, a.Paths...)
			res := run(ctx, env.RepoRoot, execTimeout(env), "git", args...)
			if res.RC != 0 && strings.TrimSpace(res.Stdout) == "" {
				return Result{}, fmt.Errorf("git diff failed: %s", errText(res))
			}
			return Result{Summary: fmt.Sprintf("diff vs %s: %d bytes", ref, len(res.Stdout)), Data: map[string]any{"ref": ref, "diff": truncate(res.Stdout, 20000)}}, nil
		},
	}
}

// RepoLogTool lists recent commits.
func RepoLogTool(env *Env) *Func {
	return &Func{
		Name:   "repo_log",
		Desc:   "Read recent commits (oneline, with signature status).",
		Params: obj(`"count":{"type":"integer","description":"how many commits (default 10)"},"paths":{"type":"array","items":{"type":"string"}}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Count int      `json:"count"`
				Paths []string `json:"paths"`
			}
			_ = json.Unmarshal(raw, &a)
			if a.Count <= 0 || a.Count > 100 {
				a.Count = 10
			}
			args := []string{"log", "--pretty=%h %G? %an | %s", "-n", strconv.Itoa(a.Count), "--"}
			args = append(args, a.Paths...)
			res := run(ctx, env.RepoRoot, execTimeout(env), "git", args...)
			if res.RC != 0 {
				return Result{}, fmt.Errorf("git log failed: %s", errText(res))
			}
			return Result{Summary: fmt.Sprintf("%d commit(s)", a.Count), Data: map[string]any{"log": res.Stdout}}, nil
		},
	}
}

// RepoMergeBaseTool resolves the merge base with a ref.
func RepoMergeBaseTool(env *Env) *Func {
	return &Func{
		Name:   "repo_merge_base",
		Desc:   "Resolve the merge base between a ref (default origin/master) and HEAD.",
		Params: obj(`"ref":{"type":"string"}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Ref string `json:"ref"`
			}
			_ = json.Unmarshal(raw, &a)
			if a.Ref == "" {
				a.Ref = "origin/master"
			}
			res := run(ctx, env.RepoRoot, execTimeout(env), "git", "merge-base", a.Ref, "HEAD")
			if res.RC != 0 {
				return Result{}, fmt.Errorf("git merge-base failed: %s", errText(res))
			}
			base := strings.TrimSpace(res.Stdout)
			return Result{Summary: "merge-base " + base, Data: map[string]any{"base": base}}, nil
		},
	}
}

// RepoPullTool fast-forwards the current branch (mutating).
func RepoPullTool(env *Env) *Func {
	return &Func{
		Name:    "repo_pull",
		Desc:    "Run `git pull --ff-only` on the current branch. Mutates the working tree.",
		Params:  obj(``),
		Mutates: true,
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			res := run(ctx, env.RepoRoot, execTimeout(env), "git", "pull", "--ff-only")
			if res.RC != 0 {
				return Result{}, fmt.Errorf("git pull --ff-only failed: %s", errText(res))
			}
			return Result{Summary: "fast-forwarded", Data: map[string]any{"output": tail(res.Stdout, 5)}}, nil
		},
	}
}
