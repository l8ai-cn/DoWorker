package agentpod

import (
	"errors"

	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/backend/pkg/secretguard"
)

var ErrAgentfileSecretLiteral = errors.New("agentfile layer contains a credential literal")

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
			if literal != "" &&
				(secretguard.IsSensitiveName(value.Name) || secretguard.ContainsCredentialLiteral(literal)) {
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
			return secretguard.IsSensitiveName(declaration.Name)
		}
		literal = value.Value
	}
	return secretguard.IsSensitiveName(declaration.Name) ||
		secretguard.ContainsCredentialLiteral(literal)
}
