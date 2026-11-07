// Package cfg loads agent/agent.json and hands out the model credentials.
package cfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Phase is one phase definition: which tools the model may call, how many
// steps it gets, and which prompt file to load.
type Phase struct {
	Tools    []string `json:"tools"`
	MaxSteps int      `json:"max_steps"`
	Prompt   string   `json:"prompt"`
}

// Config is agent/agent.json.
type Config struct {
	Model      string `json:"model"`
	BaseURL    string `json:"base_url"`
	APIKeyFile string `json:"api_key_file"`

	ExecTimeoutSec  int `json:"exec_timeout_sec"`
	BuildTimeoutSec int `json:"build_timeout_sec"`
	MaxSteps        int `json:"max_steps"`

	// DistServerPath is where the published repo is reachable anonymously
	// (pacman's Server line). Empty means "read /etc/pacman.conf".
	DistServerPath string `json:"dist_server_path"`

	Phases map[string]Phase `json:"phases"`

	// Path is where this config was loaded from (for error messages).
	Path string `json:"-"`
}

// Default returns the built-in defaults (used before a config file exists and
// as the base for missing fields).
func Default() *Config {
	return &Config{
		Model:           "opencode-go/deepseek-v4.1-flash",
		BaseURL:         "",
		APIKeyFile:      "~/.config/nekono-pacman/opencode.key",
		ExecTimeoutSec:  300,
		BuildTimeoutSec: 6 * 60 * 60,
		MaxSteps:        40,
		Phases:          map[string]Phase{},
	}
}

// Load reads path (JSON, with //-comment lines allowed) over the defaults.
func Load(path string) (*Config, error) {
	c := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if err := json.Unmarshal(stripComments(b), c); err != nil {
		return nil, fmt.Errorf("config %s: %w", path, err)
	}
	c.Path = path
	if c.BaseURL == "" {
		c.BaseURL = os.Getenv("NEKONO_AGENT_BASE_URL")
	}
	if v := os.Getenv("NEKONO_AGENT_MODEL"); v != "" {
		c.Model = v
	}
	if c.Phases == nil {
		c.Phases = map[string]Phase{}
	}
	return c, nil
}

// stripComments drops whole-line // comments (the repo's config style) so the
// file can be documented without a JSONC dependency.
func stripComments(b []byte) []byte {
	var out []byte
	for _, line := range bytes.Split(b, []byte("\n")) {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("//")) {
			continue
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out
}

// Key reads the provider API key. The key is only ever held in memory and sent
// to the configured endpoint; no tool can read this file.
func (c *Config) Key() (string, error) {
	p := expand(c.APIKeyFile)
	b, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("api key: %w", err)
	}
	k := strings.TrimSpace(string(b))
	if k == "" {
		return "", fmt.Errorf("api key: %s is empty", p)
	}
	return k, nil
}

// ResolvePath makes a possibly-relative path absolute against the repo root.
func (c *Config) ResolvePath(repoRoot, p string) string {
	p = expand(p)
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(repoRoot, p)
}

func expand(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
