package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"regexp"
)

type PlanOperation string

const (
	PlanOperationCreate PlanOperation = "create"
	PlanOperationUpdate PlanOperation = "update"
)

type PlanStatus string

const (
	PlanStatusPending   PlanStatus = "pending"
	PlanStatusApplied   PlanStatus = "applied"
	PlanStatusCancelled PlanStatus = "cancelled"
	PlanStatusExpired   PlanStatus = "expired"
)

type PlanIssueSeverity string

const (
	PlanIssueBlocking PlanIssueSeverity = "blocking"
	PlanIssueWarning  PlanIssueSeverity = "warning"
)

type SemanticChangeOperation string

const (
	SemanticChangeAdd     SemanticChangeOperation = "add"
	SemanticChangeRemove  SemanticChangeOperation = "remove"
	SemanticChangeReplace SemanticChangeOperation = "replace"
)

type PlanIssue struct {
	Severity PlanIssueSeverity `json:"severity"`
	Path     string            `json:"path"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
}

type ChangeValue struct {
	Digest       string          `json:"digest,omitempty"`
	RedactedJSON json.RawMessage `json:"redactedJson,omitempty"`
}

type SemanticChange struct {
	Operation SemanticChangeOperation `json:"operation"`
	Path      string                  `json:"path"`
	Before    ChangeValue             `json:"before,omitempty"`
	After     ChangeValue             `json:"after,omitempty"`
}

var issueCodePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[.-][a-z0-9]+)*$`)

func (operation PlanOperation) validate() error {
	switch operation {
	case PlanOperationCreate, PlanOperationUpdate:
		return nil
	default:
		return invalid("operation", "must be create or update")
	}
}

func (status PlanStatus) validate() error {
	switch status {
	case PlanStatusPending, PlanStatusApplied, PlanStatusCancelled, PlanStatusExpired:
		return nil
	default:
		return invalid("status", "is unsupported")
	}
}

func (issue PlanIssue) Validate() error {
	switch issue.Severity {
	case PlanIssueBlocking, PlanIssueWarning:
	default:
		return invalid("issue.severity", "must be blocking or warning")
	}
	if err := validateJSONPointer("issue.path", issue.Path); err != nil {
		return err
	}
	if len(issue.Code) > 100 || !issueCodePattern.MatchString(issue.Code) {
		return invalid("issue.code", "must be a stable lowercase code")
	}
	if err := validateSafeText("issue.message", issue.Message, 1000, false); err != nil {
		return err
	}
	if containsSecretLikeText(issue.Message) {
		return invalid("issue.message", "must not contain secret-like data")
	}
	return nil
}

func (change SemanticChange) Validate() error {
	switch change.Operation {
	case SemanticChangeAdd, SemanticChangeRemove, SemanticChangeReplace:
	default:
		return invalid("semanticChange.operation", "must be add, remove, or replace")
	}
	if err := validateJSONPointer("semanticChange.path", change.Path); err != nil {
		return err
	}
	before, err := change.Before.validate("semanticChange.before")
	if err != nil {
		return err
	}
	after, err := change.After.validate("semanticChange.after")
	if err != nil {
		return err
	}
	switch change.Operation {
	case SemanticChangeAdd:
		if before || !after {
			return invalid("semanticChange", "add requires only an after value")
		}
	case SemanticChangeRemove:
		if !before || after {
			return invalid("semanticChange", "remove requires only a before value")
		}
	case SemanticChangeReplace:
		if !before || !after {
			return invalid("semanticChange", "replace requires before and after values")
		}
	}
	return nil
}

func (value ChangeValue) validate(field string) (bool, error) {
	hasDigest := value.Digest != ""
	hasJSON := len(value.RedactedJSON) != 0
	if hasDigest && hasJSON {
		return false, invalid(field, "must use either digest or redacted JSON")
	}
	if !hasDigest && !hasJSON {
		return false, nil
	}
	if hasDigest {
		if !digestPattern.MatchString(value.Digest) {
			return false, invalid(field, "digest must be lowercase SHA-256")
		}
		return true, nil
	}
	canonical, err := canonicalAnyJSON(value.RedactedJSON)
	if err != nil || !bytes.Equal(canonical, value.RedactedJSON) {
		return false, invalid(field, "redacted JSON must be canonical")
	}
	if err := rejectRawSecretJSON(value.RedactedJSON); err != nil {
		return false, invalid(field, "redacted JSON must not contain raw secrets")
	}
	return true, nil
}
