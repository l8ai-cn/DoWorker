package v1

import (
	"errors"
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

func mapOrchestratorErrorToHTTP(c *gin.Context, err error) {
	switch {
	case errors.Is(err, agentpod.ErrCreateResourceUnavailable):
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Selected repository is unavailable")
	case errors.Is(err, agentpod.ErrMissingRunnerID):
		apierr.BadRequest(c, apierr.MISSING_RUNNER_ID, err.Error())
	case errors.Is(err, agentpod.ErrMissingAgentSlug):
		apierr.BadRequest(c, apierr.MISSING_AGENT_SLUG, err.Error())
	case errors.Is(err, agentpod.ErrSourcePodNotTerminated):
		apierr.BadRequest(c, apierr.SOURCE_POD_NOT_TERMINATED, "Can only resume from terminated, completed, or orphaned pods")
	case errors.Is(err, agentpod.ErrResumeRunnerMismatch):
		apierr.BadRequest(c, apierr.RESUME_RUNNER_MISMATCH, "Resume requires same runner as source pod (Sandbox is local to runner)")
	case errors.Is(err, agentpod.ErrUnsupportedInteractionMode):
		apierr.BadRequest(c, apierr.UNSUPPORTED_INTERACTION_MODE, err.Error())
	case errors.Is(err, agentpod.ErrInvalidAgentfileLayer):
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())

	case errors.Is(err, ErrQuotaExceeded):
		apierr.PaymentRequired(c, apierr.CONCURRENT_POD_QUOTA_EXCEEDED, "Concurrent pod quota exceeded. Please upgrade your plan or terminate existing pods.")
	case errors.Is(err, ErrSubscriptionFrozen):
		apierr.PaymentRequired(c, apierr.SUBSCRIPTION_FROZEN, "Your subscription has expired. Please renew to continue.")

	case errors.Is(err, agentpod.ErrSourcePodAccessDenied):
		apierr.Forbidden(c, apierr.SOURCE_POD_ACCESS_DENIED, "Source pod belongs to different organization")

	case errors.Is(err, agentpod.ErrSourcePodNotFound):
		apierr.NotFound(c, apierr.SOURCE_POD_NOT_FOUND, "Source pod not found for resume")

	case errors.Is(err, agentpod.ErrSourcePodAlreadyResumed):
		apierr.Conflict(c, apierr.SOURCE_POD_ALREADY_RESUMED, "Source pod has already been resumed by another active pod")
	case errors.Is(err, ErrSandboxAlreadyResumed):
		apierr.Conflict(c, apierr.SANDBOX_ALREADY_RESUMED, "Sandbox has already been resumed by another active pod")

	case errors.Is(err, agentpod.ErrNoAvailableRunner):
		apierr.ServiceUnavailable(c, apierr.NO_AVAILABLE_RUNNER, "No available runner supports the requested agent")

	case errors.Is(err, agentpod.ErrRunnerDispatchFailed):
		apierr.Respond(c, http.StatusBadGateway, apierr.RUNNER_DISPATCH_FAILED, "Failed to dispatch pod to runner. The runner may be offline or unreachable.")

	case errors.Is(err, agentpod.ErrConfigBuildFailed):
		apierr.Respond(c, http.StatusInternalServerError, apierr.POD_CONFIG_BUILD_FAILED, "Failed to build pod configuration")

	default:
		apierr.InternalError(c, "Failed to create pod")
	}
}
