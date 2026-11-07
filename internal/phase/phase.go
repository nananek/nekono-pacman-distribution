// Package phase runs one phase: a prompt, a closed tool set, a step budget.
//
// The isolation the old OpenCode harness got from subagents + permission
// patterns is a process fact here: the registry only contains the tools the
// phase declares, so a phase physically cannot call another phase's tools.
package phase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nananek/nekono-pacman-distribution/internal/journal"
	"github.com/nananek/nekono-pacman-distribution/internal/llm"
	"github.com/nananek/nekono-pacman-distribution/internal/tool"
)

// Def is one phase definition.
type Def struct {
	Name     string
	Tools    []string
	MaxSteps int
	Prompt   string
	Task     string
}

// Outcome is the typed result of running a phase.
type Outcome struct {
	Phase     string         `json:"phase"`
	Status    string         `json:"status"`
	Summary   string         `json:"summary"`
	Findings  []tool.Finding `json:"findings,omitempty"`
	Artifacts []string       `json:"artifacts,omitempty"`
	Escalated string         `json:"escalated,omitempty"`
	Steps     int            `json:"steps"`
	Usage     llm.Usage      `json:"usage"`
}

// Runner executes one phase.
type Runner struct {
	Def    Def
	Reg    *tool.Registry
	Env    *tool.Env
	LLM    *llm.Client
	J      *journal.Journal
	Rules  string
	Prompt string
}

// Run drives the model/tool loop until finish, a capability gap, or the step
// budget.
func (r *Runner) Run(ctx context.Context) Outcome {
	specs := r.Reg.Specs(r.Def.Tools)
	system := strings.TrimSpace(r.Rules) + "\n\n" + strings.TrimSpace(r.Prompt)
	msgs := []llm.Msg{
		{Role: "system", Content: system},
		{Role: "user", Content: r.Def.Task},
	}

	out := Outcome{Phase: r.Def.Name}
	nudged := false

	r.LLM.OnDelta = func(d string) { r.J.Emit("llm_delta", map[string]any{"phase": r.Def.Name, "text": d}) }

	for step := 1; step <= r.Def.MaxSteps; step++ {
		out.Steps = step
		r.J.Emit("llm_call", map[string]any{"phase": r.Def.Name, "step": step, "messages": len(msgs)})

		comp, err := r.LLM.Complete(ctx, msgs, specs)
		if err != nil {
			out.Status = "failed"
			out.Summary = "LLM call failed: " + err.Error()
			r.J.Emit("llm_error", map[string]any{"phase": r.Def.Name, "error": err.Error()})
			return out
		}
		out.Usage.PromptTokens += comp.Usage.PromptTokens
		out.Usage.CompletionTokens += comp.Usage.CompletionTokens
		out.Usage.TotalTokens += comp.Usage.TotalTokens
		if comp.Content != "" {
			r.J.Emit("assistant", map[string]any{"phase": r.Def.Name, "text": truncate(comp.Content, 4000)})
		}

		if len(comp.ToolCalls) == 0 {
			if r.Env.Finished != nil {
				break
			}
			if nudged {
				out.Status = "failed"
				out.Summary = "model stopped without calling finish"
				return out
			}
			msgs = append(msgs,
				llm.Msg{Role: "assistant", Content: comp.Content},
				llm.Msg{Role: "user", Content: "ツールを呼ばずに終わろうとしています。この phase の結論を finish ツールで宣言してください。"})
			nudged = true
			continue
		}

		msgs = append(msgs, llm.Msg{Role: "assistant", Content: comp.Content, ToolCalls: comp.ToolCalls})

		stop := false
		for _, tc := range comp.ToolCalls {
			name := tc.Function.Name
			raw := json.RawMessage(tc.Function.Arguments)
			if len(raw) == 0 {
				raw = json.RawMessage("{}")
			}
			if !json.Valid(raw) {
				msgs = append(msgs, toolMessage(tc.ID, "ERROR: arguments are not valid JSON"))
				r.J.Emit("tool_error", map[string]any{"phase": r.Def.Name, "tool": name, "error": "invalid json arguments"})
				continue
			}
			r.J.Emit("tool_call", map[string]any{"phase": r.Def.Name, "tool": name, "args": truncate(string(raw), 2000)})

			res, err := r.Reg.Call(ctx, name, raw)
			if err != nil {
				if gap, ok := err.(*tool.HarnessGap); ok {
					r.J.Emit("capability_gap", map[string]any{"phase": r.Def.Name, "tool": gap.Tool, "detail": gap.Detail})
					out.Status = "harness_gap"
					out.Summary = gap.Error()
					out.Escalated = "harness"
					return out
				}
				r.J.Emit("tool_error", map[string]any{"phase": r.Def.Name, "tool": name, "error": err.Error()})
				msgs = append(msgs, toolMessage(tc.ID, "ERROR: "+err.Error()))
				continue
			}
			r.J.Emit("tool_result", map[string]any{"phase": r.Def.Name, "tool": name, "summary": res.Summary})
			msgs = append(msgs, toolMessage(tc.ID, resultMessage(res)))
			if r.Env.Finished != nil {
				stop = true
				break
			}
		}
		if stop {
			break
		}
	}

	if r.Env.Finished == nil {
		out.Status = "failed"
		out.Summary = fmt.Sprintf("step budget exhausted (%d steps) without finish", r.Def.MaxSteps)
		return out
	}
	out.Status = r.Env.Finished.Status
	out.Summary = r.Env.Finished.Summary
	out.Findings = r.Env.Finished.Findings
	out.Artifacts = r.Env.Finished.Artifacts
	if r.Env.Escalated != nil {
		out.Escalated = r.Env.Escalated.Kind
	}
	return out
}

func toolMessage(id, content string) llm.Msg {
	return llm.Msg{Role: "tool", ToolCallID: id, Content: content}
}

func resultMessage(res tool.Result) string {
	b, err := json.Marshal(res.Data)
	if err != nil {
		return res.Summary
	}
	return res.Summary + "\n\n" + truncate(string(b), 20000)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("\n... [truncated %d bytes]", len(s)-max)
}
