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
	panicValue := errors.New("mutation panic")

	recovered := func() (recovered any) {
		defer func() { recovered = recover() }()
		_ = executeSkillMutationLock(
			func() error { return nil },
			func() error {
				released = true
				return errors.New("release failed")
			},
			func() error { panic(panicValue) },
		)
		return nil
	}()

	assert.True(t, released)
	assert.Same(t, panicValue, recovered)
}
