package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrganizationApplicationsAuthorizesBeforeReading(t *testing.T) {
	events := []string{}
	reader := &organizationApplicationReaderStub{
		events: &events,
		items: []OrganizationApplication{
			{
				InstallationID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				InstalledAt:    time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC),
			},
		},
	}
	authorizer := &organizationApplicationsAuthorizerStub{events: &events}
	applications := NewOrganizationApplicationsService(reader, authorizer)

	items, err := applications.ListOrganizationApplications(context.Background(), 9, 14)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, []string{"authorize", "read"}, events)
	require.Equal(t, int64(9), reader.organizationID)
}

func TestOrganizationApplicationsDoesNotReadWhenAuthorizationFails(t *testing.T) {
	reader := &organizationApplicationReaderStub{}
	authorizer := &organizationApplicationsAuthorizerStub{
		err: ErrTargetOrganizationForbidden,
	}
	applications := NewOrganizationApplicationsService(reader, authorizer)

	_, err := applications.ListOrganizationApplications(context.Background(), 9, 14)

	require.ErrorIs(t, err, ErrTargetOrganizationForbidden)
	require.Zero(t, reader.organizationID)
}

type organizationApplicationReaderStub struct {
	items          []OrganizationApplication
	err            error
	events         *[]string
	organizationID int64
}

func (s *organizationApplicationReaderStub) ListOrganizationApplications(
	_ context.Context,
	organizationID int64,
) ([]OrganizationApplication, error) {
	if s.events != nil {
		*s.events = append(*s.events, "read")
	}
	s.organizationID = organizationID
	return s.items, s.err
}

type organizationApplicationsAuthorizerStub struct {
	err    error
	events *[]string
}

func (s *organizationApplicationsAuthorizerStub) Authorize(
	context.Context,
	int64,
	int64,
) error {
	if s.events != nil {
		*s.events = append(*s.events, "authorize")
	}
	return s.err
}

func TestOrganizationApplicationsRejectsInvalidOrganizationID(t *testing.T) {
	applications := NewOrganizationApplicationsService(
		&organizationApplicationReaderStub{err: errors.New("must not read")},
		&organizationApplicationsAuthorizerStub{},
	)

	_, err := applications.ListOrganizationApplications(context.Background(), 0, 14)

	require.ErrorIs(t, err, ErrInvalidInstallationRequest)
}
