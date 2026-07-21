package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func validDraftReference() Reference {
	return Reference{
		Kind:     "WorkerTemplate",
		Name:     slugkit.MustNewForTest("agent-one"),
		UID:      "",
		Digest:   "",
		Revision: 1,
	}
}

func validResolvedReference() Reference {
	return Reference{
		APIVersion: APIVersionV1Alpha1,
		Kind:       "WorkerTemplate",
		Namespace:  slugkit.MustNewForTest("team-alpha"),
		Name:       slugkit.MustNewForTest("agent-one"),
		UID:        "uid-01",
		Revision:   7,
		Digest:     "sha256:" + strings.Repeat("a", 64),
	}
}

func TestReferenceValidateDraftAcceptsOmittedDefaultsAndDoesNotMutate(t *testing.T) {
	ref := validDraftReference()
	err := ref.ValidateDraft("team-alpha")
	require.NoError(t, err)
	require.Equal(t, "", ref.APIVersion)
	require.Equal(t, slugkit.Slug(""), ref.Namespace)
}

func TestReferenceValidateDraftAcceptsPinnedVersionAndNamespace(t *testing.T) {
	ref := validDraftReference()
	ref.APIVersion = APIVersionV1Alpha1
	ref.Namespace = slugkit.MustNewForTest("team-alpha")
	require.NoError(t, ref.ValidateDraft("team-alpha"))
}

func TestReferenceValidateDraftAcceptsRevisionZero(t *testing.T) {
	ref := validDraftReference()
	ref.Revision = 0
	require.NoError(t, ref.ValidateDraft("team-alpha"))
}

func TestReferenceValidateDraftRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		do   func(Reference) Reference
		msg  string
	}{
		{
			name: "invalid version",
			do: func(r Reference) Reference {
				r.APIVersion = "bad"
				return r
			},
			msg: "typeMeta.APIVersion",
		},
		{
			name: "invalid kind",
			do: func(r Reference) Reference {
				r.Kind = "worker-template"
				return r
			},
			msg: "typeMeta.Kind",
		},
		{
			name: "invalid name",
			do: func(r Reference) Reference {
				r.Name = "BadName"
				return r
			},
			msg: "reference.name",
		},
		{
			name: "explicit namespace mismatch",
			do: func(r Reference) Reference {
				r.Namespace = slugkit.MustNewForTest("team-beta")
				return r
			},
			msg: "reference.namespace",
		},
		{
			name: "uid is forbidden",
			do: func(r Reference) Reference {
				r.UID = "uid-01"
				return r
			},
			msg: "reference.uid",
		},
		{
			name: "digest is forbidden",
			do: func(r Reference) Reference {
				r.Digest = "sha256:" + strings.Repeat("f", 64)
				return r
			},
			msg: "reference.digest",
		},
		{
			name: "negative revision",
			do: func(r Reference) Reference {
				r.Revision = -1
				return r
			},
			msg: "reference.revision",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.do(validDraftReference()).ValidateDraft("team-alpha")
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.msg)
		})
	}
}

func TestReferenceValidateDraftRejectsInvalidDefaultOrExplicitNamespace(t *testing.T) {
	t.Run("invalid default namespace", func(t *testing.T) {
		err := validDraftReference().ValidateDraft("team_alpha")
		require.Error(t, err)
		require.Contains(t, err.Error(), "defaultNamespace")
		require.ErrorIs(t, err, slugkit.ErrInvalidFormat)
	})

	t.Run("invalid explicit namespace", func(t *testing.T) {
		ref := validDraftReference()
		ref.Namespace = "team_alpha"
		err := ref.ValidateDraft("team-alpha")
		require.Error(t, err)
	})
}

func TestReferenceValidateResolvedAcceptsValid(t *testing.T) {
	require.NoError(t, validResolvedReference().ValidateResolved("team-alpha"))
}

func TestReferenceValidateResolvedRejectsInvalidDefaultNamespace(t *testing.T) {
	err := validResolvedReference().ValidateResolved("team_alpha")
	require.Error(t, err)
	require.Contains(t, err.Error(), "defaultNamespace")
	require.ErrorIs(t, err, slugkit.ErrInvalidFormat)
}

func TestReferenceValidateResolvedRejectsMissingImmutableFields(t *testing.T) {
	tests := []struct {
		name string
		do   func(Reference) Reference
		msg  string
	}{
		{name: "missing api version", do: func(r Reference) Reference { r.APIVersion = ""; return r }, msg: "typeMeta.APIVersion"},
		{name: "missing namespace", do: func(r Reference) Reference { r.Namespace = ""; return r }, msg: "reference.namespace"},
		{name: "missing name", do: func(r Reference) Reference { r.Name = ""; return r }, msg: "reference.name"},
		{name: "missing kind", do: func(r Reference) Reference { r.Kind = ""; return r }, msg: "typeMeta.Kind"},
		{name: "missing uid", do: func(r Reference) Reference { r.UID = ""; return r }, msg: "reference.uid"},
		{name: "missing revision", do: func(r Reference) Reference { r.Revision = 0; return r }, msg: "reference.revision"},
		{name: "missing digest", do: func(r Reference) Reference { r.Digest = ""; return r }, msg: "reference.digest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.do(validResolvedReference()).ValidateResolved("team-alpha")
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.msg)
		})
	}
}

func TestReferenceValidateResolvedRejectsDigestAndUIDRules(t *testing.T) {
	tests := []struct {
		name string
		do   func(Reference) Reference
		msg  string
	}{
		{
			name: "upper-case digest",
			do: func(r Reference) Reference {
				r.Digest = "SHA256:" + strings.Repeat("A", 64)
				return r
			},
			msg: "reference.digest",
		},
		{
			name: "short digest",
			do: func(r Reference) Reference {
				r.Digest = "sha256:" + strings.Repeat("a", 10)
				return r
			},
			msg: "reference.digest",
		},
		{
			name: "long digest",
			do: func(r Reference) Reference {
				r.Digest = "sha256:" + strings.Repeat("a", 65)
				return r
			},
			msg: "reference.digest",
		},
		{
			name: "uid control char",
			do: func(r Reference) Reference {
				r.UID = "uid\n42"
				return r
			},
			msg: "reference.uid",
		},
		{
			name: "uid overlong",
			do: func(r Reference) Reference {
				r.UID = strings.Repeat("界", 129)
				return r
			},
			msg: "reference.uid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.do(validResolvedReference()).ValidateResolved("team-alpha")
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.msg)
		})
	}
}

func TestReferenceValidateResolvedRejectsCrossNamespace(t *testing.T) {
	ref := validResolvedReference()
	ref.Namespace = slugkit.MustNewForTest("team-beta")
	err := ref.ValidateResolved("team-alpha")
	require.ErrorIs(t, err, ErrCrossNamespaceReference)
	require.Contains(t, err.Error(), "reference.namespace")
}
