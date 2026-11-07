package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// agent/agent.json is the only place that decides which tools a phase can call.
// A typo there is a capability gap at run time (the run stops), so it is checked
// here: every listed tool must exist in the catalog and every prompt must exist.
func TestAgentConfigMatchesCatalog(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "agent", "agent.json"))
	if err != nil {
		t.Fatalf("read agent.json: %v", err)
	}
	var cfg struct {
		Phases map[string]struct {
			Tools  []string `json:"tools"`
			Prompt string   `json:"prompt"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatalf("parse agent.json: %v", err)
	}
	known := map[string]bool{}
	for _, f := range All(&Env{}) {
		known[f.Name] = true
	}
	for name, ph := range cfg.Phases {
		if len(ph.Tools) == 0 {
			t.Errorf("phase %s: no tools", name)
		}
		for _, toolName := range ph.Tools {
			if !known[toolName] {
				t.Errorf("phase %s: tool %q is not in the catalog", name, toolName)
			}
		}
		if ph.Prompt == "" {
			t.Errorf("phase %s: no prompt", name)
			continue
		}
		if _, err := os.Stat(filepath.Join(root, ph.Prompt)); err != nil {
			t.Errorf("phase %s: prompt %s: %v", name, ph.Prompt, err)
		}
	}
	// The phases the driver calls must all be defined.
	for _, want := range []string{"preflight", "triage", "upstream", "build", "publish"} {
		if _, ok := cfg.Phases[want]; !ok {
			t.Errorf("agent.json is missing the %q phase", want)
		}
	}
}
