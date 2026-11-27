package tool

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// requireCleanMaster refuses to build or publish from anything but the pushed
// tip of master.
//
// Without it a run builds (and publishes) PKGBUILDs that an earlier item left
// in the working tree, so the served repo gets ahead of master and stays there
// while the bump PR sits open: the 2026-11-20 run published genshin-dict 7.2.0
// and eleos 0.8.7 without them ever being merged, and master was still on
// 7.0.1 / 0.8.6 days later with #833 / #839 open.
func requireCleanMaster(ctx context.Context, env *Env, toolName string) error {
	if res := run(ctx, env.RepoRoot, execTimeout(env), "git", "status", "--porcelain"); res.RC == 0 {
		if dirty := strings.TrimSpace(res.Stdout); dirty != "" {
			return &HarnessGap{Tool: toolName, Detail: fmt.Sprintf(
				"working tree is not clean, refusing to build/publish (%s). Only committed master content may be built: commit the edits (step_bump_pr) or park them.",
				firstLines(dirty, 5))}
		}
	}
	if res := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "--abbrev-ref", "HEAD"); res.RC == 0 {
		if b := strings.TrimSpace(res.Stdout); b != "master" {
			return &HarnessGap{Tool: toolName, Detail: fmt.Sprintf(
				"HEAD is on %q, not master; build/publish only from master", b)}
		}
	}
	run(ctx, env.RepoRoot, 2*time.Minute, "git", "fetch", "--quiet", "origin", "master")
	head := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "HEAD")
	origin := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "origin/master")
	if head.RC != 0 || origin.RC != 0 {
		return &HarnessGap{Tool: toolName, Detail: "cannot compare HEAD with origin/master (fetch failed?)"}
	}
	h, o := strings.TrimSpace(head.Stdout), strings.TrimSpace(origin.Stdout)
	if h != o {
		return &HarnessGap{Tool: toolName, Detail: fmt.Sprintf(
			"HEAD (%s) != origin/master (%s); build/publish only from the pushed master tip", shortSha(h), shortSha(o))}
	}
	return nil
}

// RestoreMaster returns HEAD to master when a phase left it on a PR branch.
//
// Triage checks a deps/* head out to review or repair the bump; when the phase
// finishes without merging, nothing switches the tree back and everyone after
// it sees a PR branch. build_plan then refuses (it only builds committed
// master), which used to end the whole run before any package was built (the
// deps PR was left open and the merged one was never built). The driver calls
// this between phases so the run continues from master.
//
// It returns the branch it moved away from, or "" when HEAD was already
// master. A dirty tree is not touched: the edits belong to the model, so the
// error is reported and requireCleanMaster explains the gap to the phase.
func RestoreMaster(ctx context.Context, env *Env) (string, error) {
	res := run(ctx, env.RepoRoot, execTimeout(env), "git", "rev-parse", "--abbrev-ref", "HEAD")
	if res.RC != 0 {
		return "", fmt.Errorf("cannot read HEAD: %s", errText(res))
	}
	branch := strings.TrimSpace(res.Stdout)
	if branch == "" || branch == "master" {
		return "", nil
	}
	if st := run(ctx, env.RepoRoot, execTimeout(env), "git", "status", "--porcelain"); st.RC == 0 {
		if strings.TrimSpace(st.Stdout) != "" {
			return branch, fmt.Errorf("HEAD is on %q and the working tree is not clean; not switching to master", branch)
		}
	}
	if co := run(ctx, env.RepoRoot, execTimeout(env), "git", "checkout", "--quiet", "master"); co.RC != 0 {
		return branch, fmt.Errorf("git checkout master from %q failed: %s", branch, errText(co))
	}
	return branch, nil
}

// firstLines joins the first n lines of a porcelain status with commas.
func firstLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = append(lines[:n], fmt.Sprintf("... (%d more)", len(lines)-n))
	}
	return strings.Join(lines, ", ")
}

func shortSha(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
