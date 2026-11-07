package tool

import (
	"context"
	"debug/elf"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The archive tools shell out to bsdtar (libarchive) rather than linking a
// decompressor: Arch's tar reads tar/gz/xz/bz2/zip/7z, which is exactly the set
// of upstream release formats this repo has to inspect.
const archiveListLimit = 5000

// ArchiveListTool lists an archive's entries.
func ArchiveListTool(env *Env) *Func {
	return &Func{
		Name:   "archive_list",
		Desc:   "List the entries of a local archive (tar/gz/xz/zip/7z via bsdtar). Bound to the first 5000 entries.",
		Params: obj(`"path":{"type":"string","description":"path returned by fetch (saved_path) or inside the scratch dir"}`, "path"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			abs, err := resolveScratch(env, a.Path)
			if err != nil {
				return Result{}, err
			}
			res := run(ctx, env.RepoRoot, execTimeout(env), "tar", "-tf", abs)
			if res.RC != 0 {
				return Result{}, fmt.Errorf("archive_list failed: %s", errText(res))
			}
			lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
			truncated := false
			if len(lines) > archiveListLimit {
				lines = lines[:archiveListLimit]
				truncated = true
			}
			return Result{
				Summary: fmt.Sprintf("%d entries", len(lines)),
				Data:    map[string]any{"entries": lines, "truncated": truncated},
			}, nil
		},
	}
}

// ArchiveExtractTool extracts selected entries into the scratch dir.
func ArchiveExtractTool(env *Env) *Func {
	return &Func{
		Name:   "archive_extract",
		Desc:   "Extract entries matching globs from a local archive into a scratch dir. Returns the dest dir; use fs_read? no - use binary_scan/archive_diff or read the files by absolute path in later tool calls.",
		Params: obj(`"path":{"type":"string"},"globs":{"type":"array","items":{"type":"string"},"description":"tar patterns, e.g. ['*/Cargo.toml']"},"dest":{"type":"string","description":"scratch subdirectory name (default: extracted)"}`, "path"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Path  string   `json:"path"`
				Globs []string `json:"globs"`
				Dest  string   `json:"dest"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			abs, err := resolveScratch(env, a.Path)
			if err != nil {
				return Result{}, err
			}
			dest := a.Dest
			if dest == "" {
				dest = "extracted"
			}
			if strings.Contains(dest, "..") || strings.HasPrefix(dest, "/") {
				return Result{}, fmt.Errorf("dest must be a plain name")
			}
			out := filepath.Join(env.Scratch, dest)
			if err := os.MkdirAll(out, 0o755); err != nil {
				return Result{}, err
			}
			args := []string{"-xf", abs, "-C", out}
			args = append(args, a.Globs...)
			res := run(ctx, env.RepoRoot, execTimeout(env), "tar", args...)
			if res.RC != 0 {
				return Result{}, fmt.Errorf("archive_extract failed: %s", errText(res))
			}
			return Result{Summary: "extracted to " + out, Data: map[string]any{"dest": out}}, nil
		},
	}
}

// ArchiveDiffTool extracts two archives and returns unified diffs of the files
// matched by globs — the "旧新 tarball の差分" step of the upstream runbook.
func ArchiveDiffTool(env *Env) *Func {
	return &Func{
		Name: "archive_diff",
		Desc: "Extract two local archives and diff the files matched by globs (e.g. ['*/Makefile','*/Cargo.lock']). Returns per-file unified diffs (bounded).",
		Params: obj(`"a":{"type":"string","description":"old archive path"},"b":{"type":"string","description":"new archive path"},"globs":{"type":"array","items":{"type":"string"}}`,
			"a", "b"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				A     string   `json:"a"`
				B     string   `json:"b"`
				Globs []string `json:"globs"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			if len(a.Globs) == 0 {
				return Result{}, fmt.Errorf("globs is required")
			}
			pa, err := resolveScratch(env, a.A)
			if err != nil {
				return Result{}, err
			}
			pb, err := resolveScratch(env, a.B)
			if err != nil {
				return Result{}, err
			}
			da := filepath.Join(env.Scratch, "old")
			db := filepath.Join(env.Scratch, "new")
			for _, d := range []string{da, db} {
				_ = os.RemoveAll(d)
				if err := os.MkdirAll(d, 0o755); err != nil {
					return Result{}, err
				}
			}
			for archive, dest := range map[string]string{pa: da, pb: db} {
				args := []string{"-xf", archive, "-C", dest}
				args = append(args, a.Globs...)
				if res := run(ctx, env.RepoRoot, execTimeout(env), "tar", args...); res.RC != 0 {
					return Result{}, fmt.Errorf("extract %s failed: %s", filepath.Base(archive), errText(res))
				}
			}

			type fileDiff struct {
				Name string `json:"name"`
				Diff string `json:"diff"`
			}
			var diffs []fileDiff
			seen := map[string]bool{}
			for _, root := range []string{da, db} {
				_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() {
						return nil
					}
					rel, _ := filepath.Rel(root, p)
					if seen[rel] {
						return nil
					}
					seen[rel] = true
					fa := filepath.Join(da, rel)
					fb := filepath.Join(db, rel)
					if _, err := os.Stat(fa); err != nil {
						diffs = append(diffs, fileDiff{Name: rel, Diff: "only in new archive"})
						return nil
					}
					if _, err := os.Stat(fb); err != nil {
						diffs = append(diffs, fileDiff{Name: rel, Diff: "only in old archive"})
						return nil
					}
					res := run(ctx, env.RepoRoot, execTimeout(env), "diff", "-u", fa, fb)
					if strings.TrimSpace(res.Stdout) != "" {
						diffs = append(diffs, fileDiff{Name: rel, Diff: truncate(res.Stdout, 8000)})
					}
					return nil
				})
			}
			sort.Slice(diffs, func(i, j int) bool { return diffs[i].Name < diffs[j].Name })
			return Result{
				Summary: fmt.Sprintf("%d file(s) differ", len(diffs)),
				Data:    map[string]any{"diffs": diffs},
			}, nil
		},
	}
}

