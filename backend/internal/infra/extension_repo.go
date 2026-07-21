package infra

import (
	"strings"

	"gorm.io/gorm"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
)

// isDuplicateKeyError checks whether the given error is a database unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key value") || // PostgreSQL
		strings.Contains(errStr, "UNIQUE constraint failed") || // SQLite
		strings.Contains(errStr, "Duplicate entry") // MySQL
}

// escapeLike escapes special LIKE/ILIKE characters (%, _, \) in a search string.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// extensionRepo implements extension.Repository using GORM
type extensionRepo struct {
	db *gorm.DB
}

// NewExtensionRepository creates a new extension repository
func NewExtensionRepository(db *gorm.DB) extension.Repository {
	return &extensionRepo{db: db}
}
