package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PKGSrcinfoTool returns `makepkg --printsrcinfo` for a package. The model uses
// it to keep .SRCINFO in sync without shell redirection.
func PKGSrcinfoTool(env *Env) *Func {
	return &Func{
		Name:   "pkg_srcinfo",
		Desc:   "Run `makepkg --printsrcinfo` for pkgs/<pkg> and return its output (use it to fix .SRCINFO).",
		Params: obj(`"pkg":{"type":"string"}`, "pkg"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Pkg string `json:"pkg"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			dir, err := pkgDir(env, a.Pkg)
			if err != nil {
				return Result{}, err
			}
			res := run(ctx, dir, execTimeout(env), "makepkg", "--printsrcinfo")
			if res.RC != 0 {
				return Result{}, fmt.Errorf("makepkg --printsrcinfo failed: %s", errText(res))
			}
			return Result{Summary: fmt.Sprintf("srcinfo for %s (%d bytes)", a.Pkg, len(res.Stdout)), Data: map[string]any{"srcinfo": res.Stdout}}, nil
		},
	}
}

// PKGVerifysourceTool downloads and verifies the package's sources.
func PKGVerifysourceTool(env *Env) *Func {
	return &Func{
		Name:   "pkg_verifysource",
		Desc:   "Run `makepkg --verifysource` for pkgs/<pkg>: downloads the sources into the pkg dir and checks the checksums. Reports rc and a stale-cache hint when makepkg reused an older file.",
		Params: obj(`"pkg":{"type":"string"}`, "pkg"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Pkg string `json:"pkg"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			dir, err := pkgDir(env, a.Pkg)
			if err != nil {
				return Result{}, err
			}
			res := run(ctx, dir, buildTimeout(env), "makepkg", "--verifysource", "--noconfirm")
			out := res.Stdout + res.Stderr
			ok := res.RC == 0
			hint := ""
			if !ok {
				low := out
				switch {
				case strings.Contains(low, "整合性チェック") || strings.Contains(low, "validity check"):
					hint = "checksum mismatch: makepkg may have reused a stale download cache, or upstream replaced the tarball. Compare the cached file's sha256 with a fresh download before changing sums."
				case strings.Contains(low, "404") || strings.Contains(low, "Not Found"):
					hint = "source URL returned 404: upstream may have withdrawn the release."
				}
			}
			env.J.Emit("pkg_verifysource", map[string]any{"pkg": a.Pkg, "ok": ok, "rc": res.RC})
			return Result{
				Summary: fmt.Sprintf("verifysource %s: ok=%v rc=%d", a.Pkg, ok, res.RC),
				Data:    map[string]any{"ok": ok, "rc": res.RC, "hint": hint, "output": truncate(out, 4000)},
			}, nil
		},
	}
}

