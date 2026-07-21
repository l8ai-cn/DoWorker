package agentworkbench

import agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"

func artifactGrantRepresentationIDsEnrich(
	current *agentworkbenchv2.ArtifactDescriptor,
	next *agentworkbenchv2.ArtifactDescriptor,
) bool {
	currentRepresentationIDs, currentValid := currentRevisionRepresentationIDs(current)
	nextRepresentationIDs, nextValid := currentRevisionRepresentationIDs(next)
	if !currentValid || !nextValid || len(current.GetGrants()) != len(next.GetGrants()) {
		return false
	}
	nextGrants, valid := artifactGrantsByID(next.GetGrants())
	if !valid {
		return false
	}
	for _, currentGrant := range current.GetGrants() {
		nextGrant, exists := nextGrants[currentGrant.GetGrantId()]
		if !exists || !grantRepresentationIDsEnrich(
			currentGrant,
			nextGrant,
			currentRepresentationIDs,
			nextRepresentationIDs,
		) {
			return false
		}
		delete(nextGrants, currentGrant.GetGrantId())
	}
	return len(nextGrants) == 0
}

func artifactGrantsByID(
	grants []*agentworkbenchv2.ArtifactGrant,
) (map[string]*agentworkbenchv2.ArtifactGrant, bool) {
	byID := make(map[string]*agentworkbenchv2.ArtifactGrant, len(grants))
	for _, grant := range grants {
		id := grant.GetGrantId()
		if id == "" {
			return nil, false
		}
		if _, exists := byID[id]; exists {
			return nil, false
		}
		byID[id] = grant
	}
	return byID, true
}

func grantRepresentationIDsEnrich(
	current *agentworkbenchv2.ArtifactGrant,
	next *agentworkbenchv2.ArtifactGrant,
	currentRepresentationIDs map[string]struct{},
	nextRepresentationIDs map[string]struct{},
) bool {
	currentIDs, currentValid := uniqueGrantRepresentationIDs(
		current.GetRepresentationIds(),
		currentRepresentationIDs,
	)
	nextIDs, nextValid := uniqueGrantRepresentationIDs(
		next.GetRepresentationIds(),
		nextRepresentationIDs,
	)
	if !currentValid || !nextValid {
		return false
	}
	for id := range currentIDs {
		if _, exists := nextIDs[id]; !exists {
			return false
		}
	}
	for id := range nextIDs {
		if _, existed := currentIDs[id]; existed {
			continue
		}
		if _, existed := currentRepresentationIDs[id]; existed {
			return false
		}
	}
	return true
}

func uniqueGrantRepresentationIDs(
	ids []string,
	representationIDs map[string]struct{},
) (map[string]struct{}, bool) {
	unique := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := representationIDs[id]; !exists {
			return nil, false
		}
		if _, exists := unique[id]; exists {
			return nil, false
		}
		unique[id] = struct{}{}
	}
	return unique, true
}

func clearArtifactGrantRepresentationIDs(
	artifact *agentworkbenchv2.ArtifactDescriptor,
) {
	for _, grant := range artifact.GetGrants() {
		grant.RepresentationIds = nil
	}
}
