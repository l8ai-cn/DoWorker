package workerspec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerSpecRejectsIncompleteModelBindingSnapshot(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ModelBinding)
		match  string
	}{
		{"resource id", func(binding *ModelBinding) {
			binding.ResourceID = 0
		}, "resource id"},
		{"resource revision", func(binding *ModelBinding) {
			binding.ResourceRevision = 0
		}, "resource revision"},
		{"connection id", func(binding *ModelBinding) {
			binding.ConnectionID = 0
		}, "connection id"},
		{"connection revision", func(binding *ModelBinding) {
			binding.ConnectionRevision = 0
		}, "connection revision"},
		{"provider key", func(binding *ModelBinding) {
			binding.ProviderKey = ""
		}, "provider key"},
		{"protocol adapter", func(binding *ModelBinding) {
			binding.ProtocolAdapter = ""
		}, "protocol adapter"},
		{"model id", func(binding *ModelBinding) {
			binding.ModelID = ""
		}, "model id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validWorkerSpec()
			test.mutate(&spec.Runtime.ModelBinding)

			_, err := NormalizeAndValidate(spec)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.match)
		})
	}
}

func TestWorkerSpecSummaryPreservesImmutableModelBinding(t *testing.T) {
	spec := validWorkerSpec()

	summary, err := Summarize(spec)
	require.NoError(t, err)
	assert.Equal(t, spec.Runtime.ModelBinding, summary.ModelBinding)
}

func TestWorkerSpecAllowsWorkerWithoutMainModelBinding(t *testing.T) {
	spec := validWorkerSpec()
	spec.Runtime.ModelBinding = ModelBinding{}

	normalized, err := NormalizeAndValidate(spec)

	require.NoError(t, err)
	assert.True(t, normalized.Runtime.ModelBinding.IsEmpty())
}
