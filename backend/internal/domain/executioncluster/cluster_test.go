package executioncluster

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestClusterValidateIdentifiersRejectsInvalidSlug(t *testing.T) {
	cluster := &Cluster{Slug: slugkit.Slug("Local_Cluster")}

	require.Error(t, cluster.ValidateIdentifiers())
}

func TestClusterValidateIdentifiersAcceptsValidSlug(t *testing.T) {
	cluster := &Cluster{Slug: slugkit.Slug("local")}

	require.NoError(t, cluster.ValidateIdentifiers())
}

func TestClusterValidateRejectsInvalidKind(t *testing.T) {
	cluster := Cluster{
		OrganizationID: 1,
		Slug:           slugkit.Slug("local"),
		Name:           "Local cluster",
		Kind:           "runner",
		Status:         StatusPending,
	}

	require.Error(t, cluster.Validate())
}

func TestClusterValidateRejectsInvalidStatus(t *testing.T) {
	cluster := Cluster{
		OrganizationID: 1,
		Slug:           slugkit.Slug("online"),
		Name:           "Online cluster",
		Kind:           KindOnline,
		Status:         "online",
	}

	require.Error(t, cluster.Validate())
}

func TestClusterValidateAcceptsLocalPendingCluster(t *testing.T) {
	cluster := Cluster{
		OrganizationID: 1,
		Slug:           slugkit.Slug("local"),
		Name:           "Local cluster",
		Kind:           KindLocal,
		Status:         StatusPending,
	}

	require.NoError(t, cluster.Validate())
}

func TestClusterValidateRejectsMissingOrganization(t *testing.T) {
	cluster := Cluster{
		Slug:   slugkit.Slug("local"),
		Name:   "Local cluster",
		Kind:   KindLocal,
		Status: StatusPending,
	}

	require.Error(t, cluster.Validate())
}

func TestClusterValidateRejectsMissingName(t *testing.T) {
	cluster := Cluster{
		OrganizationID: 1,
		Slug:           slugkit.Slug("local"),
		Kind:           KindLocal,
		Status:         StatusPending,
	}

	require.Error(t, cluster.Validate())
}

func TestClusterValidateRejectsMissingSlug(t *testing.T) {
	cluster := Cluster{
		OrganizationID: 1,
		Name:           "Local cluster",
		Kind:           KindLocal,
		Status:         StatusPending,
	}

	require.Error(t, cluster.Validate())
}
