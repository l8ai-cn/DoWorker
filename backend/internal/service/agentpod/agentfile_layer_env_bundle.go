package agentpod

import (
	"context"
	"fmt"
	"strings"
)

type PrimaryCredentialResolver interface {
	PrimaryCredentialBundleName(ctx context.Context, userID, orgID int64, agentSlug string) (string, error)
}

func AppendPrimaryCredentialBundle(
	ctx context.Context,
	resolver PrimaryCredentialResolver,
	userID, orgID int64,
	agentSlug string,
	layer **string,
) {
	if resolver == nil || strings.TrimSpace(agentSlug) == "" {
		return
	}
	name, err := resolver.PrimaryCredentialBundleName(ctx, userID, orgID, agentSlug)
	if err != nil || name == "" {
		return
	}
	if layerHasEnvBundle(layer, name) {
		return
	}
	appendAgentfileLayerLines(layer, fmt.Sprintf(`USE_ENV_BUNDLE "%s"`, name))
}

func layerHasEnvBundle(layer **string, name string) bool {
	if layer == nil || *layer == nil {
		return false
	}
	return strings.Contains(**layer, fmt.Sprintf(`USE_ENV_BUNDLE "%s"`, name))
}

func appendAgentfileLayerLines(layer **string, lines ...string) {
	var merged []string
	if layer != nil && *layer != nil {
		for _, line := range strings.Split(**layer, "\n") {
			if t := strings.TrimSpace(line); t != "" {
				merged = append(merged, t)
			}
		}
	}
	for _, line := range lines {
		if t := strings.TrimSpace(line); t != "" {
			merged = append(merged, t)
		}
	}
	if len(merged) == 0 {
		return
	}
	out := strings.Join(merged, "\n")
	if layer == nil {
		return
	}
	*layer = &out
}
