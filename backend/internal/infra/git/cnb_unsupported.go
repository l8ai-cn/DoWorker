package git

import (
	"context"
	"errors"
)

var ErrCNBOperationNotSupported = errors.New("CNB git provider operation is not supported")

func (p *CNBProvider) ListBranches(ctx context.Context, projectID string) ([]*Branch, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetBranch(ctx context.Context, projectID, branchName string) (*Branch, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) CreateBranch(ctx context.Context, projectID, branchName, ref string) (*Branch, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) DeleteBranch(ctx context.Context, projectID, branchName string) error {
	return ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetMergeRequest(ctx context.Context, projectID string, mrIID int) (*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) ListMergeRequests(ctx context.Context, projectID string, state string, page, perPage int) ([]*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) ListMergeRequestsByBranch(ctx context.Context, projectID, sourceBranch, state string) ([]*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) CreateMergeRequest(ctx context.Context, req *CreateMRRequest) (*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) UpdateMergeRequest(ctx context.Context, projectID string, mrIID int, title, description string) (*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) MergeMergeRequest(ctx context.Context, projectID string, mrIID int) (*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) CloseMergeRequest(ctx context.Context, projectID string, mrIID int) (*MergeRequest, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetCommit(ctx context.Context, projectID, sha string) (*Commit, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) ListCommits(ctx context.Context, projectID, branch string, page, perPage int) ([]*Commit, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) RegisterWebhook(ctx context.Context, projectID string, config *WebhookConfig) (string, error) {
	return "", ErrCNBOperationNotSupported
}

func (p *CNBProvider) DeleteWebhook(ctx context.Context, projectID, webhookID string) error {
	return ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetFileContent(ctx context.Context, projectID, filePath, ref string) ([]byte, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) TriggerPipeline(ctx context.Context, projectID string, req *TriggerPipelineRequest) (*Pipeline, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetPipeline(ctx context.Context, projectID string, pipelineID int) (*Pipeline, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) ListPipelines(ctx context.Context, projectID string, ref, status string, page, perPage int) ([]*Pipeline, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) CancelPipeline(ctx context.Context, projectID string, pipelineID int) (*Pipeline, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) RetryPipeline(ctx context.Context, projectID string, pipelineID int) (*Pipeline, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetJob(ctx context.Context, projectID string, jobID int) (*Job, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) ListPipelineJobs(ctx context.Context, projectID string, pipelineID int) ([]*Job, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) RetryJob(ctx context.Context, projectID string, jobID int) (*Job, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) CancelJob(ctx context.Context, projectID string, jobID int) (*Job, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetJobTrace(ctx context.Context, projectID string, jobID int) (string, error) {
	return "", ErrCNBOperationNotSupported
}

func (p *CNBProvider) GetJobArtifact(ctx context.Context, projectID string, jobID int, artifactPath string) ([]byte, error) {
	return nil, ErrCNBOperationNotSupported
}

func (p *CNBProvider) DownloadJobArtifacts(ctx context.Context, projectID string, jobID int) ([]byte, error) {
	return nil, ErrCNBOperationNotSupported
}
