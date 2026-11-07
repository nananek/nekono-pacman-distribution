package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// gh_pr wraps the pull-request operations the phases need. One tool with a
// method keeps the schema small; every method maps to a fixed gh invocation
// whose argv the harness builds.
func GHPRTool(env *Env) *Func {
	return &Func{
		Name: "gh_pr",
		Desc: "Pull requests: method=list (open PRs, filter=deps to keep only deps/*), view, diff, checkout (switch the working tree to the PR branch), close.",
		Params: obj(`"method":{"type":"string","enum":["list","view","diff","checkout","close"]},"number":{"type":"integer"},"state":{"type":"string","enum":["open","closed","all"],"description":"list: default open"},"filter":{"type":"string","enum":["all","deps"],"description":"list: deps keeps headRefName starting with deps/"},"comment":{"type":"string","description":"close: optional comment"}`,
			"method"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Method  string `json:"method"`
				Number  int    `json:"number"`
				State   string `json:"state"`
				Filter  string `json:"filter"`
				Comment string `json:"comment"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			switch a.Method {
			case "list":
				state := a.State
				if state == "" {
					state = "open"
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "pr", "list",
					"--state", state, "--limit", "100",
					"--json", "number,headRefName,title,author,isDraft")
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh pr list failed: %s", errText(res))
				}
				var prs []struct {
					Number      int    `json:"number"`
					HeadRefName string `json:"headRefName"`
					Title       string `json:"title"`
					IsDraft     bool   `json:"isDraft"`
				}
				if err := json.Unmarshal([]byte(res.Stdout), &prs); err != nil {
					return Result{}, fmt.Errorf("gh pr list: %w", err)
				}
				out := make([]map[string]any, 0, len(prs))
				for _, p := range prs {
					if a.Filter == "deps" && !strings.HasPrefix(p.HeadRefName, "deps/") {
						continue
					}
					out = append(out, map[string]any{"number": p.Number, "head": p.HeadRefName, "title": p.Title, "draft": p.IsDraft})
				}
				env.J.Emit("gh_pr", map[string]any{"method": "list", "filter": a.Filter, "count": len(out)})
				return Result{Summary: fmt.Sprintf("%d PR(s)", len(out)), Data: map[string]any{"prs": out}}, nil

			case "view":
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "pr", "view", strconv.Itoa(a.Number),
					"--json", "number,title,state,baseRefName,headRefName,mergeable,body,files")
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh pr view failed: %s", errText(res))
				}
				var v map[string]any
				if err := json.Unmarshal([]byte(res.Stdout), &v); err != nil {
					return Result{}, fmt.Errorf("gh pr view: %w", err)
				}
				return Result{Summary: fmt.Sprintf("PR #%d %v", a.Number, v["state"]), Data: v}, nil

			case "diff":
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "pr", "diff", strconv.Itoa(a.Number))
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh pr diff failed: %s", errText(res))
				}
				return Result{Summary: fmt.Sprintf("diff %d bytes", len(res.Stdout)), Data: map[string]any{"diff": truncate(res.Stdout, 20000)}}, nil

			case "checkout":
				// gh pr checkout needs refs/remotes/origin/<head> to exist. A
				// clone with a narrowed remote.origin.fetch (this host's older
				// checkouts have master-only) makes it fail half-way and leaves
				// the tree switched, so widen the refspec first (harmless).
				// Refuse to switch branches under uncommitted work: a partial
				// checkout would leave the tree in a state no other tool can
				// repair (only fs_write for managed files exists).
				if st := run(ctx, env.RepoRoot, execTimeout(env), "git", "status", "--porcelain"); st.RC != 0 || strings.TrimSpace(st.Stdout) != "" {
					return Result{}, fmt.Errorf("working tree is not clean; refusing to switch branches (commit or stash first)")
				}
				run(ctx, env.RepoRoot, 2*time.Minute, "git", "fetch", "origin",
					"+refs/heads/*:refs/remotes/origin/*")
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "pr", "checkout", strconv.Itoa(a.Number))
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh pr checkout failed: %s", errText(res))
				}
				branch := ""
				if b := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "--abbrev-ref", "HEAD"); b.RC == 0 {
					branch = strings.TrimSpace(b.Stdout)
				}
				env.J.Emit("gh_pr", map[string]any{"method": "checkout", "number": a.Number, "branch": branch})
				return Result{Summary: fmt.Sprintf("checked out PR #%d (%s)", a.Number, branch), Data: map[string]any{"ok": true, "branch": branch}}, nil

			case "close":
				args := []string{"pr", "close", strconv.Itoa(a.Number)}
				if a.Comment != "" {
					args = append(args, "--comment", a.Comment)
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", args...)
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh pr close failed: %s", errText(res))
				}
				env.J.Emit("gh_pr", map[string]any{"method": "close", "number": a.Number})
				return Result{Summary: fmt.Sprintf("closed PR #%d", a.Number), Data: map[string]any{"closed": a.Number}}, nil
			}
			return Result{}, &HarnessGap{Tool: "gh_pr", Detail: "unknown method " + a.Method}
		},
	}
}

// GHIssueTool wraps issue operations.
func GHIssueTool(env *Env) *Func {
	return &Func{
		Name: "gh_issue",
		Desc: "Issues: method=list (open issues, filter=upstream keeps '[pkg] upstream version:' ones), view, comment, close.",
		Params: obj(`"method":{"type":"string","enum":["list","view","comment","close"]},"number":{"type":"integer"},"state":{"type":"string","enum":["open","closed","all"]},"filter":{"type":"string","enum":["all","upstream"]},"body":{"type":"string"}`,
			"method"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Method string `json:"method"`
				Number int    `json:"number"`
				State  string `json:"state"`
				Filter string `json:"filter"`
				Body   string `json:"body"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			switch a.Method {
			case "list":
				state := a.State
				if state == "" {
					state = "open"
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "issue", "list",
					"--state", state, "--limit", "200", "--json", "number,title,labels,createdAt")
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh issue list failed: %s", errText(res))
				}
				var issues []struct {
					Number    int    `json:"number"`
					Title     string `json:"title"`
					CreatedAt string `json:"createdAt"`
				}
				if err := json.Unmarshal([]byte(res.Stdout), &issues); err != nil {
					return Result{}, fmt.Errorf("gh issue list: %w", err)
				}
				out := make([]map[string]any, 0, len(issues))
				for _, is := range issues {
					if a.Filter == "upstream" && !isUpstreamTitle(is.Title) {
						continue
					}
					out = append(out, map[string]any{"number": is.Number, "title": is.Title, "created": is.CreatedAt})
				}
				env.J.Emit("gh_issue", map[string]any{"method": "list", "filter": a.Filter, "count": len(out)})
				return Result{Summary: fmt.Sprintf("%d issue(s)", len(out)), Data: map[string]any{"issues": out}}, nil

			case "view":
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "issue", "view", strconv.Itoa(a.Number),
					"--json", "number,title,state,body,comments")
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh issue view failed: %s", errText(res))
				}
				var v map[string]any
				if err := json.Unmarshal([]byte(res.Stdout), &v); err != nil {
					return Result{}, fmt.Errorf("gh issue view: %w", err)
				}
				return Result{Summary: fmt.Sprintf("issue #%d %v", a.Number, v["state"]), Data: v}, nil

			case "comment":
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "issue", "comment", strconv.Itoa(a.Number), "--body", a.Body)
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh issue comment failed: %s", errText(res))
				}
				env.J.Emit("gh_issue", map[string]any{"method": "comment", "number": a.Number})
				return Result{Summary: fmt.Sprintf("commented on #%d", a.Number), Data: map[string]any{"ok": true}}, nil

			case "close":
				args := []string{"issue", "close", strconv.Itoa(a.Number)}
				if a.Body != "" {
					args = append(args, "--comment", a.Body)
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", args...)
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh issue close failed: %s", errText(res))
				}
				env.J.Emit("gh_issue", map[string]any{"method": "close", "number": a.Number})
				return Result{Summary: fmt.Sprintf("closed #%d", a.Number), Data: map[string]any{"closed": a.Number}}, nil
			}
			return Result{}, &HarnessGap{Tool: "gh_issue", Detail: "unknown method " + a.Method}
		},
	}
}

