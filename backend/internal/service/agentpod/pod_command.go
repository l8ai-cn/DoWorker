package agentpod

import (
	"fmt"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/ticket"
)

func BuildTicketPrompt(t *ticket.Ticket) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Working on ticket: %s", t.Slug))
	parts = append(parts, fmt.Sprintf("Title: %s", t.Title))
	return strings.Join(parts, "\n")
}
