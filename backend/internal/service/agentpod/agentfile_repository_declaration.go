package agentpod

import (
	"errors"

	"github.com/anthropics/agentsmesh/agentfile/parser"
)

type repositoryDeclaration struct {
	slug    string
	present bool
}

func parseRepositoryDeclaration(agentfileSrc string) (repositoryDeclaration, error) {
	prog, parseErrors := parser.Parse(agentfileSrc)
	if len(parseErrors) > 0 {
		return repositoryDeclaration{}, errors.New(parseErrors[0])
	}

	result := repositoryDeclaration{}
	for _, declaration := range prog.Declarations {
		repo, ok := declaration.(*parser.RepoDecl)
		if !ok {
			continue
		}
		result.present = true
		result.slug = ""
		if value, ok := repo.Value.(*parser.StringLit); ok {
			result.slug = value.Value
		}
	}
	return result, nil
}
