package orchestrationworker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type workerCreationPreflighter interface {
	Revision() string
	Preflight(
		context.Context,
		specservice.Scope,
		workercreation.Draft,
	) (workercreation.PreflightResult, error)
}

type WorkerCreationCompiler struct {
	service  workerCreationPreflighter
	artifact func(*workercreation.Prepared) []byte
}

func NewWorkerCreationCompiler(
	service workerCreationPreflighter,
) (*WorkerCreationCompiler, error) {
	if service == nil || service.Revision() == "" {
		return nil, fmt.Errorf("worker creation service is unavailable")
	}
	return newWorkerCreationCompiler(service, nil), nil
}

func newWorkerCreationCompiler(
	service workerCreationPreflighter,
	artifact func(*workercreation.Prepared) []byte,
) *WorkerCreationCompiler {
	if artifact == nil {
		artifact = func(prepared *workercreation.Prepared) []byte {
			return prepared.Snapshot.SpecJSON()
		}
	}
	return &WorkerCreationCompiler{service: service, artifact: artifact}
}

func (compiler *WorkerCreationCompiler) Revision() string {
	if compiler == nil || compiler.service == nil {
		return ""
	}
	return compiler.service.Revision()
}

func (compiler *WorkerCreationCompiler) Compile(
	ctx context.Context,
	scope control.Scope,
	draft workercreation.Draft,
) (WorkerCompilation, error) {
	if compiler == nil || compiler.service == nil || compiler.artifact == nil {
		return WorkerCompilation{}, fmt.Errorf("worker creation compiler is unavailable")
	}
	result, err := compiler.service.Preflight(ctx, specservice.Scope{
		OrgID: scope.OrganizationID, UserID: scope.ActorID,
	}, draft)
	if err != nil {
		return WorkerCompilation{}, err
	}
	if result.OptionsRevision != compiler.Revision() {
		return WorkerCompilation{}, control.ErrCorrupt
	}
	issues := workerPlanIssues(result)
	if len(result.BlockingErrors) != 0 {
		return WorkerCompilation{Issues: issues}, nil
	}
	if result.Resolved == nil {
		return WorkerCompilation{}, control.ErrCorrupt
	}
	artifact, err := control.CanonicalJSONObject(compiler.artifact(result.Resolved))
	if err != nil {
		return WorkerCompilation{}, fmt.Errorf(
			"%w: invalid prepared WorkerSpec artifact",
			control.ErrCorrupt,
		)
	}
	return WorkerCompilation{ArtifactJSON: artifact, Issues: issues}, nil
}

func workerPlanIssues(result workercreation.PreflightResult) []control.PlanIssue {
	issues := make(
		[]control.PlanIssue,
		0,
		len(result.BlockingErrors)+len(result.Warnings),
	)
	for _, issue := range result.BlockingErrors {
		issues = append(issues, workerPlanIssue(
			issue,
			control.PlanIssueBlocking,
			"Worker template contains an invalid field.",
			"worker-template-invalid",
		))
	}
	for _, issue := range result.Warnings {
		issues = append(issues, workerPlanIssue(
			issue,
			control.PlanIssueWarning,
			"Worker template requires review.",
			"worker-template-warning",
		))
	}
	return issues
}

var workerIssueCodePattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*(?:[.-][a-z0-9]+)*$`,
)

func workerPlanIssue(
	source workercreation.Issue,
	severity control.PlanIssueSeverity,
	message string,
	defaultCode string,
) control.PlanIssue {
	code := source.Code
	if len(code) > 100 || !workerIssueCodePattern.MatchString(code) {
		code = defaultCode
	}
	return control.PlanIssue{
		Severity: severity,
		Path:     workerIssuePath(source.Field),
		Code:     code,
		Message:  message,
	}
}

func workerIssuePath(field string) string {
	switch field {
	case "", "draft", "worker_spec":
		return "/spec"
	case "options_revision":
		return "/spec/optionsRevision"
	}
	field = strings.TrimPrefix(field, "worker_spec.")
	segments := strings.Split(field, ".")
	path := "/spec"
	for _, segment := range segments {
		if segment == "" || !isWorkerIssuePathSegment(segment) {
			return "/spec"
		}
		path += "/" + snakeToLowerCamel(segment)
	}
	return path
}

func isWorkerIssuePathSegment(value string) bool {
	for _, char := range value {
		if char != '_' && (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') {
			return false
		}
	}
	return value != ""
}

func snakeToLowerCamel(value string) string {
	parts := strings.Split(value, "_")
	for index := 1; index < len(parts); index++ {
		if parts[index] == "" {
			continue
		}
		parts[index] = strings.ToUpper(parts[index][:1]) + parts[index][1:]
	}
	return strings.Join(parts, "")
}
