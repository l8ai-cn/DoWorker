package agentpod

import (
	"errors"
	"regexp"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/parser"
)

var ErrAgentfileSecretLiteral = errors.New("agentfile layer contains a credential literal")

var sensitiveCredentialName = regexp.MustCompile(
	`(?i)(?:^|[_-])(?:api[_-]?key|access[_-]?token|auth[_-]?token|refresh[_-]?token|token|secret|client[_-]?secret|password|credential|private[_-]?key)$`,
)

func validateAgentfileLayerSecrets(layer string) error {
	program, parseErrors := parser.Parse(layer)
	if len(parseErrors) != 0 {
		return ErrInvalidAgentfileLayer
	}
	for _, declaration := range program.Declarations {
		switch value := declaration.(type) {
		case *parser.EnvDecl:
			if envContainsCredentialLiteral(value) {
				return ErrAgentfileSecretLiteral
			}
		case *parser.ConfigDecl:
			literal, _ := value.Default.(string)
			if literal != "" && (sensitiveCredentialName.MatchString(value.Name) || looksLikeCredential(literal)) {
				return ErrAgentfileSecretLiteral
			}
		}
	}
	return nil
}

func envContainsCredentialLiteral(declaration *parser.EnvDecl) bool {
	if declaration.Source != "" {
		return false
	}
	literal := declaration.Value
	if declaration.ValueExpr != nil {
		value, ok := declaration.ValueExpr.(*parser.StringLit)
		if !ok {
			return sensitiveCredentialName.MatchString(declaration.Name)
		}
		literal = value.Value
	}
	return sensitiveCredentialName.MatchString(declaration.Name) || looksLikeCredential(literal)
}

var (
	base64Credential      = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{32,}={0,2}|[A-Za-z0-9_-]{32,})$`)
	hexCredential         = regexp.MustCompile(`^[a-fA-F0-9]{32,}$`)
	jwtCredential         = regexp.MustCompile(`^eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)
	knownCredentialPrefix = regexp.MustCompile(`(?i)^(sk-[a-z0-9][a-z0-9_-]{16,}|sk-ant-[a-z0-9_-]{16,}|gh[pousr]_[a-z0-9_]{16,}|github_pat_[a-z0-9_]{20,}|glpat-[a-z0-9_-]{16,}|xox[abprs]-[a-z0-9-]{16,}|AKIA[0-9A-Z]{16}|ASIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{20,})`)
)

func looksLikeCredential(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 20 {
		return false
	}
	if knownCredentialPrefix.MatchString(value) || jwtCredential.MatchString(value) {
		return true
	}
	return len(value) >= 32 && (base64Credential.MatchString(value) || hexCredential.MatchString(value))
}