// GHRepoTool covers the upstream-investigation reads.
func GHRepoTool(env *Env) *Func {
	return &Func{
		Name: "gh_repo",
		Desc: "Upstream metadata: method=release_view (repo=owner/name, tag), compare (repo, base, head), api_get (path, GET only; for /repos/... endpoints).",
		Params: obj(`"method":{"type":"string","enum":["release_view","compare","api_get"]},"repo":{"type":"string","description":"owner/name"},"tag":{"type":"string"},"base":{"type":"string"},"head":{"type":"string"},"path":{"type":"string","description":"api_get: e.g. /repos/owner/name/releases/tags/v1"}`,
			"method"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Method string `json:"method"`
				Repo   string `json:"repo"`
				Tag    string `json:"tag"`
				Base   string `json:"base"`
				Head   string `json:"head"`
				Path   string `json:"path"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			switch a.Method {
			case "release_view":
				if a.Repo == "" || a.Tag == "" {
					return Result{}, fmt.Errorf("release_view needs repo and tag")
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "release", "view", a.Tag, "-R", a.Repo,
					"--json", "tagName,name,publishedAt,url,assets")
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh release view failed: %s", errText(res))
				}
				var v map[string]any
				if err := json.Unmarshal([]byte(res.Stdout), &v); err != nil {
					return Result{}, fmt.Errorf("release view: %w", err)
				}
				return Result{Summary: fmt.Sprintf("release %s", a.Tag), Data: v}, nil

			case "compare":
				if a.Repo == "" || a.Base == "" || a.Head == "" {
					return Result{}, fmt.Errorf("compare needs repo, base, head")
				}
				path := fmt.Sprintf("/repos/%s/compare/%s...%s", a.Repo, a.Base, a.Head)
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "api", path)
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh api compare failed: %s", errText(res))
				}
				var v struct {
					Status       string `json:"status"`
					AheadBy      int    `json:"ahead_by"`
					BehindBy     int    `json:"behind_by"`
					TotalCommits int    `json:"total_commits"`
					Files        []struct {
						Filename  string `json:"filename"`
						Status    string `json:"status"`
						Additions int    `json:"additions"`
						Deletions int    `json:"deletions"`
					} `json:"files"`
				}
				if err := json.Unmarshal([]byte(res.Stdout), &v); err != nil {
					return Result{}, fmt.Errorf("compare: %w", err)
				}
				files := make([]string, 0, len(v.Files))
				for _, f := range v.Files {
					if len(files) >= 200 {
						break
					}
					files = append(files, fmt.Sprintf("%s (%s +%d/-%d)", f.Filename, f.Status, f.Additions, f.Deletions))
				}
				return Result{
					Summary: fmt.Sprintf("%s: %d commits, %d files", v.Status, v.TotalCommits, len(v.Files)),
					Data:    map[string]any{"status": v.Status, "ahead_by": v.AheadBy, "behind_by": v.BehindBy, "files": files},
				}, nil

			case "api_get":
				if !strings.HasPrefix(a.Path, "/") {
					return Result{}, fmt.Errorf("api_get path must start with /")
				}
				if strings.Contains(a.Path, " ") || strings.Contains(a.Path, "--method") {
					return Result{}, fmt.Errorf("api_get accepts a plain GET path only")
				}
				res := run(ctx, env.RepoRoot, execTimeout(env), "gh", "api", a.Path)
				if res.RC != 0 {
					return Result{}, fmt.Errorf("gh api failed: %s", errText(res))
				}
				var v any
				if err := json.Unmarshal([]byte(res.Stdout), &v); err != nil {
					return Result{Summary: "api_get ok", Data: map[string]any{"raw": truncate(res.Stdout, 8000)}}, nil
				}
				b, _ := json.Marshal(v)
				return Result{Summary: fmt.Sprintf("api_get %s (%d bytes)", a.Path, len(b)), Data: map[string]any{"json": truncate(string(b), 12000)}}, nil
			}
			return Result{}, &HarnessGap{Tool: "gh_repo", Detail: "unknown method " + a.Method}
		},
	}
}

// isUpstreamTitle recognises the mechanical Issue titles bin/check-upstream creates.
func isUpstreamTitle(title string) bool {
	return strings.HasPrefix(title, "[") && strings.Contains(title, "] upstream version:")
}

// GHIssueList returns the open upstream-version Issues. The driver uses it to
// re-read the list after detect has created new ones.
func GHIssueList(ctx context.Context, env *Env) ([]map[string]any, error) {
	res, err := GHIssueTool(env).Run(ctx, json.RawMessage(`{"method":"list","filter":"upstream"}`))
	if err != nil {
		return nil, err
	}
	data, ok := res.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("gh issue list: unexpected result")
	}
	raw, ok := data["issues"].([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("gh issue list: unexpected issues payload")
	}
	return raw, nil
}
