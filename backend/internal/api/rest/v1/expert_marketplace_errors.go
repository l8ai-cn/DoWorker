package v1

import (
	"errors"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	expertsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func (h *ExpertHandler) marketplaceError(
	c *gin.Context,
	err error,
	internalMessage string,
) {
	var dependencyError *expertsvc.MarketDependencyError
	switch {
	case errors.Is(err, expertdom.ErrNotFound),
		errors.Is(err, expertmarket.ErrNotFound),
		errors.Is(err, expertsvc.ErrMarketApplicationNotFound):
		apierr.ResourceNotFound(c, "Marketplace resource not found")
	case errors.Is(err, expertsvc.ErrMarketUnavailable):
		apierr.ServiceUnavailable(
			c,
			apierr.SERVICE_UNAVAILABLE,
			"Expert marketplace is unavailable",
		)
	case errors.Is(err, expertsvc.ErrMarketApplicationOwnership):
		apierr.Forbidden(
			c,
			apierr.ACCESS_DENIED,
			"Marketplace application belongs to another organization",
		)
	case errors.Is(err, expertsvc.ErrMarketInvalidTransition),
		errors.Is(err, expertsvc.ErrMarketApplicationSlugMismatch),
		errors.Is(err, expertsvc.ErrMarketReleaseNotPublished),
		errors.Is(err, expertsvc.ErrMarketSourceSnapshotRequired),
		errors.Is(err, expertsvc.ErrMarketSnapshotInvalid),
		errors.Is(err, expertmarket.ErrPendingReleaseExists),
		errors.Is(err, expertmarket.ErrLifecycleStatusConflict),
		errors.Is(err, expertmarket.ErrConflict),
		errors.Is(err, expertdom.ErrConflict),
		errors.As(err, &dependencyError):
		apierr.Conflict(
			c,
			apierr.ALREADY_EXISTS,
			"Marketplace state changed or is not ready for this operation",
		)
	default:
		apierr.InternalError(c, internalMessage)
	}
}
