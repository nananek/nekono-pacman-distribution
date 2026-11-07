package tool

import (
	"encoding/json"
	"testing"

	"github.com/nananek/nekono-pacman-distribution/internal/llm"
)

// The provider receives one request containing every tool schema; a single
// malformed schema makes the whole call fail. This reproduces the exact
// marshalling the client does.
func TestSpecsMarshal(t *testing.T) {
	env := &Env{}
	reg := NewRegistry()
	for _, f := range All(env) {
		reg.Add(f)
	}
	specs := reg.Specs(reg.Names())
	if len(specs) == 0 {
		t.Fatal("no specs")
	}
	b, err := json.Marshal(specs)
	if err != nil {
		t.Fatalf("marshal specs: %v", err)
	}
	var back []llm.Tool
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal specs: %v", err)
	}
	if len(back) != len(specs) {
		t.Fatalf("roundtrip lost tools: %d != %d", len(back), len(specs))
	}
}
