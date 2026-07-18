package workerdependency

import (
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/parser"
)

func validateAgentfileSyntax(source string) error {
	if containsRawSecretText(source) {
		return fmt.Errorf("worker dependency agentfile source contains raw secret-like data")
	}
	_, parseErrors := parser.Parse(source)
	if len(parseErrors) != 0 {
		return fmt.Errorf("worker dependency agentfile source is invalid")
	}
	return nil
}

func validateAgentfileFieldOwnership(document Document) error {
	program, parseErrors := parser.Parse(document.Worker.AgentfileSource)
	if len(parseErrors) != 0 {
		return fmt.Errorf("worker dependency agentfile source is invalid")
	}
	protected := modelManagedEnvironmentFields(document)
	for _, field := range document.Worker.CredentialBundleFields {
		protected[field] = struct{}{}
	}
	for _, declaration := range program.Declarations {
		switch value := declaration.(type) {
		case *parser.EnvDecl:
			_, managed := protected[value.Name]
			if (managed || isSensitiveFieldName(value.Name)) &&
				value.Source == "" {
				return fmt.Errorf(
					"agentfile field %q must use a live reference",
					value.Name,
				)
			}
		case *parser.ConfigDecl:
			_, managed := protected[value.Name]
			if (managed || isSensitiveFieldName(value.Name)) &&
				!emptyAgentfileDefault(value.Default) {
				return fmt.Errorf(
					"agentfile field %q must use a live reference",
					value.Name,
				)
			}
		}
	}
	return nil
}

func emptyAgentfileDefault(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}
