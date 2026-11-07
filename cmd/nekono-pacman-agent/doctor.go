package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/tool"
)

type checkRow struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Optional bool   `json:"optional,omitempty"`
	Detail   string `json:"detail"`
}

// cmdDoctor checks everything a run needs before it starts. It never prints
// secret values (only presence, mode and reachability).
func cmdDoctor(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	cfgPath := fs.String("config", "agent/agent.json", "path to agent config")
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := filepath.Abs(*repo)
	if err != nil {
		return err
	}
	var rows []checkRow
	add := func(name string, ok bool, detail string) {
		rows = append(rows, checkRow{Name: name, OK: ok, Detail: detail})
	}

	c, cerr := cfg.Load(filepath.Join(root, *cfgPath))
	if cerr != nil {
		add("config", false, cerr.Error())
	} else {
		add("config", true, fmt.Sprintf("model=%s base_url=%s phases=%d", c.Model, c.BaseURL, len(c.Phases)))
	}

	if cerr == nil {
		if k, kerr := c.Key(); kerr != nil {
			add("api_key", false, kerr.Error())
		} else {
			add("api_key", true, fmt.Sprintf("loaded (%d bytes, value not shown)", len(k)))
			rows = append(rows, probeEndpoint(ctx, c, k))
		}
	}

	// gh
	if res := tool.Exec(ctx, root, 30*time.Second, "gh", "auth", "status"); res.RC == 0 {
		add("gh_auth", true, "authenticated")
	} else {
		add("gh_auth", false, "gh auth status failed (run: gh auth login)")
	}

	// signing + repo state come from bin/preflight --json (one source of truth
	// for the VM/host signing difference).
	if data, res, err := tool.ExecJSON(ctx, root, 120*time.Second, "bin/preflight", "--json"); err != nil {
		add("preflight", false, fmt.Sprintf("bin/preflight --json failed (rc=%d)", res.RC))
	} else {
		add("preflight", true, "bin/preflight --json ok")
		ok, _ := data["signing_ok"].(bool)
		detail, _ := data["signing_detail"].(string)
		add("signing", ok, detail)
		if ok2, _ := data["build_host"].(bool); ok2 {
			add("build_host", true, "build/publish can run here")
		} else {
			// Informational: a non-build-host is a normal state (the VM that
			// builds mounts the artifacts disk). It must not fail doctor.
			rows = append(rows, checkRow{Name: "build_host", OK: false, Optional: true,
				Detail: "not a build host (artifacts db / publish.env / signing missing here)"})
		}
		if pub, _ := data["publish"].(string); pub != "" {
			add("publish_state", true, pub)
		}
	}

	// every entrypoint the agent calls must support --json
	for _, script := range []string{"preflight", "build-all", "publish", "check-upstream", "check-deps"} {
		b, rerr := os.ReadFile(filepath.Join(root, "bin", script))
		ok := rerr == nil && strings.Contains(string(b), "--json")
		add("json:"+script, ok, filepath.Join("bin", script))
	}

	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(map[string]any{"kind": "doctor", "checks": rows})
	failed := 0
	for _, r := range rows {
		mark := "OK  "
		if !r.OK {
			mark = "FAIL"
			if r.Optional {
				mark = "WARN"
			} else {
				failed++
			}
		}
		fmt.Fprintf(os.Stderr, "%-5s %-16s %s\n", mark, r.Name, r.Detail)
	}
	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	return nil
}

func probeEndpoint(ctx context.Context, c *cfg.Config, key string) checkRow {
	if c.BaseURL == "" {
		return checkRow{Name: "endpoint", OK: false, Detail: "base_url is empty in the config"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(c.BaseURL, "/")+"/models", nil)
	if err != nil {
		return checkRow{Name: "endpoint", OK: false, Detail: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return checkRow{Name: "endpoint", OK: false, Detail: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return checkRow{Name: "endpoint", OK: false, Detail: fmt.Sprintf("GET /models: http %d", resp.StatusCode)}
	}
	return checkRow{Name: "endpoint", OK: true, Detail: "GET /models: 200"}
}
