package workerdependency

import (
	"fmt"
	"sort"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/google/uuid"
)

func normalizePin(pin ResourcePin) ResourcePin {
	pin.Reference = normalizeReference(pin.Reference)
	return pin
}

func normalizeReference(
	reference orchestrationresource.Reference,
) orchestrationresource.Reference {
	reference.UID = strings.TrimSpace(reference.UID)
	reference.Digest = strings.TrimSpace(reference.Digest)
	if parsed, err := uuid.Parse(reference.UID); err == nil {
		reference.UID = parsed.String()
	}
	return reference
}

func normalizeFields(values []string) []string {
	normalized := append([]string{}, values...)
	for index := range normalized {
		normalized[index] = strings.TrimSpace(normalized[index])
	}
	sort.Strings(normalized)
	return normalized
}

func referenceKey(pin ResourcePin) string {
	return pin.Reference.UID + "\x00" +
		fmt.Sprintf("%d\x00%d", pin.Reference.Revision, pin.DomainID)
}
