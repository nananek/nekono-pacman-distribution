package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// EscalateTool is the single exit for "a human has to decide this".
//
// Two kinds, deliberately different:
//
//   - content: the world (upstream / package / supply chain) needs a human.
//     A repository Issue is created once per fingerprint and the phase stops
//     for that item. The repo tracker stays about packages.
//   - harness: the closed tool set could not express what the phase needed.
//     No Issue (the tracker is not a bug tracker for this agent); the run
//     stops and the report/Discord carries it.
func EscalateTool(env *Env) *Func {
	return &Func{
		Name:    "escalate",
		Desc:    "Hand a decision to a human. kind=content creates (or reuses) a repository Issue once per fingerprint, then stops the item. kind=harness records a capability gap in the run journal and stops the whole run.",
		Params:  obj(`"kind":{"type":"string","enum":["content","harness"]},"phase":{"type":"string"},"pkg":{"type":"string"},"subject":{"type":"string","description":"one-line subject, e.g. 'checksum mismatch'"},"detail":{"type":"string"},"evidence":{"type":"array","items":{"type":"string"},"description":"URLs / sha256 / commands that support the claim"},"source_issue":{"type":"integer","description":"existing Issue this escalation belongs to (optional)"}`, "kind", "subject"),
		Mutates: true,
		Run: func(ctx context.Context, raw json.RawMessage) (Result, error) {
			var a struct {
				Kind        string   `json:"kind"`
				Phase       string   `json:"phase"`
				Pkg         string   `json:"pkg"`
				Subject     string   `json:"subject"`
				Detail      string   `json:"detail"`
				Evidence    []string `json:"evidence"`
				SourceIssue int      `json:"source_issue"`
			}
			if err := json.Unmarshal(raw, &a); err != nil {
				return Result{}, err
			}
			fp := fingerprint(a.Phase, a.Pkg, a.Subject)

			if a.Kind == "harness" {
				env.J.Emit("capability_gap", map[string]any{
					"phase": a.Phase, "subject": a.Subject, "detail": a.Detail,
					"fingerprint": fp, "evidence": a.Evidence,
				})
				env.Escalated = &Escalation{Kind: "harness", Phase: a.Phase, Pkg: a.Pkg, Reason: a.Detail}
				return Result{}, &HarnessGap{Tool: "escalate", Detail: a.Subject + ": " + a.Detail}
			}
			if a.Kind != "content" {
				return Result{}, &HarnessGap{Tool: "escalate", Detail: "kind must be content or harness"}
			}
			if env.Escalated == nil {
				env.Escalated = &Escalation{Kind: "content", Phase: a.Phase, Pkg: a.Pkg, Reason: a.Subject}
			}

			title := fmt.Sprintf("[agent:%s] %s", a.Phase, a.Subject)
			if a.Pkg != "" {
				title = fmt.Sprintf("[agent:%s] %s: %s", a.Phase, a.Pkg, a.Subject)
			}

			// Idempotency: one Issue per fingerprint, ever (open or closed).
			existing := 0
			list := run(ctx, env.RepoRoot, execTimeout(env),
				"gh", "issue", "list", "--state", "all", "--limit", "200",
				"--search", fmt.Sprintf("in:body %q", fp),
				"--json", "number,title")
			if list.RC == 0 && strings.TrimSpace(list.Stdout) != "" {
				var issues []struct {
					Number int    `json:"number"`
					Title  string `json:"title"`
				}
				if err := json.Unmarshal([]byte(list.Stdout), &issues); err == nil && len(issues) > 0 {
					existing = issues[0].Number
				}
			}
			if existing > 0 {
				env.Escalated.Issue = existing
				env.J.Emit("escalate", map[string]any{"kind": "content", "pkg": a.Pkg, "subject": a.Subject, "issue": existing, "deduplicated": true})
				return Result{
					Summary: fmt.Sprintf("already escalated: Issue #%d (not duplicated)", existing),
					Data:    map[string]any{"issue": existing, "deduplicated": true},
				}, nil
			}

			body := escalationBody(a.Phase, a.Pkg, a.Subject, a.Detail, a.Evidence, fp)
			created := run(ctx, env.RepoRoot, execTimeout(env),
				"gh", "issue", "create", "--title", title, "--body", body)
			if created.RC != 0 {
				return Result{}, fmt.Errorf("gh issue create failed: %s", errText(created))
			}
			env.J.Emit("escalate", map[string]any{"kind": "content", "pkg": a.Pkg, "subject": a.Subject, "created": true, "url": lastURL(created.Stdout)})

			if a.SourceIssue > 0 {
				comment := fmt.Sprintf("保留 (agent escalated): %s\n\n%s", a.Subject, a.Detail)
				if c := run(ctx, env.RepoRoot, execTimeout(env), "gh", "issue", "comment", fmt.Sprint(a.SourceIssue), "--body", comment); c.RC != 0 {
					env.J.Emit("warn", map[string]any{"msg": "gh issue comment failed", "issue": a.SourceIssue, "stderr": errText(c)})
				}
			}
			return Result{
				Summary: "escalated to a human: " + title,
				Data:    map[string]any{"created": true, "title": title, "url": lastURL(created.Stdout)},
			}, nil
		},
	}
}

func fingerprint(phase, pkg, subject string) string {
	sum := sha256.Sum256([]byte(phase + "|" + pkg + "|" + strings.ToLower(strings.TrimSpace(subject))))
	return hex.EncodeToString(sum[:8])
}

func escalationBody(phase, pkg, subject, detail string, evidence []string, fp string) string {
	var b strings.Builder
	b.WriteString("## 状況\n")
	b.WriteString(fmt.Sprintf("- phase: `%s`\n", phase))
	if pkg != "" {
		b.WriteString(fmt.Sprintf("- pkg: `%s`\n", pkg))
	}
	b.WriteString(fmt.Sprintf("- subject: %s\n", subject))
	b.WriteString("\n## 何が判断できなかったか\n")
	b.WriteString(detail)
	b.WriteString("\n")
	if len(evidence) > 0 {
		b.WriteString("\n## 確認したこと (証跡)\n")
		for _, e := range evidence {
			b.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}
	b.WriteString("\n## 再開に必要なこと\n")
	b.WriteString("人間の判断 (verdict) をこの Issue に書いてください。agent は次回の run で同じ fingerprint を再利用し、重複 Issue は作りません。\n")
	b.WriteString(fmt.Sprintf("\n<!-- nekono-agent-fingerprint: %s -->\n", fp))
	return b.String()
}

// lastURL extracts the trailing URL printed by `gh issue create`.
func lastURL(s string) string {
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if strings.HasPrefix(line, "https://") {
			return line
		}
	}
	return ""
}
