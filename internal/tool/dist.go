package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DistPublishTool pushes the signed repo to the distribution host.
func DistPublishTool(env *Env) *Func {
	return &Func{
		Name:    "dist_publish",
		Desc:    "Run `bin/publish` (two passes: sync packages, then force-refresh db metadata + signatures). Returns copied/deleted counts and the real exit code.",
		Params:  obj(``),
		Mutates: true,
		Run: func(ctx context.Context, _ json.RawMessage) (Result, error) {
			res := runStreaming(ctx, env, "publish", env.RepoRoot, buildTimeout(env), "bin/publish", "--json")
			var data map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(strings.Split(res.Stdout, "\n")[0])), &data); err != nil {
				return Result{}, &HarnessGap{Tool: "dist_publish", Detail: fmt.Sprintf("bin/publish --json produced no JSON (rc=%d): %s", res.RC, errText(res))}
			}
			env.J.Emit("publish", map[string]any{
				"ok": data["ok"], "pass": data["pass"], "copied_new": data["copied_new"],
				"deleted": data["deleted"], "rc": data["rc"], "error": data["error"],
			})
			return Result{Summary: fmt.Sprintf("publish ok=%v copied=%v deleted=%v rc=%v", data["ok"], data["copied_new"], data["deleted"], data["rc"]), Data: data}, nil
		},
	}
}

// DistVerifyTool compares what a pacman client would download with what was
// built, over the same anonymous path clients use.
//
// The server URL is never returned to the model or journaled: only MATCH /
// MISMATCH per file (the old runbook made the same rule for `$srv`).
func DistVerifyTool(env *Env) *Func {
	return &Func{
		Name:   "dist_verify",
		Desc:   "Compare published files against the local repo over the anonymous client path: sha256 per file, plus 404 checks for retired packages. Reports MATCH/MISMATCH only.",
		Params: obj(`"files":{"type":"array","items":{"type":"string"},"description":"file names in repo/x86_64 (db, sig, packages)"},"retired":{"type":"array","items":{"type":"string"},"description":"retired package file names that must return 404"}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Files   []string `json:"files"`
				Retired []string `json:"retired"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			srv, err := distServer(env)
			if err != nil {
				return Result{}, err
			}
			client := &http.Client{Timeout: 90 * time.Second}
			type fileCheck struct {
				File   string `json:"file"`
				Result string `json:"result"`
				Detail string `json:"detail,omitempty"`
			}
			var checks []fileCheck
			for _, f := range a.Files {
				local, err := sha256File(filepath.Join(env.RepoRoot, "repo", "x86_64", f))
				if err != nil {
					checks = append(checks, fileCheck{File: f, Result: "LOCAL_MISSING", Detail: err.Error()})
					continue
				}
				remote, code, err := sha256URL(ctx, client, srv+"/"+f)
				switch {
				case err != nil || code != http.StatusOK:
					checks = append(checks, fileCheck{File: f, Result: "MISMATCH", Detail: fmt.Sprintf("http %d", code)})
				case local == remote:
					checks = append(checks, fileCheck{File: f, Result: "MATCH"})
				default:
					checks = append(checks, fileCheck{File: f, Result: "MISMATCH", Detail: "sha256 differs"})
				}
			}
			for _, f := range a.Retired {
				_, code, _ := sha256URL(ctx, client, srv+"/"+f)
				r := "STILL_PUBLISHED"
				if code == http.StatusNotFound {
					r = "GONE"
				}
				checks = append(checks, fileCheck{File: f, Result: r, Detail: fmt.Sprintf("http %d", code)})
			}
			ok := true
			for _, c := range checks {
				if c.Result != "MATCH" && c.Result != "GONE" {
					ok = false
				}
				env.J.Emit("verify", map[string]any{"check": "dist_hash", "file": c.File, "result": c.Result})
			}
			return Result{Summary: fmt.Sprintf("distribution verification ok=%v (%d files)", ok, len(checks)), Data: map[string]any{"ok": ok, "checks": checks}}, nil
		},
	}
}

// distServer resolves the [nekono] Server line, preferring the configured file.
func distServer(env *Env) (string, error) {
	path := env.Cfg.DistServerPath
	if path == "" {
		path = "/etc/pacman.conf"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", &HarnessGap{Tool: "dist_verify", Detail: fmt.Sprintf("cannot read %s for the [nekono] Server line", path)}
	}
	var srv string
	inSection := false
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "[") {
			inSection = t == "[nekono]"
			continue
		}
		if !inSection || !strings.HasPrefix(t, "Server") {
			continue
		}
		parts := strings.SplitN(t, "=", 2)
		if len(parts) != 2 {
			continue
		}
		srv = strings.TrimSpace(parts[1])
		srv = strings.ReplaceAll(srv, "$repo", "nekono")
		srv = strings.ReplaceAll(srv, "$arch", "x86_64")
		break
	}
	if srv == "" {
		return "", &HarnessGap{Tool: "dist_verify", Detail: "no Server line in the [nekono] section"}
	}
	return strings.TrimSuffix(srv, "/"), nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sha256URL(ctx context.Context, c *http.Client, url string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	h := sha256.New()
	_, _ = io.Copy(h, io.LimitReader(resp.Body, 1<<31))
	return hex.EncodeToString(h.Sum(nil)), resp.StatusCode, nil
}
