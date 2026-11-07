package tool

import (
	"encoding/json"
	"io"
	"testing"
)

// Every tool parameter schema must be a valid JSON object with type=object.
// A malformed schema poisons the whole request (the provider rejects the body,
// not just that one tool), so this is a hard gate.
func TestToolSchemasAreValidJSON(t *testing.T) {
	env := &Env{}
	seen := map[string]bool{}
	for _, f := range All(env) {
		if f.Name == "" {
			t.Fatal("tool with empty name")
		}
		if seen[f.Name] {
			t.Fatalf("duplicate tool name %q", f.Name)
		}
		seen[f.Name] = true
		if len(f.Params) == 0 {
			t.Fatalf("%s: empty params", f.Name)
		}
		var v map[string]any
		dec := json.NewDecoder(bytesReader(f.Params))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("%s: params are not valid JSON: %v (%s)", f.Name, err, string(f.Params))
		}
		if v["type"] != "object" {
			t.Fatalf("%s: params type must be object, got %v", f.Name, v["type"])
		}
		if _, ok := v["properties"]; !ok {
			t.Fatalf("%s: params must have properties", f.Name)
		}
	}
	if len(seen) != 18 {
		t.Logf("catalog has %d tools", len(seen))
	}
}

func bytesReader(b []byte) io.Reader { return &sliceReader{b: b} }

type sliceReader struct {
	b []byte
	i int
}

func (r *sliceReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
