package orchestrationresource

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/robfig/cron/v3"
)

var resourceWorkflowCronParser = cron.NewParser(
	cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
)

type WorkflowResourceSpec struct {
	WorkerTemplateRef  Reference         `json:"workerTemplateRef" yaml:"workerTemplateRef"`
	PromptRef          Reference         `json:"promptRef" yaml:"promptRef"`
	Inputs             map[string]string `json:"inputs" yaml:"inputs"`
	ExecutionMode      string            `json:"executionMode" yaml:"executionMode"`
	CronExpression     string            `json:"cronExpression,omitempty" yaml:"cronExpression,omitempty"`
	SandboxStrategy    string            `json:"sandboxStrategy" yaml:"sandboxStrategy"`
	SessionPersistence bool              `json:"sessionPersistence" yaml:"sessionPersistence"`
	ConcurrencyPolicy  string            `json:"concurrencyPolicy" yaml:"concurrencyPolicy"`
	MaxConcurrentRuns  int               `json:"maxConcurrentRuns" yaml:"maxConcurrentRuns"`
	MaxRetainedRuns    int               `json:"maxRetainedRuns" yaml:"maxRetainedRuns"`
	TimeoutMinutes     int               `json:"timeoutMinutes" yaml:"timeoutMinutes"`
	IdleTimeoutSeconds int               `json:"idleTimeoutSeconds" yaml:"idleTimeoutSeconds"`
	CallbackURL        string            `json:"callbackUrl,omitempty" yaml:"callbackUrl,omitempty"`
}

func workflowResourceSchema() Schema {
	return Schema{
		NewSpec: func() any { return &WorkflowResourceSpec{} },
		Validate: func(metadata Metadata, value any) error {
			return validateWorkflowResource(metadata, value.(*WorkflowResourceSpec))
		},
	}
}

func validateWorkflowResource(
	metadata Metadata,
	spec *WorkflowResourceSpec,
) error {
	if err := validateDefinitionReference(
		metadata,
		"workerTemplateRef",
		KindWorkerTemplate,
		spec.WorkerTemplateRef,
	); err != nil {
		return err
	}
	if err := validateDefinitionReference(
		metadata,
		"promptRef",
		KindPrompt,
		spec.PromptRef,
	); err != nil {
		return err
	}
	if err := validateDefinitionStringMap(
		"inputs",
		spec.Inputs,
		128,
		8_192,
	); err != nil {
		return err
	}
	if spec.ExecutionMode != "direct" && spec.ExecutionMode != "autopilot" {
		return fmt.Errorf("executionMode must be direct or autopilot")
	}
	if spec.SandboxStrategy != "fresh" &&
		spec.SandboxStrategy != "persistent" {
		return fmt.Errorf("sandboxStrategy must be fresh or persistent")
	}
	if spec.ConcurrencyPolicy != "skip" {
		return fmt.Errorf("concurrencyPolicy currently only supports skip")
	}
	if spec.MaxConcurrentRuns < 1 || spec.MaxConcurrentRuns > 100 {
		return fmt.Errorf("maxConcurrentRuns must be between 1 and 100")
	}
	if spec.SessionPersistence &&
		(spec.SandboxStrategy != "persistent" || spec.MaxConcurrentRuns != 1) {
		return fmt.Errorf(
			"sessionPersistence requires persistent sandbox and maxConcurrentRuns 1",
		)
	}
	if spec.MaxRetainedRuns < 0 || spec.MaxRetainedRuns > 10_000 {
		return fmt.Errorf("maxRetainedRuns must be between 0 and 10000")
	}
	if spec.TimeoutMinutes < 1 || spec.TimeoutMinutes > 1_440 {
		return fmt.Errorf("timeoutMinutes must be between 1 and 1440")
	}
	if spec.IdleTimeoutSeconds < 1 || spec.IdleTimeoutSeconds > 86_400 {
		return fmt.Errorf("idleTimeoutSeconds must be between 1 and 86400")
	}
	if err := validateDefinitionText(
		"cronExpression",
		spec.CronExpression,
		100,
		false,
	); err != nil {
		return err
	}
	if spec.CronExpression != "" {
		if _, err := resourceWorkflowCronParser.Parse(
			spec.CronExpression,
		); err != nil {
			return fmt.Errorf("cronExpression is invalid")
		}
	}
	return validateWorkflowCallbackURL(spec.CallbackURL)
}

func validateWorkflowCallbackURL(value string) error {
	if value == "" {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("callbackUrl must be an absolute HTTP(S) URL")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return fmt.Errorf("callbackUrl must not target a private or local host")
	}
	if ip := net.ParseIP(host); ip != nil &&
		(ip.IsLoopback() || ip.IsPrivate() ||
			ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return fmt.Errorf("callbackUrl must not target a private or local host")
	}
	return validateDefinitionText("callbackUrl", value, 500, false)
}
