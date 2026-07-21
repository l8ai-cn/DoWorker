package agentpod

import (
	"context"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/agentfile/capability"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func (o *PodOrchestrator) buildPodCommand(
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
	pod *podDomain.Pod,
	sourcePod *podDomain.Pod,
	isResumeMode bool,
	resolved *agentfileResolved,
	adapterID string,
) (*runnerv1.CreatePodCommand, error) {
	localPath := ""
	if isResumeMode && sourcePod != nil && sourcePod.SandboxPath != nil {
		localPath = *sourcePod.SandboxPath
	} else if req.LocalPath != "" {
		localPath = req.LocalPath
	}

	effectiveBranch := firstNonEmptyPtr(resolved.BranchName, req.BranchName)
	effectiveRepoID := firstNonNilInt64(resolved.RepositoryID, req.RepositoryID)
	repository, err := podCommandRepository(req, resolved, effectiveBranch)
	if err != nil {
		return nil, err
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

	credential, err := o.podCommandGitCredential(ctx, req)
	if err != nil {
		return nil, err
	}

	if localPath != "" {
		repository.httpCloneURL = ""
		repository.sshCloneURL = ""
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
	requiredEnvBundleIDs, requiredSkillIDs, requiredConfigDocuments := workerSpecResourceRequirements(
		req.preparedWorkerSpec,
	)
	pinnedEnvBundles, pinnedConfigDocuments, err := workerDependencyRuntimeInputs(
		req.preResolvedDependencies,
	)
	if err != nil {
		return nil, err
	}

	buildReq := &agent.ConfigBuildRequest{
		AgentSlug:                      req.AgentSlug,
		OrganizationID:                 req.OrganizationID,
		UserID:                         req.UserID,
		RepositoryID:                   effectiveRepoID,
		HttpCloneURL:                   repository.httpCloneURL,
		SshCloneURL:                    repository.sshCloneURL,
		SourceBranch:                   repository.sourceBranch,
		SourceCommitSha:                repository.sourceCommitSha,
		CredentialType:                 credential.credentialType,
		GitToken:                       credential.token,
		SSHPrivateKey:                  credential.sshPrivateKey,
		TicketSlug:                     ticketSlug,
		PreparationScript:              repository.preparationScript,
		PreparationTimeout:             repository.preparationTimeout,
		LocalPath:                      localPath,
		Prompt:                         resolved.Prompt,
		PodKey:                         pod.PodKey,
		MCPPort:                        19000,
		Cols:                           req.Cols,
		Rows:                           req.Rows,
		RunnerAgentVersions:            runnerAgentVersions,
		MergedAgentfileSource:          resolved.MergedAgentfileSource,
		KnowledgeMounts:                knowledgeMounts,
		SessionMcpInstalled:            req.SessionMcpServers,
		SessionConfigBundles:           req.SessionConfigBundles,
		RequiredEnvBundleIDs:           requiredEnvBundleIDs,
		PinnedEnvBundles:               pinnedEnvBundles,
		PinnedConfigDocuments:          pinnedConfigDocuments,
		RequiredSkillIDs:               requiredSkillIDs,
		RequiredSkillPackages:          artifactSkillPackages(req.preResolvedDependencies),
		RequiredConfigDocumentBindings: requiredConfigDocuments,
	}
	if req.preResolvedDependencies != nil {
		buildReq.RequiredSkillIDs = nil
		buildReq.RequiredEnvBundleIDs = workerSpecSecretEnvBundleIDs(req.preparedWorkerSpec)
		buildReq.RequiredConfigDocumentBindings = nil
	}

	cmd, err := o.configBuilder.BuildPodCommand(ctx, buildReq)
	if err != nil {
		return nil, err
	}
	cmd.AdapterId = adapterID
	if len(req.ModelResourceEnv) > 0 {
		if cmd.EnvVars == nil {
			cmd.EnvVars = map[string]string{}
		}
		if err := applyModelResourceEnv(cmd.EnvVars, req.ModelResourceEnv); err != nil {
			return nil, err
		}
	}
	cmd.LaunchArgs, err = applyModelResourceArgs(cmd.LaunchArgs, req.ModelResourceArgs)
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
		cmd.EnvVars["AGENTCLOUD_RESUME_EXTERNAL_SESSION"] = extID
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
