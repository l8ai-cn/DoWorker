package conversationitem

import (
	"context"
	"errors"
	"testing"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/stretchr/testify/require"
)

func TestAppendUserTextRequiresWriter(t *testing.T) {
	err := AppendUserText(context.Background(), nil, "conv-review", "Review this.")

	require.EqualError(t, err, "conversation item service unavailable")
}

func TestAppendUserTextPropagatesPositionFailure(t *testing.T) {
	expected := errors.New("position unavailable")
	writer := &failingPositionedAppender{nextErr: expected}

	err := AppendUserText(
		context.Background(),
		writer,
		"conv-review",
		"Review this.",
	)

	require.ErrorIs(t, err, expected)
}

func TestAppendUserTextPropagatesAppendFailure(t *testing.T) {
	expected := errors.New("append unavailable")
	writer := &failingPositionedAppender{position: 1, appendErr: expected}

	err := AppendUserText(
		context.Background(),
		writer,
		"conv-review",
		"Review this.",
	)

	require.ErrorIs(t, err, expected)
}

type failingPositionedAppender struct {
	position  int64
	nextErr   error
	appendErr error
}

func (writer *failingPositionedAppender) NextPosition(
	context.Context,
	string,
) (int64, error) {
	return writer.position, writer.nextErr
}

func (writer *failingPositionedAppender) Append(
	context.Context,
	*domain.Item,
) error {
	return writer.appendErr
}
