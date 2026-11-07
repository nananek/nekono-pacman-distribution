package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nananek/nekono-pacman-distribution/internal/journal"
)

// CmdResult is the outcome of one process the harness ran on the model's behalf.
type CmdResult struct {
	RC       int
	Stdout   string
	Stderr   string
	TimedOut bool
}

// run executes name with a fixed argv (never a shell string) in dir.
func run(ctx context.Context, dir string, timeout time.Duration, name string, args ...string) CmdResult {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	cmd.Dir = dir
	// Deterministic output regardless of the host locale (the old harness hit
	// this with readelf: without LC_ALL=C the awk extraction came back empty).
	cmd.Env = append(cmd.Environ(), "LC_ALL=C", "LANG=C")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	res := CmdResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var ee *exec.ExitError
		switch {
		case errors.As(err, &ee):
			res.RC = ee.ExitCode()
		case cctx.Err() == context.DeadlineExceeded:
			res.RC = -1
			res.TimedOut = true
		default:
			res.RC = -2
			res.Stderr += "\n" + err.Error()
		}
	}
	return res
}

// runStreaming runs a long process and journals its notable stderr lines while
// it runs, so the host sees progress instead of waiting for the whole build.
func runStreaming(ctx context.Context, env *Env, stage, dir string, timeout time.Duration, name string, args ...string) CmdResult {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(), "LC_ALL=C", "LANG=C")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return CmdResult{RC: -2, Stderr: err.Error()}
	}
	if err := cmd.Start(); err != nil {
		return CmdResult{RC: -2, Stderr: err.Error()}
	}

	var stderr bytes.Buffer
	buf := make([]byte, 0, 4096)
	for {
		n, rerr := stderrPipe.Read(buf[:cap(buf)])
		if n > 0 {
			chunk := string(buf[:n])
			stderr.WriteString(chunk)
			for _, line := range strings.Split(chunk, "\n") {
				line = strings.TrimSpace(line)
				if line == "" || !notableLine(line) {
					continue
				}
				env.J.Emit("progress", map[string]any{"stage": stage, "line": line})
				env.J.Human("  %s", line)
			}
		}
		if rerr != nil {
			break
		}
	}

	res := CmdResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if werr := cmd.Wait(); werr != nil {
		var ee *exec.ExitError
		switch {
		case errors.As(werr, &ee):
			res.RC = ee.ExitCode()
		case cctx.Err() == context.DeadlineExceeded:
			res.RC = -1
			res.TimedOut = true
		default:
			res.RC = -2
			res.Stderr += "\n" + werr.Error()
		}
	}
	return res
}

// notableLine filters the build/publish log down to the lines a progress
// watcher cares about (phases, failures, artifact hand-offs).
func notableLine(s string) bool {
	prefixes := []string{
		"[*]", "[pending]", "[cascade-warn]", "[OK]", "[FATAL]", "[ERROR]",
		"[WARN]", "[!]", "[daily-ops]", "Building ", "Retiring ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// Exec runs a command with a fixed argv. Exported for the driver, which needs
// the deterministic entrypoints without going through the model.
func Exec(ctx context.Context, dir string, timeout time.Duration, name string, args ...string) CmdResult {
	return run(ctx, dir, timeout, name, args...)
}

// ExecJSON runs a --json entrypoint and decodes its single JSON document.
func ExecJSON(ctx context.Context, dir string, timeout time.Duration, name string, args ...string) (map[string]any, CmdResult, error) {
	res := run(ctx, dir, timeout, name, args...)
	line := strings.TrimSpace(res.Stdout)
	if i := strings.LastIndexByte(line, '\n'); i >= 0 {
		line = line[i+1:]
	}
	var data map[string]any
	if line == "" {
		return nil, res, fmt.Errorf("%s %s: no JSON on stdout (rc=%d)", name, strings.Join(args, " "), res.RC)
	}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return nil, res, fmt.Errorf("%s %s: invalid JSON: %w", name, strings.Join(args, " "), err)
	}
	return data, res, nil
}

// tail returns at most n trailing lines of s.
func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// truncate bounds a string for the model context.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	head := max * 3 / 4
	return s[:head] + fmt.Sprintf("\n... [truncated %d bytes] ...\n", len(s)-max) + tail(s, (max-head)/80)
}

// errText turns a CmdResult into a short error string.
func errText(res CmdResult) string {
	if res.TimedOut {
		return "timed out"
	}
	return strings.TrimSpace(tail(res.Stderr, 20))
}

var _ = journal.Redact
