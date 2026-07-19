package workercreation

func (fixture *workspaceFixture) deps() workspaceResolverDeps {
	return workspaceResolverDeps{
		Repositories: fixture.repositories,
		Skills:       fixture.skills,
		Knowledge:    fixture.knowledge,
		EnvBundles:   fixture.envBundles,
		Definitions:  workspaceTestDefinitions(),
		Commits:      fixture.commits,
	}
}

func workspaceTestDefinitions() staticWorkerDefinitions {
	source := "AGENT codex\nEXECUTABLE codex\nMODE pty\nMODE acp\nENV OPENAI_API_KEY SECRET OPTIONAL\nENV SIGNING_KEY SECRET OPTIONAL\n"
	return staticWorkerDefinitions{
		"codex-cli": workerDefinition(
			"codex-cli", "codex", source, "pty", "acp",
		),
	}
}
