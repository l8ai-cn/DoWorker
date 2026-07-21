package coordinator

import (
	"context"
	"strings"
	"testing"

	coordinatordom "github.com/l8ai-cn/agentcloud/backend/internal/domain/coordinator"
)

func TestRunProjectBlocksDispatchWithoutWorkerSpecSnapshot(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	dispatch := &fakeDispatch{}
	svc := NewService(Deps{
		Store:         newFakeStore(),
		Tickets:       &fakeTickets{},
		Dispatch:      dispatch,
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, CreatedByID: 2, RepositoryID: 3,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 1,
	}

	result, err := svc.RunProject(context.Background(), project)

	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if dispatch.n != 0 {
		t.Fatalf("dispatches = %d, want 0", dispatch.n)
	}
	if len(result.Errors) != 1 ||
		!strings.Contains(result.Errors[0], ErrCoordinatorWorkerSpecSnapshotRequired.Error()) {
		t.Fatalf("errors = %v, want snapshot required", result.Errors)
	}
}
