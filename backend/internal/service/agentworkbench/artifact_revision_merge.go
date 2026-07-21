package agentworkbench

import (
	"sort"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func mergeArtifactRevisions(
	current []*agentworkbenchv2.ArtifactRevision,
	next []*agentworkbenchv2.ArtifactRevision,
) []*agentworkbenchv2.ArtifactRevision {
	byRevision := make(
		map[uint64]*agentworkbenchv2.ArtifactRevision,
		len(current)+len(next),
	)
	for _, revision := range current {
		if revision != nil {
			byRevision[revision.GetRevision()] =
				proto.Clone(revision).(*agentworkbenchv2.ArtifactRevision)
		}
	}
	for _, revision := range next {
		if revision != nil {
			byRevision[revision.GetRevision()] =
				proto.Clone(revision).(*agentworkbenchv2.ArtifactRevision)
		}
	}
	revisions := make([]uint64, 0, len(byRevision))
	for revision := range byRevision {
		revisions = append(revisions, revision)
	}
	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i] < revisions[j]
	})
	merged := make(
		[]*agentworkbenchv2.ArtifactRevision,
		0,
		len(revisions),
	)
	for _, revision := range revisions {
		merged = append(merged, byRevision[revision])
	}
	return merged
}
