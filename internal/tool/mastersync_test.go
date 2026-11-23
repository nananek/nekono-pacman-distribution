package tool

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
)

// requireCleanMaster is the guard that keeps build/publish on committed master:
// a bump left in the working tree by an earlier item must never be built or
// served (2026-11-20: genshin-dict 7.2.0 / eleos 0.8.7 were published while
// master stayed on the older versions with their PRs open).
func TestRequireCleanMaster(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	work := filepath.Join(root, "work")

	git := func(dir string, args ...string) {
		t.Helper()
		full := append([]string{"-C", dir, "-c", "commit.gpgsign=false",
			"-c", "user.name=test", "-c", "user.email=test@example.invalid"}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	git(root, "init", "--quiet", "--bare", origin)
	git(root, "init", "--quiet", work)
	git(work, "checkout", "--quiet", "-b", "master")
	if err := os.WriteFile(filepath.Join(work, "PKGBUILD"), []byte("pkgname=x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(work, "add", "-A")
	git(work, "commit", "--quiet", "-m", "init")
	git(work, "remote", "add", "origin", origin)
	git(work, "push", "--quiet", "-u", "origin", "master")

	env := &Env{RepoRoot: work, Cfg: &cfg.Config{ExecTimeoutSec: 30}, J: journal.New(io.Discard, io.Discard)}

	if err := requireCleanMaster(ctx, env, "build_plan"); err != nil {
		t.Fatalf("clean master at origin tip must pass, got %v", err)
	}

	// A PKGBUILD edit nobody committed yet (an unmerged bump).
	if err := os.WriteFile(filepath.Join(work, "PKGBUILD"), []byte("pkgname=x\npkgver=9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := requireCleanMaster(ctx, env, "build_plan")
	if _, ok := err.(*HarnessGap); !ok {
		t.Fatalf("dirty tree must be a HarnessGap, got %#v", err)
	}
	git(work, "checkout", "--quiet", "--", "PKGBUILD")

	// Committed but not pushed: still not the master clients see.
	git(work, "commit", "--quiet", "--allow-empty", "-m", "local only")
	if _, ok := requireCleanMaster(ctx, env, "build_run").(*HarnessGap); !ok {
		t.Fatal("HEAD != origin/master must be a HarnessGap")
	}
}
