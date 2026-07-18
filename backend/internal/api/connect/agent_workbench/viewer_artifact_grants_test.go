package agentworkbenchconnect

import (
	"context"
	"testing"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestViewerArtifactGrantsIntersectImageManifestWithCapabilities(t *testing.T) {
	actions := decoratedArtifactActions(
		t,
		ownerContext(),
		[]string{"artifact.download", "presentation.regenerate_slide"},
		testImageEditArtifact(),
	)

	require.Equal(t, []string{"artifact.download"}, actions)
}

func TestViewerArtifactGrantsIntersectPresentationManifestWithCapabilities(t *testing.T) {
	actions := decoratedArtifactActions(
		t,
		ownerContext(),
		[]string{
			"image.edit",
			"presentation.regenerate_slide",
			"presentation.select_version",
		},
		testPresentationArtifact(),
	)

	require.ElementsMatch(t, []string{
		"presentation.regenerate_slide",
		"presentation.select_version",
	}, actions)
}

func TestViewerArtifactGrantsRequireEmbedWriteInAdditionToCapability(t *testing.T) {
	capabilities := []string{"artifact.download", "image.edit"}
	tests := []struct {
		name     string
		ctx      context.Context
		expected []string
	}{
		{"owner", ownerContext(), capabilities},
		{"embed read only", embedContext("read"), []string{"artifact.download"}},
		{"embed write", embedContext("read", "write"), capabilities},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actions := decoratedArtifactActions(
				t,
				test.ctx,
				capabilities,
				testImageEditArtifact(),
			)

			require.ElementsMatch(t, test.expected, actions)
		})
	}
}

func TestViewerArtifactDeltaUsesChangedCapabilitiesForLaterArtifact(t *testing.T) {
	authorization, err := viewerAuthorizationFor(
		ownerContext(),
		activeSession(testUserID, 7),
	)
	require.NoError(t, err)
	require.NoError(t, authorization.decorateSnapshot(&agentworkbenchv2.SessionSnapshot{
		SessionId: testSessionID,
		Capabilities: &agentworkbenchv2.SupportCapabilities{
			ArtifactOperations: []string{"artifact.download"},
		},
	}))
	delta, err := authorization.decorateDelta(&agentworkbenchv2.SessionDeltaBatch{
		SessionId: testSessionID,
		Events: []*agentworkbenchv2.AgentEvent{
			{
				Event: &agentworkbenchv2.AgentEvent_CapabilitiesChanged{
					CapabilitiesChanged: &agentworkbenchv2.CapabilitiesChanged{
						Capabilities: &agentworkbenchv2.SupportCapabilities{
							ArtifactOperations: []string{"artifact.download", "image.edit"},
						},
					},
				},
			},
			{
				Event: &agentworkbenchv2.AgentEvent_ArtifactChanged{
					ArtifactChanged: &agentworkbenchv2.ArtifactChanged{
						Artifact: testImageEditArtifact(),
					},
				},
			},
		},
	})

	require.NoError(t, err)
	artifact := delta.GetEvents()[1].GetArtifactChanged().GetArtifact()
	require.ElementsMatch(t, []string{"artifact.download", "image.edit"},
		artifact.GetGrants()[0].GetActions())
}

func decoratedArtifactActions(
	t *testing.T,
	ctx context.Context,
	artifactOperations []string,
	artifact *agentworkbenchv2.ArtifactDescriptor,
) []string {
	t.Helper()
	authorization, err := viewerAuthorizationFor(
		ctx,
		activeSession(testUserID, 7),
	)
	require.NoError(t, err)
	snapshot := &agentworkbenchv2.SessionSnapshot{
		SessionId: testSessionID,
		Capabilities: &agentworkbenchv2.SupportCapabilities{
			ArtifactOperations: artifactOperations,
		},
		Artifacts: []*agentworkbenchv2.ArtifactDescriptor{artifact},
	}
	require.NoError(t, authorization.decorateSnapshot(snapshot))
	require.Len(t, snapshot.GetArtifacts()[0].GetGrants(), 1)
	return snapshot.GetArtifacts()[0].GetGrants()[0].GetActions()
}

func testPresentationArtifact() *agentworkbenchv2.ArtifactDescriptor {
	return &agentworkbenchv2.ArtifactDescriptor{
		ArtifactId: "presentation-1",
		Revision:   3,
		Filename:   "deck.pptx",
		MediaType:  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Status:     agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		Manifest: &agentworkbenchv2.ArtifactManifest{
			Manifest: &agentworkbenchv2.ArtifactManifest_Presentation{
				Presentation: &agentworkbenchv2.PresentationManifest{
					DeckRevision: 3,
				},
			},
		},
	}
}
