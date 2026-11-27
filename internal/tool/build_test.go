package tool

import (
	"fmt"
	"strings"
	"testing"
)

// A failing makepkg only explains itself on stderr, so build_run keeps the end
// of it for the phase. It must stay bounded: one line of a chatty compiler log
// must not push the model context over.
func TestLogTail(t *testing.T) {
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %03d", i)
	}
	out := logTail(strings.Join(lines, "\n"), 80, 8000)
	kept := strings.Split(out, "\n")
	if len(kept) != 80 {
		t.Fatalf("want 80 lines, got %d", len(kept))
	}
	if kept[0] != "line 120" || kept[79] != "line 199" {
		t.Fatalf("tail must be the end of the log, got %q .. %q", kept[0], kept[79])
	}

	if got := logTail("short", 80, 8000); got != "short" {
		t.Fatalf("short log must pass through, got %q", got)
	}

	big := strings.Repeat("x", 20000)
	if got := logTail(big, 5, 100); len(got) > 100 {
		t.Fatalf("byte cap not applied: %d bytes", len(got))
	}
}
