package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type submitMarketApplicationRequest struct {
	Slug        string   `json:"slug"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Icon        string   `json:"icon"`
	Tags        []string `json:"tags"`
	Outcomes    []string `json:"outcomes"`
}

func (request *submitMarketApplicationRequest) validate() error {
	request.Slug = strings.TrimSpace(request.Slug)
	request.Summary = strings.TrimSpace(request.Summary)
	request.Description = strings.TrimSpace(request.Description)
	request.Category = strings.TrimSpace(request.Category)
	request.Icon = strings.TrimSpace(request.Icon)
	if err := slugkit.ValidateIdentifier(
		"expert_market_applications.slug",
		request.Slug,
	); err != nil {
		return err
	}
	if request.Summary == "" {
		return fmt.Errorf("summary is required")
	}
	if request.Category == "" {
		return fmt.Errorf("category is required")
	}
	if !marketplaceIconAllowed(request.Icon) {
		return fmt.Errorf("icon is unsupported")
	}
	return nil
}

func marketplaceIconAllowed(icon string) bool {
	switch icon {
	case "rocket", "network", "git-compare",
		"clapperboard", "scissors", "film":
		return true
	default:
		return false
	}
}

func marketplacePagination(c *gin.Context) (int, int, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100")
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		return 0, 0, fmt.Errorf("offset must be zero or greater")
	}
	return limit, offset, nil
}

func positivePathID(c *gin.Context, name string) (int64, error) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}
