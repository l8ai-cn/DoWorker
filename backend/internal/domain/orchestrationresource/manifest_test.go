package orchestrationresource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func validManifestForTest() Manifest {
	return Manifest{
		TypeMeta: TypeMeta{
			APIVersion: APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		},
		Metadata: validMetadataForTest(),
		Spec:     json.RawMessage(`{"runtime":"codex"}`),
	}
}

func TestManifestValidationAcceptsValidManifest(t *testing.T) {
	submission := validManifestForTest()
	submission.Metadata.UID = ""
	submission.Metadata.ResourceVersion = ""
	submission.Metadata.Generation = 0
	require.NoError(t, submission.ValidateSubmission())

	stored := validManifestForTest()
	stored.Status = json.RawMessage(`{}`)
	require.NoError(t, stored.ValidateStored())
}

func TestManifestValidationCallsTypeMetaAndMetadataValidation(t *testing.T) {
	validators := []struct {
		name     string
		validate func(Manifest) error
	}{
		{name: "submission", validate: Manifest.ValidateSubmission},
		{name: "stored", validate: Manifest.ValidateStored},
	}
	tests := []struct {
		name   string
		path   string
		mutate func(*Manifest)
	}{
		{
			name: "type metadata",
			path: "typeMeta.APIVersion",
			mutate: func(manifest *Manifest) {
				manifest.APIVersion = "v1"
			},
		},
		{
			name: "metadata",
			path: "metadata.name",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.Name = ""
			},
		},
	}

	for _, validator := range validators {
		for _, tt := range tests {
			t.Run(validator.name+"/"+tt.name, func(t *testing.T) {
				manifest := validManifestForTest()
				manifest.Metadata.UID = ""
				manifest.Metadata.ResourceVersion = ""
				manifest.Metadata.Generation = 0
				tt.mutate(&manifest)

				err := validator.validate(manifest)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.path)
			})
		}
	}
}

func TestManifestValidationRequiresNonEmptyJSONObjectSpec(t *testing.T) {
	tests := []struct {
		name string
		spec json.RawMessage
	}{
		{name: "nil", spec: nil},
		{name: "empty", spec: json.RawMessage{}},
		{name: "whitespace", spec: json.RawMessage(" \n\t")},
		{name: "null", spec: json.RawMessage(`null`)},
		{name: "array", spec: json.RawMessage(`[]`)},
		{name: "string", spec: json.RawMessage(`"worker"`)},
		{name: "number", spec: json.RawMessage(`42`)},
		{name: "boolean", spec: json.RawMessage(`true`)},
		{name: "empty object", spec: json.RawMessage(`{}`)},
		{name: "invalid json", spec: json.RawMessage(`{"runtime":`)},
		{name: "trailing json", spec: json.RawMessage(`{"runtime":"codex"} {}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			manifest.Metadata.UID = ""
			manifest.Metadata.ResourceVersion = ""
			manifest.Metadata.Generation = 0
			manifest.Spec = tt.spec

			submissionErr := manifest.ValidateSubmission()
			require.Error(t, submissionErr)
			require.Contains(t, submissionErr.Error(), "spec")

			storedErr := manifest.ValidateStored()
			require.Error(t, storedErr)
			require.Contains(t, storedErr.Error(), "spec")
		})
	}
}

func TestManifestValidationRejectsOversizedRawMessages(t *testing.T) {
	oversized := json.RawMessage(`{"payload":"` + strings.Repeat("a", 1<<20) + `"}`)
	tests := []struct {
		name   string
		path   string
		mutate func(*Manifest)
	}{
		{
			name: "spec",
			path: "spec",
			mutate: func(manifest *Manifest) {
				manifest.Spec = oversized
			},
		},
		{
			name: "status",
			path: "status",
			mutate: func(manifest *Manifest) {
				manifest.Status = oversized
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			tt.mutate(&manifest)

			err := manifest.ValidateStored()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestManifestSubmissionRejectsServerManagedFields(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mutate func(*Manifest)
	}{
		{
			name: "uid",
			path: "metadata.uid",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.UID = "res_123"
			},
		},
		{
			name: "uid containing control character",
			path: "metadata.uid",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.UID = "res\n123"
			},
		},
		{
			name: "resource version",
			path: "metadata.resourceVersion",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.ResourceVersion = "42"
			},
		},
		{
			name: "resource version exceeding maximum length",
			path: "metadata.resourceVersion",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.ResourceVersion = strings.Repeat("版", 129)
			},
		},
		{
			name: "generation",
			path: "metadata.generation",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.Generation = 1
			},
		},
		{
			name: "negative generation",
			path: "metadata.generation",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.Generation = -1
			},
		},
		{
			name: "status object",
			path: "status",
			mutate: func(manifest *Manifest) {
				manifest.Status = json.RawMessage(`{"ready":true}`)
			},
		},
		{
			name: "status whitespace",
			path: "status",
			mutate: func(manifest *Manifest) {
				manifest.Status = json.RawMessage(" ")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			manifest.Metadata.UID = ""
			manifest.Metadata.ResourceVersion = ""
			manifest.Metadata.Generation = 0
			tt.mutate(&manifest)

			err := manifest.ValidateSubmission()
			require.ErrorIs(t, err, ErrServerManagedField)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestManifestStoredValidatesServerManagedFields(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mutate func(*Manifest)
	}{
		{
			name: "uid containing control character",
			path: "metadata.uid",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.UID = "res\n123"
			},
		},
		{
			name: "resource version exceeding maximum length",
			path: "metadata.resourceVersion",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.ResourceVersion = strings.Repeat("版", 129)
			},
		},
		{
			name: "negative generation",
			path: "metadata.generation",
			mutate: func(manifest *Manifest) {
				manifest.Metadata.Generation = -1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			tt.mutate(&manifest)

			err := manifest.ValidateStored()
			require.Error(t, err)
			require.NotErrorIs(t, err, ErrServerManagedField)
			require.Contains(t, err.Error(), tt.path)
		})
	}
}

func TestManifestStoredStatusRules(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Status = json.RawMessage(`{"ready":true}`)
	require.NoError(t, manifest.ValidateStored())

	manifest.Status = json.RawMessage(`{}`)
	require.NoError(t, manifest.ValidateStored())

	tests := []struct {
		name   string
		status json.RawMessage
	}{
		{name: "whitespace", status: json.RawMessage(" \n")},
		{name: "null", status: json.RawMessage(`null`)},
		{name: "array", status: json.RawMessage(`[]`)},
		{name: "string", status: json.RawMessage(`"ready"`)},
		{name: "number", status: json.RawMessage(`1`)},
		{name: "boolean", status: json.RawMessage(`false`)},
		{name: "invalid json", status: json.RawMessage(`{"ready":`)},
		{name: "trailing json", status: json.RawMessage(`{} {}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifestForTest()
			manifest.Status = tt.status

			err := manifest.ValidateStored()
			require.Error(t, err)
			require.Contains(t, err.Error(), "status")
		})
	}
}
