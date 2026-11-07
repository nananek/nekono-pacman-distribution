package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

// FinishTool is how a phase declares it is done. The agent loop treats a turn
// without a finish call as incomplete: the status is machine-checked, so the
// run report never depends on prose.
func FinishTool(env *Env) *Func {
	return &Func{
		Name:   "finish",
		Desc:   "Declare the phase finished. status=ok when the phase did what it should; needs_human when a person must look at the result; blocked when a pre-approved boundary was hit; failed when the phase could not do its job.",
		Params: obj(`"status":{"type":"string","enum":["ok","needs_human","blocked","failed"]},"summary":{"type":"string","description":"one paragraph, Japanese ok"},"findings":{"type":"array","items":{"type":"object","properties":{"severity":{"type":"string"},"pkg":{"type":"string"},"file":{"type":"string"},"note":{"type":"string"}},"required":["note"]}},"artifacts":{"type":"array","items":{"type":"string"},"description":"PR/Issue numbers, pkg-ver strings, file paths"}`, "status", "summary"),
		Run: func(_ context.Context, raw json.RawMessage) (Result, error) {
			var a PhaseResult
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, fmt.Errorf("finish: %w", err)
			}
			switch a.Status {
			case "ok", "needs_human", "blocked", "failed":
			default:
				return Result{}, fmt.Errorf("finish: invalid status %q", a.Status)
			}
			env.Finished = &a
			env.J.Emit("verdict", map[string]any{
				"status": a.Status, "summary": a.Summary,
				"findings": len(a.Findings), "artifacts": a.Artifacts,
			})
			env.J.Human("[finish] %s: %s", a.Status, a.Summary)
			return Result{Summary: "phase finished: " + a.Status, Data: a}, nil
		},
	}
}
