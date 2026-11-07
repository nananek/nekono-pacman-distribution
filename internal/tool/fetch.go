package tool

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// fetchBodyLimit is how much of the response body is returned to the model.
const fetchBodyLimit = 96 * 1024

// fetchHardLimit bounds a single download (tarballs); larger responses are
// refused rather than filling the scratch dir.
const fetchHardLimit = 512 * 1024 * 1024

// defaultFetchHosts is the static allowlist for supply-chain investigation.
// A pkg's own source= hosts are added on top (see pkgSourceHosts).
var defaultFetchHosts = []string{
	"github.com",
	"api.github.com",
	"objects.githubusercontent.com",
	"codeload.github.com",
	"raw.githubusercontent.com",
	"aur.archlinux.org",
	"registry.npmjs.org",
	"pypi.org",
	"files.pythonhosted.org",
	"crates.io",
	"static.crates.io",
	"index.crates.io",
	"dl.google.com",
	"releases.astral.sh",
}

var sourceURLRe = regexp.MustCompile(`https?://[^\s"'()]+`)

// FetchTool downloads a URL and returns its hash, so "measured sha256" is a tool
// result rather than a shell pipeline. Hosts are allow-listed; redirects are
// re-checked at every hop.
func FetchTool(env *Env) *Func {
	return &Func{
		Name: "fetch",
		Desc: "HTTP GET an allow-listed host. Returns status, byte count and the sha256 of the raw bytes; pass save=true to also write the body to the scratch dir (saved_path) for archive_* tools. Optional pkg= adds the hosts of that PKGBUILD's source= URLs.",
		Params: obj(`"url":{"type":"string"},"pkg":{"type":"string","description":"package whose source= hosts are allowed in addition to the defaults"},"save":{"type":"boolean"},"max_bytes":{"type":"integer","description":"refuse larger responses (default 512MiB)"}`,
			"url"),
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				URL      string `json:"url"`
				Pkg      string `json:"pkg"`
				Save     bool   `json:"save"`
				MaxBytes int64  `json:"max_bytes"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			u, err := url.Parse(a.URL)
			if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
				return Result{}, fmt.Errorf("invalid url: %s", a.URL)
			}
			allow := allowedHosts(env, a.Pkg)
			if !allow[u.Host] {
				return Result{}, &HarnessGap{
					Tool:   "fetch",
					Detail: fmt.Sprintf("host %q is not allow-listed (phase %s, pkg %q)", u.Host, env.Phase, a.Pkg),
				}
			}
			limit := a.MaxBytes
			if limit <= 0 || limit > fetchHardLimit {
				limit = fetchHardLimit
			}

			client := &http.Client{
				Timeout: 10 * time.Minute,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					if len(via) >= 10 {
						return fmt.Errorf("too many redirects")
					}
					if !allow[req.URL.Host] {
						return fmt.Errorf("redirect to non-allow-listed host %q", req.URL.Host)
					}
					return nil
				},
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
			if err != nil {
				return Result{}, err
			}
			req.Header.Set("User-Agent", "nekono-pacman-agent")
			resp, err := client.Do(req)
			if err != nil {
				return Result{}, fmt.Errorf("fetch %s: %w", u.Host, err)
			}
			defer resp.Body.Close()

			h := sha256.New()
			var head bytes.Buffer
			w := io.MultiWriter(h, &boundedWriter{w: &head, limit: fetchBodyLimit})
			var saved string
			if a.Save {
				name := filepath.Base(u.Path)
				if name == "" || name == "/" || name == "." {
					name = "download"
				}
				saved = filepath.Join(env.Scratch, name)
				f, ferr := os.Create(saved)
				if ferr != nil {
					return Result{}, ferr
				}
				defer f.Close()
				w = io.MultiWriter(w, f)
			}
			n, cerr := io.Copy(w, io.LimitReader(resp.Body, limit+1))
			if cerr != nil {
				return Result{}, fmt.Errorf("fetch body: %w", cerr)
			}
			if n > limit {
				return Result{}, fmt.Errorf("response exceeds %d bytes", limit)
			}
			sum := hex.EncodeToString(h.Sum(nil))

			env.J.Emit("fetch", map[string]any{
				"host": u.Host, "status": resp.StatusCode, "bytes": n,
				"sha256": sum, "saved": saved != "",
			})
			data := map[string]any{
				"status": resp.StatusCode, "bytes": n, "sha256": sum,
				"content_type": resp.Header.Get("Content-Type"),
			}
			if saved != "" {
				data["saved_path"] = saved
			}
			if head.Len() > 0 {
				data["body"] = head.String()
			}
			return Result{
				Summary: fmt.Sprintf("%s: http %d, %d bytes, sha256=%s", u.Host, resp.StatusCode, n, sum[:16]),
				Data:    data,
			}, nil
		},
	}
}

// boundedWriter keeps the first limit bytes and silently drops the rest; the
// copy still sees the full length so byte counts and hashes stay correct.
type boundedWriter struct {
	w     io.Writer
	limit int
	n     int
}

func (b *boundedWriter) Write(p []byte) (int, error) {
	if b.n < b.limit {
		room := b.limit - b.n
		if room > len(p) {
			room = len(p)
		}
		if _, err := b.w.Write(p[:room]); err != nil {
			return 0, err
		}
		b.n += room
	}
	return len(p), nil
}

// allowedHosts merges the static list, the config list and (optionally) the
// hosts referenced by one package's PKGBUILD.
func allowedHosts(env *Env, pkg string) map[string]bool {
	out := map[string]bool{}
	for _, h := range defaultFetchHosts {
		out[h] = true
	}
	if env.Cfg != nil {
		for _, h := range env.Cfg.FetchAllowHosts {
			out[h] = true
		}
	}
	if pkg != "" {
		for _, h := range pkgSourceHosts(env, pkg) {
			out[h] = true
		}
	}
	return out
}

// pkgSourceHosts extracts hosts from pkgs/<pkg>/PKGBUILD source= URLs.
func pkgSourceHosts(env *Env, pkg string) []string {
	if !validPkgName(pkg) {
		return nil
	}
	b, err := os.ReadFile(filepath.Join(env.RepoRoot, "pkgs", pkg, "PKGBUILD"))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var hosts []string
	for _, m := range sourceURLRe.FindAllString(string(b), -1) {
		u, err := url.Parse(m)
		if err != nil || u.Host == "" {
			continue
		}
		if !seen[u.Host] {
			seen[u.Host] = true
			hosts = append(hosts, u.Host)
		}
	}
	return hosts
}

var pkgNameRe = regexp.MustCompile(`^[a-zA-Z0-9@._+-]+$`)

func validPkgName(s string) bool {
	return s != "" && len(s) < 128 && pkgNameRe.MatchString(s) && !strings.Contains(s, "..")
}
