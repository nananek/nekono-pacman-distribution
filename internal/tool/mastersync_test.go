package tool

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
)

// RestoreMaster is the recovery that keeps a run alive after a phase left a PR
// branch checked out: triage takes deps/* to review it, and a finish without
// merging used to leave HEAD there, so build_plan refused and the merged
// package of the same run was never built.
func TestRestoreMaster(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	work := filepath.Join(root, "work")

	git := func(args ...string) {
		t.Helper()
		full := append([]string{"-C", work, "-c", "commit.gpgsign=false",
			"-c", "user.name=test", "-c", "user.email=test@example.invalid"}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	head := func() string {
		t.Helper()
		out, err := exec.Command("git", "-C", work, "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			t.Fatalf("rev-parse: %v", err)
		}
		return strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", root, "init", "--quiet", "--bare", origin).CombinedOutput(); err != nil {
		t.Fatalf("init bare: %v: %s", err, out)
	}
	if out, err := exec.Command("git", "-C", root, "init", "--quiet", work).CombinedOutput(); err != nil {
		t.Fatalf("init work: %v: %s", err, out)
	}
	git("checkout", "--quiet", "-b", "master")
	if err := os.WriteFile(filepath.Join(work, "PKGBUILD"), []byte("pkgname=x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "--quiet", "-m", "init")
	git("remote", "add", "origin", origin)
	git("push", "--quiet", "-u", "origin", "master")

	env := &Env{RepoRoot: work, Cfg: &cfg.Config{ExecTimeoutSec: 30}, J: journal.New(io.Discard, io.Discard)}

	// Already on master: nothing to restore.
	if branch, err := RestoreMaster(ctx, env); err != nil || branch != "" {
		t.Fatalf("master must be a no-op, got branch=%q err=%v", branch, err)
	}

	// The phase left the deps PR checked out (clean).
	git("checkout", "--quiet", "-b", "deps/x-pkgrel-2")
	branch, err := RestoreMaster(ctx, env)
	if err != nil {
		t.Fatalf("clean PR branch must be restored, got %v", err)
	}
	if branch != "deps/x-pkgrel-2" {
		t.Fatalf("restore must report the branch it left, got %q", branch)
	}
	if got := head(); got != "master" {
		t.Fatalf("HEAD must be master after restore, got %q", got)
	}

	// Uncommitted model edits on the branch stay where they are; the build guard
	// reports the gap instead of the harness throwing the edits away.
	git("checkout", "--quiet", "deps/x-pkgrel-2")
	if err := os.WriteFile(filepath.Join(work, "PKGBUILD"), []byte("pkgname=x\npkgrel=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RestoreMaster(ctx, env); err == nil {
		t.Fatal("dirty tree on a PR branch must not be switched silently")
	}
	if got := head(); got != "deps/x-pkgrel-2" {
		t.Fatalf("dirty tree must be left untouched, HEAD moved to %q", got)
	}
}

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