// BinaryScanTool reports ELF NEEDED libraries, the maximum required GLIBC
// version and the URL literals of binaries inside an archive. It replaces the
// readelf/strings pipelines (and their LC_ALL pitfall) with one typed result.
func BinaryScanTool(env *Env) *Func {
	return &Func{
		Name: "binary_scan",
		Desc: "Scan binaries inside an archive: ELF NEEDED sonames, maximum required GLIBC_x.y and the scheme-qualified URLs found in the bytes. Use it instead of readelf/strings.",
		Params: obj(`"path":{"type":"string","description":"archive path in the scratch dir"},"globs":{"type":"array","items":{"type":"string"},"description":"entries to scan, e.g. ['*/usr/bin/*']"},"dest":{"type":"string"}`,
			"path"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Path  string   `json:"path"`
				Globs []string `json:"globs"`
				Dest  string   `json:"dest"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			abs, err := resolveScratch(env, a.Path)
			if err != nil {
				return Result{}, err
			}
			dest := a.Dest
			if dest == "" {
				dest = "scan"
			}
			out := filepath.Join(env.Scratch, dest)
			_ = os.RemoveAll(out)
			if err := os.MkdirAll(out, 0o755); err != nil {
				return Result{}, err
			}
			args := []string{"-xf", abs, "-C", out}
			args = append(args, a.Globs...)
			res := run(ctx, env.RepoRoot, execTimeout(env), "tar", args...)
			if res.RC != 0 {
				return Result{}, fmt.Errorf("extract for scan failed: %s", errText(res))
			}

			type binInfo struct {
				File     string   `json:"file"`
				ELF      bool     `json:"elf"`
				Needed   []string `json:"needed,omitempty"`
				MaxGLIBC string   `json:"max_glibc,omitempty"`
				URLs     []string `json:"urls,omitempty"`
			}
			var infos []binInfo
			_ = filepath.Walk(out, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(out, p)
				bi := binInfo{File: rel}
				if f, ferr := elf.Open(p); ferr == nil {
					bi.ELF = true
					bi.Needed, _ = f.ImportedLibraries()
					bi.MaxGLIBC = maxGLIBCVersion(f)
					f.Close()
				}
				if data, rerr := os.ReadFile(p); rerr == nil && !bi.ELF {
					bi.URLs = findURLs(data, 20)
				}
				if bi.ELF || len(bi.URLs) > 0 {
					infos = append(infos, bi)
				}
				return nil
			})
			sort.Slice(infos, func(i, j int) bool { return infos[i].File < infos[j].File })
			return Result{
				Summary: fmt.Sprintf("%d file(s) scanned", len(infos)),
				Data:    map[string]any{"bins": infos},
			}, nil
		},
	}
}

// maxGLIBCVersion returns the highest GLIBC_x.y required by the binary.
func maxGLIBCVersion(f *elf.File) string {
	libs, err := f.ImportedSymbols()
	if err != nil {
		return ""
	}
	maxMajor, maxMinor := 0, 0
	for _, l := range libs {
		if !strings.HasPrefix(l.Version, "GLIBC_") {
			continue
		}
		ver := strings.TrimPrefix(l.Version, "GLIBC_")
		parts := strings.SplitN(ver, ".", 2)
		ma, _ := strconv.Atoi(parts[0])
		mi := 0
		if len(parts) > 1 {
			mi, _ = strconv.Atoi(parts[1])
		}
		if ma > maxMajor || (ma == maxMajor && mi > maxMinor) {
			maxMajor, maxMinor = ma, mi
		}
	}
	if maxMajor == 0 {
		return ""
	}
	return fmt.Sprintf("GLIBC_%d.%d", maxMajor, maxMinor)
}

var urlLiteralRe = regexp.MustCompile(`https?://[A-Za-z0-9][A-Za-z0-9._~:/?#\[\]@!$&'()*+,;=%-]{4,200}`)

func findURLs(data []byte, limit int) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range urlLiteralRe.FindAllString(string(data), -1) {
		if seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// resolveScratch confines archive paths to the scratch dir (fetch writes there).
func resolveScratch(env *Env, p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	clean := filepath.Clean(p)
	if !filepath.IsAbs(clean) {
		clean = filepath.Join(env.Scratch, clean)
	}
	if !strings.HasPrefix(clean, env.Scratch+string(os.PathSeparator)) {
		return "", fmt.Errorf("path must be inside the scratch dir: %s", p)
	}
	if _, err := os.Stat(clean); err != nil {
		return "", fmt.Errorf("archive not found: %s", p)
	}
	return clean, nil
}
