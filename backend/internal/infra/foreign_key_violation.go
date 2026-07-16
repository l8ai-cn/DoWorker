package infra

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23503") ||
		strings.Contains(msg, "foreign key constraint fails") ||
		strings.Contains(msg, "FOREIGN KEY constraint failed")
}
