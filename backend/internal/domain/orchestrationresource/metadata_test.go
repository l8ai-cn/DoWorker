package orchestrationresource

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func validMetadataForTest() Metadata {
	return Metadata{
		Name:            slugkit.MustNewForTest("worker-one"),
		Namespace:       slugkit.MustNewForTest("team-one"),
		DisplayName:     "  Worker 名称  ",
		Labels:          map[string]string{"role": "build-agent", "marker": ""},
		UID:             "res_123",
		ResourceVersion: "42+abc",
		Generation:      2,
	}
}

func TestMetadataValidateAcceptsValidMetadataWithoutRewriting(t *testing.T) {
	metadata := validMetadataForTest()

	require.NoError(t, metadata.Validate())
	require.Equal(t, "  Worker 名称  ", metadata.DisplayName)
	require.Equal(t, "build-agent", metadata.Labels["role"])
	require.Equal(t, "", metadata.Labels["marker"])
}

func TestMetadataValidateRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
		mutate  func(*Metadata)
	}{
		{
			name:    "empty name",
			path:    "metadata.name",
			wantErr: slugkit.ErrEmpty,
			mutate:  func(metadata *Metadata) { metadata.Name = "" },
		},
		{
			name:    "invalid name",
			path:    "metadata.name",
			wantErr: slugkit.ErrInvalidFormat,
			mutate:  func(metadata *Metadata) { metadata.Name = "Worker_One" },
		},
		{
			name:    "empty namespace",
			path:    "metadata.namespace",
			wantErr: slugkit.ErrEmpty,
			mutate:  func(metadata *Metadata) { metadata.Namespace = "" },
		},
		{
			name:    "invalid namespace",
			path:    "metadata.namespace",
			wantErr: slugkit.ErrInvalidFormat,
			mutate:  func(metadata *Metadata) { metadata.Namespace = "team.one" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := validMetadataForTest()
			tt.mutate(&metadata)

			err := metadata.Validate()
			require.ErrorIs(t, err, tt.wantErr)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestMetadataValidateDisplayNameRules(t *testing.T) {
	metadata := validMetadataForTest()
	metadata.DisplayName = strings.Repeat("界", 200)
	require.NoError(t, metadata.Validate())

	tests := []struct {
		name  string
		value string
	}{
		{name: "too long", value: strings.Repeat("界", 201)},
		{name: "invalid utf8", value: string([]byte{0xff})},
		{name: "ascii control", value: "worker\nname"},
		{name: "unicode control", value: "worker\u0085name"},
		{name: "bidi control", value: "worker\u202ename"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := validMetadataForTest()
			metadata.DisplayName = tt.value

			err := metadata.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), "metadata.displayName")
		})
	}
}

func TestMetadataValidateLabelRules(t *testing.T) {
	t.Run("accepts sixty four labels and empty marker values", func(t *testing.T) {
		metadata := validMetadataForTest()
		metadata.Labels = make(map[string]string, 64)
		for i := range 64 {
			key := slugkit.MustNewForTest("key-" + string(rune('a'+i/26)) + string(rune('a'+i%26)))
			metadata.Labels[key.String()] = ""
		}

		require.NoError(t, metadata.Validate())
	})

	tests := []struct {
		name   string
		labels map[string]string
		path   string
	}{
		{
			name:   "too many labels",
			labels: labelsForTest(65),
			path:   "metadata.labels",
		},
		{
			name:   "invalid key",
			labels: map[string]string{"Bad.Key": ""},
			path:   `metadata.labels["Bad.Key"]`,
		},
		{
			name:   "invalid non-empty value",
			labels: map[string]string{"valid-key": "Bad_Value"},
			path:   `metadata.labels["valid-key"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := validMetadataForTest()
			metadata.Labels = tt.labels

			err := metadata.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestMetadataValidateLabelErrorsSummarizeUntrustedKeys(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		untrustedKey := "key-" + strings.Repeat("a", 120)
		metadata := validMetadataForTest()
		metadata.Labels = map[string]string{untrustedKey: ""}

		err := metadata.Validate()
		require.ErrorIs(t, err, slugkit.ErrTooLong)
		require.Contains(t, err.Error(), "metadata.labels[")
		require.NotContains(t, err.Error(), untrustedKey)
	})

	t.Run("invalid value", func(t *testing.T) {
		untrustedKey := "k-" + strings.Repeat("a", 98)
		metadata := validMetadataForTest()
		metadata.Labels = map[string]string{untrustedKey: "Bad_Value"}

		err := metadata.Validate()
		require.ErrorIs(t, err, slugkit.ErrInvalidFormat)
		require.Contains(t, err.Error(), "metadata.labels[")
		require.NotContains(t, err.Error(), untrustedKey)
	})
}

func TestMetadataValidateServerFieldRules(t *testing.T) {
	metadata := validMetadataForTest()
	metadata.UID = strings.Repeat("界", 128)
	metadata.ResourceVersion = strings.Repeat("版", 128)
	metadata.Generation = 0
	require.NoError(t, metadata.Validate())

	tests := []struct {
		name   string
		path   string
		mutate func(*Metadata)
	}{
		{
			name: "negative generation",
			path: "metadata.generation",
			mutate: func(metadata *Metadata) {
				metadata.Generation = -1
			},
		},
		{
			name: "uid too long",
			path: "metadata.uid",
			mutate: func(metadata *Metadata) {
				metadata.UID = strings.Repeat("界", 129)
			},
		},
		{
			name: "uid contains control",
			path: "metadata.uid",
			mutate: func(metadata *Metadata) {
				metadata.UID = "res\n123"
			},
		},
		{
			name: "resource version too long",
			path: "metadata.resourceVersion",
			mutate: func(metadata *Metadata) {
				metadata.ResourceVersion = strings.Repeat("版", 129)
			},
		},
		{
			name: "resource version contains control",
			path: "metadata.resourceVersion",
			mutate: func(metadata *Metadata) {
				metadata.ResourceVersion = "42\u0085abc"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := validMetadataForTest()
			tt.mutate(&metadata)

			err := metadata.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func labelsForTest(count int) map[string]string {
	labels := make(map[string]string, count)
	for i := range count {
		labels["key-"+strings.Repeat("a", i+1)] = ""
	}
	return labels
}
