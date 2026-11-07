// Package tool holds the closed tool surface of nekono-pacman-agent.
//
// Design rules (they are the reason this package looks the way it does):
//
//   - There is no generic shell tool. Every capability is a named function with
//     a JSON schema, and the argv of the underlying program is assembled in Go.
//     The model can only pass typed arguments, so shell splitting, locale
//     effects, `| tail` exit-code loss and command-string allowlist gaps cannot
//     happen by construction.
//   - Tools return structured data, never log text. Human output from the
//     underlying bin/* entrypoints goes to stderr; the JSON document on stdout
//     (`--json`) is what this package parses.
//   - Anything the closed set cannot express is a *capability gap*: the phase
//     stops and the run reports it. It is never worked around with a new
//     ad-hoc command at run time.
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
	"github.com/nananek/nekono-pacman-distribution/internal/llm"
)

// Finding is one judgement result reported by a phase.
type Finding struct {
	Severity string `json:"severity,omitempty"`
	Pkg      string `json:"pkg,omitempty"`
	File     string `json:"file,omitempty"`
	Note     string `json:"note"`
}

// PhaseResult is the typed end of a phase (produced by the finish tool).
type PhaseResult struct {
	Status    string    `json:"status"`
	Summary   string    `json:"summary"`
	Findings  []Finding `json:"findings,omitempty"`
	Artifacts []string  `json:"artifacts,omitempty"`
}

// Escalation records that the phase handed a decision to a human.
type Escalation struct {
	Kind   string // content | harness
	Phase  string
	Pkg    string
	Issue  int // content: issue number created (0 = deduplicated/reused)
	Reason string
}

// Env is the shared state of one phase run.
type Env struct {
	RepoRoot  string
	Scratch   string
	Cfg       *cfg.Config
	J         *journal.Journal
	Escalated *Escalation
	Finished  *PhaseResult
}

// Result is what a tool hands back to the model.
type Result struct {
	// Summary is a one-line human-readable outcome (kept short; the model sees
	// it before the JSON payload).
	Summary string
	// Data is marshalled to JSON for the model.
	Data any
}

// HarnessGap means the closed tool set could not express what the phase
// needed. It stops the phase and the run; it never becomes a repo Issue.
type HarnessGap struct {
	Tool   string
	Detail string
}

func (e *HarnessGap) Error() string {
	return fmt.Sprintf("capability gap: %s: %s", e.Tool, e.Detail)
}

// Func is one callable tool.
type Func struct {
	Name    string
	Desc    string
	Params  json.RawMessage
	Mutates bool
	Run     func(ctx context.Context, args json.RawMessage) (Result, error)
}

// Registry is the set of tools available in a phase.
type Registry struct {
	funcs map[string]*Func
	names []string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{funcs: map[string]*Func{}}
}

// Add registers a tool.
func (r *Registry) Add(f *Func) {
	r.funcs[f.Name] = f
	r.names = append(r.names, f.Name)
}

// Names lists the registered tools in insertion order.
func (r *Registry) Names() []string { return r.names }

// Get looks a tool up by name.
func (r *Registry) Get(name string) (*Func, bool) {
	f, ok := r.funcs[name]
	return f, ok
}

// Specs returns the llm tool definitions for the given tool names.
func (r *Registry) Specs(names []string) []llm.Tool {
	out := make([]llm.Tool, 0, len(names))
	for _, n := range names {
		f, ok := r.funcs[n]
		if !ok {
			continue
		}
		out = append(out, llm.Tool{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        f.Name,
				Description: f.Desc,
				Parameters:  f.Params,
			},
		})
	}
	return out
}

// Call invokes a tool by name with raw JSON arguments.
func (r *Registry) Call(ctx context.Context, name string, args json.RawMessage) (Result, error) {
	f, ok := r.funcs[name]
	if !ok {
		return Result{}, &HarnessGap{Tool: name, Detail: "tool not registered in this phase"}
	}
	return f.Run(ctx, args)
}

// obj is a helper for building JSON schemas compactly.
func obj(props string, required ...string) json.RawMessage {
	req := "[]"
	if len(required) > 0 {
		b, _ := json.Marshal(required)
		req = string(b)
	}
	return json.RawMessage(`{"type":"object","properties":{` + props + `},"required":` + req + `,"additionalProperties":false}`)
}
