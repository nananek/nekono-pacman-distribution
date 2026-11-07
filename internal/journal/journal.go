// Package journal is the append-only event log of a daily-ops run.
//
// Two streams come out of one run:
//
//   - ndjson events on stdout (kind + fields + timestamp). The host side
//     (nekono-build-vm / quagent) tails this to see progress while a phase is
//     running, which the OpenCode-based harness could not do: its output was
//     only summarised at the end of the run.
//   - human text on stderr, for the same run when a person is watching.
//
// Values that must never leave the machine are redacted here rather than at
// each call site: tailnet hostnames are replaced with <host>.
package journal

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"
)

// redactions are applied to every string field before it is written.
var redactions = []struct {
	re   *regexp.Regexp
	with string
}{
	// Tailnet hosts (nekono-dist0.tail2c8c7.ts.net, ocgo.tail2c8c7.ts.net, ...).
	{regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9-]*\.[a-z0-9]+\.ts\.net`), "<host>"},
	// Any remaining *.ts.net host.
	{regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9.-]*\.ts\.net`), "<host>"},
}

// Redact removes machine-specific hostnames from s.
func Redact(s string) string {
	for _, r := range redactions {
		s = r.re.ReplaceAllString(s, r.with)
	}
	return s
}

// Journal writes ndjson events and human-readable progress.
type Journal struct {
	mu    sync.Mutex
	out   io.Writer
	human io.Writer
	now   func() time.Time
}

// New returns a journal writing events to out and human text to human.
func New(out, human io.Writer) *Journal {
	return &Journal{out: out, human: human, now: time.Now}
}

// Emit appends one event. Fields are redacted and must be JSON-marshalable.
func (j *Journal) Emit(kind string, fields map[string]any) {
	ev := make(map[string]any, len(fields)+2)
	ev["ts"] = j.now().UTC().Format(time.RFC3339)
	ev["kind"] = kind
	for k, v := range fields {
		ev[k] = redactValue(v)
	}
	b, err := json.Marshal(ev)
	if err != nil {
		b, _ = json.Marshal(map[string]any{"ts": ev["ts"], "kind": kind, "error": "marshal failed"})
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	fmt.Fprintf(j.out, "%s\n", b)
}

// Human prints progress for a person watching the run (never machine-read).
func (j *Journal) Human(format string, a ...any) {
	j.mu.Lock()
	defer j.mu.Unlock()
	fmt.Fprintf(j.human, format+"\n", a...)
}

func redactValue(v any) any {
	switch t := v.(type) {
	case string:
		return Redact(t)
	case []string:
		out := make([]string, len(t))
		for i := range t {
			out[i] = Redact(t[i])
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = redactValue(t[i])
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, x := range t {
			out[k] = redactValue(x)
		}
		return out
	default:
		return v
	}
}
