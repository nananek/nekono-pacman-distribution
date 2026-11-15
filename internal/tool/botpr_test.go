package tool

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
)

// BotPRList has to return the shape bin/preflight reports (headRefName), so the
// driver can swap the preflight list for the re-read one without special cases.
// gh_pr normalises the head branch to "head", so this pins the mapping.
func TestBotPRListNormalisesHead(t *testing.T) {
	dir := t.TempDir()
	script := `#!/bin/sh
cat <<'JSON'
[{"number":809,"headRefName":"deps/nekonolock-pkgrel-2","title":"nekonolock: pkgrel bump to 2 (deps changed)","author":{"login":"nananek"},"isDraft":false},
 {"number":810,"headRefName":"master","title":"not a deps branch","author":{"login":"nananek"},"isDraft":false}]
JSON
`
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	env := &Env{
		RepoRoot: dir,
		Cfg:      &cfg.Config{ExecTimeoutSec: 30},
		J:        journal.New(io.Discard, io.Discard),
	}
	prs, err := BotPRList(context.Background(), env)
	if err != nil {
		t.Fatalf("BotPRList: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("want 1 deps/* PR, got %d: %v", len(prs), prs)
	}
	if got := prs[0]["headRefName"]; got != "deps/nekonolock-pkgrel-2" {
		t.Errorf("headRefName = %v, want deps/nekonolock-pkgrel-2", got)
	}
	if got := prs[0]["number"]; got != 809 {
		t.Errorf("number = %v, want 809", got)
	}
}
