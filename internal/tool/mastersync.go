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
