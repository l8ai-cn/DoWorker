package operatorcatalog

import (
	"context"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"gorm.io/gorm"
)

type bootstrapDependencyArtifactStore struct {
	createCalls int
	rows        map[int64]workerdependency.Document
}

func (store *bootstrapDependencyArtifactStore) Create(
	_ context.Context,
	_ int64,
	snapshotID int64,
	_ []byte,
	_ string,
) error {
	store.createCalls++
	if store.rows == nil {
		store.rows = map[int64]workerdependency.Document{}
	}
	store.rows[snapshotID] = bootstrapArtifactDocument(currentArtifactAgentfile())
	return nil
}

func (store *bootstrapDependencyArtifactStore) Delete(
	_ context.Context,
	_ int64,
	snapshotID int64,
) error {
	delete(store.rows, snapshotID)
	return nil
}

func (store *bootstrapDependencyArtifactStore) GetBySnapshotID(
	_ context.Context,
	_ int64,
	snapshotID int64,
) (workerdependency.Document, error) {
	row, ok := store.rows[snapshotID]
	if !ok {
		return workerdependency.Document{}, gorm.ErrRecordNotFound
	}
	return row, nil
}

func bootstrapArtifactDocument(agentfileSource string) workerdependency.Document {
	return workerdependency.Document{
		Version:        workerdependency.VersionV1,
		OrganizationID: 7,
		Namespace:      slugkit.MustNewForTest("dev-org"),
		Worker: workerdependency.Worker{
			WorkerType:      slugkit.MustNewForTest("video-studio"),
			AdapterID:       slugkit.MustNewForTest("codex-app-server"),
			SpecVersion:     1,
			SpecDigest:      "sha256:" + strings.Repeat("a", 64),
			DefinitionHash:  strings.Repeat("b", 64),
			AgentfileSource: agentfileSource,
		},
	}
}

func currentArtifactAgentfile() string {
	return `AGENT "video-studio-codex"
EXECUTABLE "video-studio-codex"
MODE pty
PROMPT_POSITION append
file sandbox.work_dir + "/AGENTS.md" "Produce video."
`
}

func legacyArtifactAgentfile() string {
	return `AGENT "video-studio-codex"
EXECUTABLE "video-studio-codex"
MODE pty
PROMPT_POSITION append
PROMPT "Produce video."
`
}
