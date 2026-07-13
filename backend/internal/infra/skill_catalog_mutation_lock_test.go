package infra

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteSkillMutationLockAlwaysReleases(t *testing.T) {
	var events []string
	mutationErr := errors.New("mutation failed")
	releaseErr := errors.New("release failed")

	err := executeSkillMutationLock(
		func() error {
			events = append(events, "acquire")
			return nil
		},
		func() error {
			events = append(events, "release")
			return releaseErr
		},
		func() error {
			events = append(events, "mutate")
			return mutationErr
		},
	)

	assert.Equal(t, []string{"acquire", "mutate", "release"}, events)
	require.ErrorIs(t, err, mutationErr)
	require.ErrorIs(t, err, releaseErr)
}

func TestExecuteSkillMutationLockRepanicsAfterRelease(t *testing.T) {
	var released bool
	mutationErr := errors.New("mutation panic")
	releaseErr := errors.New("release failed")

	recovered := func() (recovered any) {
		defer func() { recovered = recover() }()
		_ = executeSkillMutationLock(
			func() error { return nil },
			func() error {
				released = true
				return releaseErr
			},
			func() error { panic(mutationErr) },
		)
		return nil
	}()

	assert.True(t, released)
	combined, ok := recovered.(error)
	require.True(t, ok)
	require.ErrorIs(t, combined, mutationErr)
	require.ErrorIs(t, combined, releaseErr)
	require.ErrorContains(t, combined, "mutation panic")
	var panicErr *SkillMutationPanic
	require.ErrorAs(t, combined, &panicErr)
	assert.Same(t, mutationErr, panicErr.Value)
}

func TestExecuteSkillMutationLockCombinesMutationAndReleasePanics(t *testing.T) {
	mutationErr := errors.New("mutation panic")
	releasePanic := errors.New("release panic")

	recovered := func() (recovered any) {
		defer func() { recovered = recover() }()
		_ = executeSkillMutationLock(
			func() error { return nil },
			func() error { panic(releasePanic) },
			func() error { panic(mutationErr) },
		)
		return nil
	}()

	combined, ok := recovered.(error)
	require.True(t, ok)
	require.ErrorIs(t, combined, mutationErr)
	require.ErrorIs(t, combined, releasePanic)
	require.ErrorContains(t, combined, "mutation panic")
	require.ErrorContains(t, combined, "release panic")
}