// PKGThreePointsTool checks the three files a bot pkgrel PR must keep in sync:
// REVIEW.md (更新履歴 row), .SRCINFO (pkgrel) and .deps.lock (# MISSING lines).
func PKGThreePointsTool(env *Env) *Func {
	return &Func{
		Name:   "pkg_3points",
		Desc:   "Check the three managed files of pkgs/<pkg>: .SRCINFO vs makepkg --printsrcinfo, the # MISSING lines in .deps.lock, and whether REVIEW.md mentions the current pkgrel. Read-only; report findings to the model.",
		Params: obj(`"pkg":{"type":"string"}`, "pkg"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Pkg string `json:"pkg"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			dir, err := pkgDir(env, a.Pkg)
			if err != nil {
				return Result{}, err
			}
			type check struct {
				Name   string `json:"name"`
				OK     bool   `json:"ok"`
				Detail string `json:"detail,omitempty"`
			}
			var checks []check

			// .SRCINFO vs makepkg --printsrcinfo
			srcinfoPath := filepath.Join(dir, ".SRCINFO")
			onDisk, _ := os.ReadFile(srcinfoPath)
			res := run(ctx, dir, execTimeout(env), "makepkg", "--printsrcinfo")
			if res.RC != 0 {
				checks = append(checks, check{Name: "srcinfo_readable", OK: false, Detail: errText(res)})
			} else {
				same := normalizeSrcinfo(string(onDisk)) == normalizeSrcinfo(res.Stdout)
				detail := ""
				if !same {
					d := run(ctx, env.RepoRoot, execTimeout(env), "diff", "-u",
						writeTemp(env, "printsrcinfo", res.Stdout), srcinfoPath)
					detail = truncate(d.Stdout, 3000)
				}
				checks = append(checks, check{Name: "srcinfo_in_sync", OK: same, Detail: detail})
			}

			// .deps.lock # MISSING lines: the bot rewrites the lock from
			// header + name=version pairs and drops the "# MISSING ..." markers
			// that record AUR-only / virtual / in-repo deps. Compare with the
			// merge-base version instead of demanding markers (a pkg with no
			// such deps legitimately has none).
			lockPath := filepath.Join(dir, ".deps.lock")
			lock, _ := os.ReadFile(lockPath)
			curMissing := missingMarkers(string(lock))
			baseRef := "origin/master"
			baseHash := ""
			if r := run(ctx, env.RepoRoot, execTimeout(env), "git", "merge-base", baseRef, "HEAD"); r.RC == 0 {
				baseHash = strings.TrimSpace(r.Stdout)
			}
			lost := []string{}
			if baseHash != "" {
				r := run(ctx, env.RepoRoot, execTimeout(env), "git", "show",
					baseHash+":pkgs/"+a.Pkg+"/.deps.lock")
				if r.RC == 0 {
					baseSet := map[string]bool{}
					for _, m := range missingMarkers(r.Stdout) {
						baseSet[m] = true
					}
					for m := range baseSet {
						if !contains(curMissing, m) {
							lost = append(lost, m)
						}
					}
				}
			}
			sort.Strings(lost)
			checks = append(checks, check{
				Name: "deps_lock_missing_preserved", OK: len(lost) == 0,
				Detail: fmt.Sprintf("current=%d lost=%v", len(curMissing), lost),
			})

			// REVIEW.md mentions the current pkgrel
			review, _ := os.ReadFile(filepath.Join(dir, "REVIEW.md"))
			rel := pkgrelFromPKGBUILD(filepath.Join(dir, "PKGBUILD"))
			mentions := rel != "" && strings.Contains(string(review), "-"+rel)
			checks = append(checks, check{Name: "review_mentions_pkgrel", OK: mentions,
				Detail: fmt.Sprintf("pkgrel=%s", rel)})

			ok := true
			for _, c := range checks {
				if !c.OK {
					ok = false
				}
				env.J.Emit("verify", map[string]any{"check": "3points:" + c.Name, "pkg": a.Pkg, "ok": c.OK})
			}
			return Result{Summary: fmt.Sprintf("3-point check %s: ok=%v", a.Pkg, ok), Data: map[string]any{"ok": ok, "checks": checks}}, nil
		},
	}
}

func pkgDir(env *Env, pkg string) (string, error) {
	if !validPkgName(pkg) {
		return "", fmt.Errorf("invalid package name: %q", pkg)
	}
	dir := filepath.Join(env.RepoRoot, "pkgs", pkg)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return "", fmt.Errorf("unknown package: %s", pkg)
	}
	return dir, nil
}

// normalizeSrcinfo ignores trailing whitespace and empty lines so the
// comparison matches makepkg's output.
func normalizeSrcinfo(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		b.WriteString(strings.TrimRight(line, " \t\r"))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func writeTemp(env *Env, name, content string) string {
	p := filepath.Join(env.Scratch, name)
	_ = os.WriteFile(p, []byte(content), 0o644)
	return p
}

var pkgrelRe = regexp.MustCompile(`(?m)^pkgrel=("?)([0-9]+)`)

// missingMarkers returns the "# MISSING ..." lines of a .deps.lock (they record
// AUR-only / virtual / in-repo deps that the Arch repo snapshot cannot resolve).
func missingMarkers(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# MISSING") {
			out = append(out, t)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func pkgrelFromPKGBUILD(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	m := pkgrelRe.FindStringSubmatch(string(b))
	if len(m) < 3 {
		return ""
	}
	return m[2]
}

var _ = strconv.Itoa
