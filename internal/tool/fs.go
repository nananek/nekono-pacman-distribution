package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// deniedPaths are unreadable through fs_read. They are the secrets the old
// harness denied with `read:*publish.env*` policies; keeping the list in code
// makes it independent of a config a run could get wrong.
var deniedPaths = []*regexp.Regexp{
	regexp.MustCompile(`(^|/)publish\.env$`),
	regexp.MustCompile(`(^|/)\.ssh(/|$)`),
	regexp.MustCompile(`remote-gpg-agent`),
	regexp.MustCompile(`\.key$`),
	regexp.MustCompile(`(^|/)\.gnupg(/|$)`),
}

// managedFiles are the only files fs_write may touch: one package's four
// managed files. Anything else is a capability gap, not a retry.
var managedFiles = regexp.MustCompile(`^pkgs/[^/]+/(PKGBUILD|REVIEW\.md|\.SRCINFO|\.deps\.lock)$`)

func resolveInRepo(env *Env, p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	clean := filepath.Clean(p)
	if filepath.IsAbs(clean) {
		// Only /etc/pacman.conf is readable from outside the repo (publish
		// verification needs the [nekono] Server line).
		if clean == "/etc/pacman.conf" {
			return clean, nil
		}
		return "", fmt.Errorf("absolute path not allowed: %s", p)
	}
	abs := filepath.Join(env.RepoRoot, clean)
	if !strings.HasPrefix(abs, env.RepoRoot+string(os.PathSeparator)) && abs != env.RepoRoot {
		return "", fmt.Errorf("path escapes the repository: %s", p)
	}
	for _, re := range deniedPaths {
		if re.MatchString(abs) {
			return "", fmt.Errorf("path is denied by policy: %s", p)
		}
	}
	return abs, nil
}

// FSReadTool reads a repo file.
func FSReadTool(env *Env) *Func {
	return &Func{
		Name:   "fs_read",
		Desc:   "Read a file in the repository (or /etc/pacman.conf). Secrets (publish.env, ssh keys, gpg agent config) are denied.",
		Params: obj(`"path":{"type":"string","description":"repo-relative path"},"max_bytes":{"type":"integer","description":"default 60000"}`),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Path     string `json:"path"`
				MaxBytes int    `json:"max_bytes"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			abs, err := resolveInRepo(env, a.Path)
			if err != nil {
				return Result{}, err
			}
			b, err := os.ReadFile(abs)
			if err != nil {
				return Result{}, fmt.Errorf("read %s: %w", a.Path, err)
			}
			limit := a.MaxBytes
			if limit <= 0 || limit > 200000 {
				limit = 60000
			}
			env.J.Emit("fs_read", map[string]any{"path": a.Path, "bytes": len(b)})
			return Result{
				Summary: fmt.Sprintf("read %s (%d bytes)", a.Path, len(b)),
				Data:    map[string]any{"path": a.Path, "content": truncate(string(b), limit)},
			}, nil
		},
	}
}

// FSWriteTool writes one managed package file.
func FSWriteTool(env *Env) *Func {
	return &Func{
		Name:    "fs_write",
		Desc:    "Write one managed file under pkgs/<pkg>/ (PKGBUILD, REVIEW.md, .SRCINFO, .deps.lock). The old/new hashes are journaled for audit; use repo_diff afterwards to review.",
		Params:  obj(`"path":{"type":"string"},"content":{"type":"string"},"reason":{"type":"string","description":"why this edit is needed (journaled)"}`, "path", "content"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Path    string `json:"path"`
				Content string `json:"content"`
				Reason  string `json:"reason"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			clean := filepath.Clean(a.Path)
			if !managedFiles.MatchString(clean) {
				return Result{}, &HarnessGap{
					Tool:   "fs_write",
					Detail: fmt.Sprintf("%s is outside the managed set (pkgs/<pkg>/{PKGBUILD,REVIEW.md,.SRCINFO,.deps.lock})", a.Path),
				}
			}
			abs, err := resolveInRepo(env, clean)
			if err != nil {
				return Result{}, err
			}
			before, _ := os.ReadFile(abs)
			oldSum := sha256.Sum256(before)
			newSum := sha256.Sum256([]byte(a.Content))
			if err := os.WriteFile(abs, []byte(a.Content), 0o644); err != nil {
				return Result{}, fmt.Errorf("write %s: %w", clean, err)
			}
			env.J.Emit("fs_write", map[string]any{
				"path":     clean,
				"reason":   a.Reason,
				"old_sha":  hex.EncodeToString(oldSum[:8]),
				"new_sha":  hex.EncodeToString(newSum[:8]),
				"old_size": len(before),
				"new_size": len(a.Content),
			})
			return Result{Summary: fmt.Sprintf("wrote %s (%d -> %d bytes)", clean, len(before), len(a.Content)), Data: map[string]any{"path": clean, "bytes": len(a.Content)}}, nil
		},
	}
}
