package agentpod

import (
	"context"
	"fmt"
	"log/slog"

	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// resolveKnowledgeMounts merges three mount sources — agent default mounts
// (DB), Agentfile KNOWLEDGE declarations, and per-request selections — into
// the KnowledgeMount list shipped to the Runner. Request selections are
// applied last so they win mode conflicts. Unknown slugs fail pod creation.
func (o *PodOrchestrator) resolveKnowledgeMounts(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	resolved *agentfileResolved,
) ([]*runnerv1.KnowledgeMount, error) {
	if o.knowledgeBases == nil {
		if len(resolved.Knowledge) > 0 || len(req.KnowledgeMounts) > 0 {
			return nil, fmt.Errorf("%w: knowledge base feature is not configured", ErrConfigBuildFailed)
		}
		return nil, nil
	}

	requested := make([]kbservice.MountRequest, 0, len(resolved.Knowledge)+len(req.KnowledgeMounts))
	for _, k := range resolved.Knowledge {
		requested = append(requested, kbservice.MountRequest{KBSlug: k.Slug, Mode: k.Mode})
	}
	for _, m := range req.KnowledgeMounts {
		requested = append(requested, kbservice.MountRequest{KBSlug: m.Slug, Mode: m.Mode})
	}

	mounts, err := o.knowledgeBases.ResolveMountsForPod(ctx, req.OrganizationID, req.AgentSlug, requested)
	if err != nil {
		return nil, err
	}
	if len(mounts) == 0 {
		return nil, nil
	}

	out := make([]*runnerv1.KnowledgeMount, 0, len(mounts))
	for _, m := range mounts {
		out = append(out, &runnerv1.KnowledgeMount{
			Slug:          m.KB.Slug,
			HttpCloneUrl:  m.KB.HTTPCloneURL,
			SshCloneUrl:   m.SSHCloneURL,
			Branch:        m.KB.DefaultBranch,
			MountPath:     "kb/" + m.KB.Slug,
			Mode:          m.Mode,
			GitKnownHosts: m.GitKnownHosts,
			GitPrivateKey: m.GitPrivateKey,
		})
	}
	slog.InfoContext(ctx, "knowledge mounts resolved", "org_id", req.OrganizationID, "agent_slug", req.AgentSlug, "count", len(out))
	return out, nil
}
