package podconnect

import (
	"context"
	"errors"
	"strings"

	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	itemservice "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
)

func (s *Server) prepareWorkerInitialMessage(
	initialTask string,
) func(context.Context, *sessiondomain.Session) error {
	initialTask = strings.TrimSpace(initialTask)
	if initialTask == "" {
		return nil
	}
	return func(ctx context.Context, session *sessiondomain.Session) error {
		if s.conversationItems == nil {
			return errors.New("conversation item service unavailable")
		}
		if session == nil {
			return errors.New("session is required")
		}
		return itemservice.AppendUserText(
			ctx,
			s.conversationItems,
			session.ID,
			initialTask,
		)
	}
}
