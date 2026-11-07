package tool

// All is the full tool catalog. A phase only sees the subset named in
// agent/agent.json; this list is what the registry is built from.
func All(env *Env) []*Func {
	return []*Func{
		PreflightTool(env),
		RepoStateTool(env),
		RepoDiffTool(env),
		RepoLogTool(env),
		RepoMergeBaseTool(env),
		RepoPullTool(env),
		FSReadTool(env),
		FSWriteTool(env),
		SigningProbeTool(env),
		SigningVerifyTool(env),
		DiskFreeTool(env),
		EscalateTool(env),
		FinishTool(env),
		BuildPlanTool(env),
		BuildRunTool(env),
		BuildVerifyTool(env),
		DistPublishTool(env),
		DistVerifyTool(env),
		GHPRTool(env),
		GHIssueTool(env),
		GHRepoTool(env),
		FetchTool(env),
		ArchiveListTool(env),
		ArchiveExtractTool(env),
		ArchiveDiffTool(env),
		BinaryScanTool(env),
		PKGSrcinfoTool(env),
		PKGVerifysourceTool(env),
		PKGThreePointsTool(env),
		StepBotPRTool(env),
		StepMergeTool(env),
		StepBumpPRTool(env),
	}
}
