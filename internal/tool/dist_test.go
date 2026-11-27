package tool

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
)

// Retired package names can only come from dist_publish's deleted_files: the
// publish phase cannot list repo/x86_64, so without them the 404 check of a
// pruned version is impossible and the phase escalated as a harness gap
// (2026-11-28). This pins both directions: gone -> ok, still served -> not ok.
func TestDistVerifyRetired(t *testing.T) {
	dir := t.TempDir()
	repodir := filepath.Join(dir, "repo", "x86_64")
	if err := os.MkdirAll(repodir, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := "keep-1.0-1-x86_64.pkg.tar.zst"
	if err := os.WriteFile(filepath.Join(repodir, keep), []byte("pkg"), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/x86_64/"+keep {
			_, _ = w.Write([]byte("pkg"))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	conf := filepath.Join(dir, "pacman.conf")
	if err := os.WriteFile(conf, []byte("[nekono]\nServer = "+srv.URL+"/x86_64\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := &Env{RepoRoot: dir, Cfg: &cfg.Config{DistServerPath: conf, ExecTimeoutSec: 30}, J: journal.New(io.Discard, io.Discard)}

	run := func(retired []string) map[string]any {
		t.Helper()
		args, _ := json.Marshal(map[string]any{"files": []string{keep}, "retired": retired})
		res, err := DistVerifyTool(env).Run(context.Background(), args)
		if err != nil {
			t.Fatalf("dist_verify: %v", err)
		}
		data, _ := res.Data.(map[string]any)
		return data
	}

	gone := run([]string{"gone-0.9-1-x86_64.pkg.tar.zst"})
	if ok, _ := gone["ok"].(bool); !ok {
		t.Fatalf("a 404 retired file must verify ok, got %v", gone["checks"])
	}

	stale := run([]string{keep})
	if ok, _ := stale["ok"].(bool); ok {
		t.Fatalf("a retired file that is still served must fail, got %v", stale["checks"])
	}
}
