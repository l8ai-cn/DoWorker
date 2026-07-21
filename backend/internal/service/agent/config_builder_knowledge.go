package agent

import (
	"fmt"
	"strings"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

// knowledgeContextFile renders kb/README.md — the agent-visible manifest of
// mounted knowledge bases and the llm-wiki working agreement (read llms.txt
// first, follow AGENTS.md, commit+push on rw mounts).
func knowledgeContextFile(mounts []*runnerv1.KnowledgeMount) *runnerv1.FileToCreate {
	if len(mounts) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("# Mounted Knowledge Bases\n\n")
	b.WriteString("The following knowledge bases are cloned as git repositories under this directory:\n\n")
	for _, m := range mounts {
		access := "read-only"
		if m.Mode == "rw" {
			access = "read-write"
		}
		fmt.Fprintf(&b, "- `%s/` (%s, branch `%s`)\n", m.Slug, access, m.Branch)
	}
	b.WriteString(`
## How to use them

1. Start with each KB's llms.txt — it is the navigation index (title, summary, section links).
2. AGENTS.md in each KB defines its wiki schema and maintenance workflow. Follow it when editing.
3. raw/ holds immutable source material; wiki/ is the LLM-maintained compiled layer.
4. For read-write mounts: after updating wiki pages, record the change in wiki/log.md,
   then commit and push (the remote already has push credentials configured).
5. Read-only mounts cannot be pushed; propose changes elsewhere instead.
`)

	return &runnerv1.FileToCreate{
		Path:    PlaceholderSandboxRoot + "/kb/README.md",
		Content: b.String(),
		Mode:    0644,
	}
}
