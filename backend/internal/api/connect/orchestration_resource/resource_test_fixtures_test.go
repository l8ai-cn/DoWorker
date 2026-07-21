package orchestrationresourceconnect

import (
	"encoding/json"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

const (
	testPlanID     = "11111111-1111-4111-8111-111111111111"
	testResourceID = "22222222-2222-4222-8222-222222222222"
	testRefID      = "33333333-3333-4333-8333-333333333333"
)

var testTime = time.Date(
	2026,
	7,
	14,
	8,
	30,
	0,
	0,
	time.FixedZone("test", 8*60*60),
)

func testScope() control.Scope {
	return control.Scope{
		OrganizationID:   81,
		OrganizationSlug: slugkit.Slug("acme"),
		ActorID:          42,
	}
}

func testTarget() control.ResourceTarget {
	return control.ResourceTarget{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		},
		Namespace: slugkit.Slug("acme"),
		Name:      slugkit.Slug("builder"),
	}
}

func testHead() control.ResourceHead {
	return control.ResourceHead{
		ID:             9,
		OrganizationID: 81,
		Identity: control.ResourceIdentity{
			ResourceTarget: testTarget(),
			UID:            testResourceID,
		},
		DisplayName:     "Builder",
		Labels:          map[string]string{"team": "platform"},
		Status:          json.RawMessage(`{"phase":"ready"}`),
		Revision:        3,
		Generation:      2,
		ResourceVersion: 7,
		CreatedByID:     40,
		UpdatedByID:     42,
		CreatedAt:       testTime,
		UpdatedAt:       testTime.Add(time.Hour),
	}
}

func testPlan() control.Plan {
	return control.Plan{
		ID:                  testPlanID,
		Scope:               testScope(),
		ActorID:             42,
		Operation:           control.PlanOperationUpdate,
		Target:              testTarget(),
		TargetResourceID:    9,
		BaseUID:             testResourceID,
		BaseResourceVersion: 7,
		DraftHash:           "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PlanHash:            "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CanonicalManifest:   json.RawMessage(`{"private":"canonical-payload"}`),
		ResolvedReferences: []control.ResolvedReference{{
			TypeMeta: resource.TypeMeta{
				APIVersion: resource.APIVersionV1Alpha1,
				Kind:       "Prompt",
			},
			Namespace: slugkit.Slug("acme"),
			Name:      slugkit.Slug("review-prompt"),
			UID:       testRefID,
			Revision:  4,
			Digest:    "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		}},
		SemanticChanges: []control.SemanticChange{{
			Operation: control.SemanticChangeReplace,
			Path:      "/spec/model",
			Before: control.ChangeValue{
				Digest: "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			},
			After: control.ChangeValue{
				RedactedJSON: json.RawMessage(`{"value":"redacted"}`),
			},
		}},
		Issues: []control.PlanIssue{{
			Severity: control.PlanIssueWarning,
			Path:     "/spec/model",
			Code:     "model.changed",
			Message:  "The model changed.",
		}},
		ArtifactKind:    "WorkerSpec",
		ArtifactJSON:    json.RawMessage(`{"private":"artifact-payload"}`),
		ArtifactDigest:  "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		OptionsRevision: "catalog-7",
		CreatedAt:       testTime,
		ExpiresAt:       testTime.Add(15 * time.Minute),
		Status:          control.PlanStatusPending,
	}
}
