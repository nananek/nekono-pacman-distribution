// Command nekono-pacman-agent is the daily-ops harness: a closed tool set for
// the nekono-pacman-distribution runbook plus a deterministic driver.
//
// Why not OpenCode: an OpenCode run owns the terminal and only explains itself
// when it is over, and its permission model is best-effort matching on shell
// strings. Here the scheduling and all verification are Go code, the model is
// called only where a judgement is needed, and every step is an ndjson event
// on stdout that the host can tail.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/nananek/nekono-pacman-distribution/internal/cfg"
	"github.com/nananek/nekono-pacman-distribution/internal/journal"
	"github.com/nananek/nekono-pacman-distribution/internal/llm"
	"github.com/nananek/nekono-pacman-distribution/internal/phase"
	"github.com/nananek/nekono-pacman-distribution/internal/tool"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var err error
	switch os.Args[1] {
	case "doctor":
		err = cmdDoctor(ctx, os.Args[2:])
	case "phase":
		err = cmdPhase(ctx, os.Args[2:])
	case "run":
		err = cmdRun(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `nekono-pacman-agent — daily-ops harness for nekono-pacman-distribution

  doctor            check prerequisites (config, key, gh, signing, endpoint)
  phase <name>      run one phase (preflight | build | publish)
  run [flags]       deterministic daily run: preflight -> detect -> build -> publish

common flags: -config agent/agent.json  -repo .  -json (ndjson on stdout)
`)
}

// setup loads config, builds the journal and env, and returns everything a
// phase needs.
type app struct {
	cfg  *cfg.Config
	j    *journal.Journal
	env  *tool.Env
	llm  *llm.Client
	reg  *tool.Registry
	root string
}

func setup(ctx context.Context, args []string) (*app, *flag.FlagSet, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	cfgPath := fs.String("config", "agent/agent.json", "path to agent config")
	repo := fs.String("repo", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	root, err := filepath.Abs(*repo)
	if err != nil {
		return nil, nil, err
	}
	c, err := cfg.Load(filepath.Join(root, *cfgPath))
	if err != nil {
		return nil, nil, err
	}
	j := journal.New(os.Stdout, os.Stderr)
	scratch, err := os.MkdirTemp("", "nekono-agent-")
	if err != nil {
		return nil, nil, err
	}
	env := &tool.Env{RepoRoot: root, Scratch: scratch, Cfg: c, J: j}
	a := &app{cfg: c, j: j, env: env, root: root, reg: tool.NewRegistry()}
	for _, f := range tool.All(env) {
		a.reg.Add(f)
	}
	return a, fs, nil
}

// phaseDef builds a phase.Def from the config.
func (a *app) phaseDef(name string) (phase.Def, error) {
	pc, ok := a.cfg.Phases[name]
	if !ok {
		return phase.Def{}, fmt.Errorf("phase %q is not defined in %s", name, a.cfg.Path)
	}
	prompt := pc.Prompt
	if prompt == "" {
		prompt = "agent/phases/" + name + ".md"
	}
	promptPath := a.cfg.ResolvePath(a.root, prompt)
	pb, err := os.ReadFile(promptPath)
	if err != nil {
		return phase.Def{}, fmt.Errorf("phase %q: %w", name, err)
	}
	task := fmt.Sprintf("phase: %s\nrepository: %s\n\nphase prompt と利用可能なツールだけを使い、最後に必ず finish ツールで結論を宣言してください。", name, a.root)
	return phase.Def{
		Name:     name,
		Tools:    pc.Tools,
		MaxSteps: pc.MaxSteps,
		Prompt:   string(pb),
		Task:     task,
	}, nil
}

func (a *app) rules() (string, error) {
	b, err := os.ReadFile(a.cfg.ResolvePath(a.root, "agent/rules.md"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// runPhase executes one phase and journals its outcome.
func (a *app) runPhase(ctx context.Context, name string) (phase.Outcome, error) {
	def, err := a.phaseDef(name)
	if err != nil {
		return phase.Outcome{}, err
	}
	rules, err := a.rules()
	if err != nil {
		return phase.Outcome{}, err
	}
	if def.MaxSteps <= 0 {
		def.MaxSteps = a.cfg.MaxSteps
	}
	if a.llm == nil {
		key, err := a.cfg.Key()
		if err != nil {
			return phase.Outcome{}, err
		}
		a.llm = llm.New(a.cfg.BaseURL, key, a.cfg.Model)
		a.llm.SessionID = sessionID()
	}
	a.j.Emit("phase_start", map[string]any{"phase": name, "tools": def.Tools, "max_steps": def.MaxSteps})
	a.j.Human("\n== phase %s (tools: %s)", name, strings.Join(def.Tools, ","))

	r := &phase.Runner{Def: def, Reg: a.reg, Env: a.env, LLM: a.llm, J: a.j, Rules: rules}
	out := r.Run(ctx)
	a.j.Emit("phase_end", map[string]any{
		"phase": name, "status": out.Status, "summary": out.Summary,
		"steps": out.Steps, "tokens": out.Usage.TotalTokens, "escalated": out.Escalated,
	})
	return out, nil
}

func cmdPhase(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: nekono-pacman-agent phase <name> [flags]")
	}
	name := args[0]
	a, _, err := setup(ctx, args[1:])
	if err != nil {
		return err
	}
	out, err := a.runPhase(ctx, name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		return err
	}
	return exitFor(out)
}

func exitFor(out phase.Outcome) error {
	switch out.Status {
	case "ok":
		return nil
	case "harness_gap":
		return fmt.Errorf("capability gap: %s", out.Summary)
	default:
		return fmt.Errorf("phase %s: %s: %s", out.Phase, out.Status, out.Summary)
	}
}

// sessionID is a per-process identifier for the ocgo endpoint (it routes per
// session). It is not a secret and carries no state.
func sessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("nekono-agent-%d", os.Getpid())
	}
	return "nekono-agent-" + hex.EncodeToString(b[:])
}
