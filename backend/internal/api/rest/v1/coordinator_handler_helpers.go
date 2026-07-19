package v1

import (
	"errors"
	"strconv"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func (h *CoordinatorHandler) parseID(c *gin.Context) int64 {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Invalid project id")
		return 0
	}
	return id
}

func (h *CoordinatorHandler) notFoundOrInternal(c *gin.Context, err error) {
	if errors.Is(err, coordinatordom.ErrNotFound) {
		apierr.ResourceNotFound(c, "Coordinator project not found")
		return
	}
	apierr.InternalError(c, "Coordinator request failed")
}
