package consumer

import (
	"strconv"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

func mapPlanResult(result service.InstallationPlanResult) gin.H {
	return gin.H{
		"installation_id": result.InstallationID,
		"operation_id":    result.OperationID,
		"plan": gin.H{
			"plan_id":                 result.PlanID,
			"plan_digest":             result.PlanDigest,
			"expires_at":              result.ExpiresAt,
			"listing_version_id":      strconv.FormatInt(result.ListingVersionID, 10),
			"estimated_credits_micro": strconv.FormatInt(result.EstimatedCredits, 10),
			"required_permissions":    result.Permissions,
			"blocking_issues":         []any{},
			"warnings":                []any{},
		},
	}
}

func mapApplyResult(result service.ApplyResult) gin.H {
	response := gin.H{
		"installation_id": result.InstallationID,
		"operation_id":    result.OperationID,
		"status":          result.Status,
		"stage":           result.Stage,
	}
	if result.RuntimeRef != "" {
		response["runtime_ref"] = result.RuntimeRef
	}
	if result.ErrorCode != "" {
		response["error"] = gin.H{
			"code":    result.ErrorCode,
			"message": result.ErrorMessage,
		}
	}
	return response
}
