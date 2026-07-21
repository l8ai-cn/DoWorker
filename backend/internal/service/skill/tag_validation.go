package skill

import (
	"fmt"
	"unicode/utf8"

	"github.com/lib/pq"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
)

const (
	maxSkillTags       = 20
	maxSkillTagRunes   = 40
	maxSkillTotalRunes = 400
)

func ValidateTags(tags []string) (pq.StringArray, error) {
	normalized := skilldom.NormalizeTags(tags)
	if len(normalized) > maxSkillTags {
		return nil, fmt.Errorf("%w: at most %d tags", ErrInvalidTags, maxSkillTags)
	}
	total := 0
	for _, tag := range normalized {
		length := utf8.RuneCountInString(tag)
		if length > maxSkillTagRunes {
			return nil, fmt.Errorf(
				"%w: tag %q exceeds %d characters",
				ErrInvalidTags,
				tag,
				maxSkillTagRunes,
			)
		}
		total += length
	}
	if total > maxSkillTotalRunes {
		return nil, fmt.Errorf(
			"%w: total tag length exceeds %d characters",
			ErrInvalidTags,
			maxSkillTotalRunes,
		)
	}
	return normalized, nil
}
