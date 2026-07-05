package agentpod

import (
	"context"
	"log/slog"

	"github.com/anthropics/agentsmesh/agentfile/capability"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func (o *PodOrchestrator) buildPodCommand(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	pod *podDomain.Pod,
	sourcePod *podDomain.Pod,
	isResumeMode bool,
	resolved *agentfileResolved,
) (*runnerv1.CreatePodCommand, error) {
	localPath := ""
	if isResumeMode && sourcePod != nil && sourcePod.SandboxPath != nil {
		localPath = *sourcePod.SandboxPath
	}

	effectiveBranch := firstNonEmptyPtr(resolved.BranchName, req.BranchName)
	effectiveRepoID := firstNonNilInt64(resolved.RepositoryID, req.RepositoryID)

	httpCloneURL, sshCloneURL := "", ""
	sourceBranch, preparationScript := "", ""
	preparationTimeout := 300
	if effectiveRepoID != nil && o.repoService != nil {
		repo, err := o.repoService.GetByID(ctx, *effectiveRepoID)
		if err == nil && repo != nil {
			httpCloneURL = repo.HttpCloneURL
			sshCloneURL = repo.SshCloneURL
			if repo.DefaultBranch != "" {
				sourceBranch = repo.DefaultBranch
			}
			if repo.PreparationScript != nil {
				preparationScript = *repo.PreparationScript
			}
			if repo.PreparationTimeout != nil {
				preparationTimeout = *repo.PreparationTimeout
			}
		}
	}
	if effectiveBranch != nil && *effectiveBranch != "" {
		sourceBranch = *effectiveBranch
	}

	ticketSlug := ""
	if req.TicketSlug != nil && *req.TicketSlug != "" {
		ticketSlug = *req.TicketSlug
	} else if req.TicketID != nil && o.ticketService != nil {
		t, err := o.ticketService.GetTicket(ctx, *req.TicketID)
		if err == nil && t != nil {
			ticketSlug = t.Slug
		}
	}

	credentialType, gitToken, sshPrivateKey := "", "", ""
	if o.userService != nil {
		gitCred := o.getUserGitCredential(ctx, req.UserID)
		if gitCred != nil {
			credentialType = gitCred.Type
			switch gitCred.Type {
			case "oauth", "pat":
				gitToken = gitCred.Token
			case "ssh_key":
				sshPrivateKey = gitCred.SSHPrivateKey
			}
		}
	}

	if localPath != "" {
		httpCloneURL = ""
		sshCloneURL = ""
	}

	var runnerAgentVersions map[string]string
	if o.runnerQuery != nil && req.RunnerID > 0 {
		r, err := o.runnerQuery.GetRunner(ctx, req.RunnerID)
		if err == nil && r != nil && len(r.AgentVersions) > 0 {
			runnerAgentVersions = make(map[string]string, len(r.AgentVersions))
			for _, v := range r.AgentVersions {
				runnerAgentVersions[v.Slug] = v.Version
			}
		}
	}

	knowledgeMounts, err := o.resolveKnowledgeMounts(ctx, req, resolved)
	if err != nil {
		return nil, err
	}

	buildReq := &agent.ConfigBuildRequest{
		AgentSlug:           req.AgentSlug,
		OrganizationID:      req.OrganizationID,
		UserID:              req.UserID,
		RepositoryID:        effectiveRepoID,
		HttpCloneURL:        httpCloneURL,
		SshCloneURL:         sshCloneURL,
		SourceBranch:        sourceBranch,
		CredentialType:      credentialType,
		GitToken:            gitToken,
		SSHPrivateKey:       sshPrivateKey,
		TicketSlug:          ticketSlug,
		PreparationScript:   preparationScript,
		PreparationTimeout:  preparationTimeout,
		LocalPath:           localPath,
		Prompt:              resolved.Prompt,
		PodKey:              pod.PodKey,
		MCPPort:             19000,
		Cols:                req.Cols,
		Rows:                req.Rows,
		RunnerAgentVersions: runnerAgentVersions,
		MergedAgentfileSource: resolved.MergedAgentfileSource,
		KnowledgeMounts:       knowledgeMounts,
		SessionMcpInstalled:   req.SessionMcpServers,
	}

	cmd, err := o.configBuilder.BuildPodCommand(ctx, buildReq)
	if err != nil {
		return nil, err
	}
	cmd.Perpetual = req.Perpetual
	if o.permissionPolicy != nil {
		var rules []*runnerv1.PolicyRuleSnapshot
		var err error
		if req.AgentSessionID != "" {
			rules, err = o.permissionPolicy.SnapshotForSession(ctx, req.OrganizationID, req.AgentSessionID, req.AgentSlug)
		} else {
			rules, err = o.permissionPolicy.SnapshotForPodCreate(ctx, req.OrganizationID, req.AgentSlug)
		}
		if err != nil {
			slog.WarnContext(ctx, "policy snapshot failed, pod will use fail-safe ASK", "org_id", req.OrganizationID, "error", err)
		} else {
			cmd.PolicyRules = rules
		}
	}
	if resolved.MergedAgentfileSource != "" {
		if caps := capability.ScanDeclarations(resolved.MergedAgentfileSource); len(caps) > 0 {
			cmd.DeclaredCapabilities = caps
		}
	}
	if extID := externalSessionIDForResume(req, sourcePod, isResumeMode); extID != "" {
		if cmd.EnvVars == nil {
			cmd.EnvVars = map[string]string{}
		}
		cmd.EnvVars["AGENTSMESH_RESUME_EXTERNAL_SESSION"] = extID
	}
	return cmd, nil
}

func externalSessionIDForResume(req *OrchestrateCreatePodRequest, sourcePod *podDomain.Pod, isResumeMode bool) string {
	if req.ResumeExternalSessionID != "" {
		return req.ResumeExternalSessionID
	}
	if isResumeMode && sourcePod != nil && sourcePod.ExternalSessionID != nil {
		return *sourcePod.ExternalSessionID
	}
	return ""
}
