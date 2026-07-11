package workercreation

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/agentfile/serialize"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type compilationReferenceResolver interface {
	ResolveCompilationReferences(
		context.Context,
		specservice.Scope,
		slugkit.Slug,
		specdomain.Workspace,
		map[string]specdomain.SecretReference,
	) (compilationReferences, error)
}

type compiler struct {
	references compilationReferenceResolver
}

func newCompiler(references compilationReferenceResolver) *compiler {
	return &compiler{references: references}
}

func (compiler *compiler) Compile(
	ctx context.Context,
	scope specservice.Scope,
	spec specdomain.Spec,
) (string, error) {
	if compiler == nil || compiler.references == nil {
		return "", specservice.ErrResolverUnavailable
	}
	normalized, err := specdomain.NormalizeAndValidate(spec)
	if err != nil {
		return "", fmt.Errorf("%w: compile workerspec: %v", specservice.ErrInvalidDraft, err)
	}
	references, err := compiler.references.ResolveCompilationReferences(
		ctx,
		scope,
		normalized.Runtime.WorkerType.Slug,
		normalized.Workspace,
		normalized.TypeConfig.SecretRefs,
	)
	if err != nil {
		return "", err
	}
	program := &parser.Program{
		Declarations: compileDeclarations(normalized, references),
	}
	layer := serialize.Serialize(program)
	if _, parseErrors := parser.Parse(layer); len(parseErrors) > 0 {
		return "", fmt.Errorf("compile workerspec AgentFile: %s", parseErrors[0])
	}
	return layer, nil
}

func compileDeclarations(
	spec specdomain.Spec,
	references compilationReferences,
) []parser.Declaration {
	declarations := make([]parser.Declaration, 0, len(spec.TypeConfig.Values)+8)
	fields := make([]string, 0, len(spec.TypeConfig.Values))
	for field := range spec.TypeConfig.Values {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		declarations = append(declarations, &parser.ConfigDecl{
			Name:    field,
			Default: spec.TypeConfig.Values[field],
		})
	}
	declarations = append(declarations, &parser.ModeDecl{
		Mode: string(spec.TypeConfig.InteractionMode),
	})
	if references.RepositorySlug != "" {
		declarations = append(declarations,
			&parser.RepoDecl{Value: &parser.StringLit{Value: references.RepositorySlug}},
			&parser.BranchDecl{Value: &parser.StringLit{Value: spec.Workspace.Branch}},
		)
	}
	if len(references.SkillSlugs) > 0 {
		declarations = append(declarations, &parser.SkillsDecl{
			Slugs: append([]string{}, references.SkillSlugs...),
		})
	}
	if len(references.Knowledge) > 0 {
		mounts := make([]parser.KnowledgeMountRef, 0, len(references.Knowledge))
		for _, reference := range references.Knowledge {
			mode := ""
			if reference.Mode == specdomain.KnowledgeMountReadWrite {
				mode = string(reference.Mode)
			}
			mounts = append(mounts, parser.KnowledgeMountRef{
				Slug: reference.Slug,
				Mode: mode,
			})
		}
		declarations = append(declarations, &parser.KnowledgeDecl{Mounts: mounts})
	}
	for _, name := range references.EnvBundleNames {
		declarations = append(declarations, &parser.UseEnvBundleDecl{Name: name})
	}
	if prompt := compilePrompt(spec.Workspace); prompt != "" {
		declarations = append(declarations, &parser.PromptDecl{Content: prompt})
	}
	return declarations
}

func compilePrompt(workspace specdomain.Workspace) string {
	parts := make([]string, 0, 2)
	if instructions := strings.TrimSpace(workspace.Instructions); instructions != "" {
		parts = append(parts, instructions)
	}
	if task := strings.TrimSpace(workspace.InitialTask); task != "" {
		parts = append(parts, task)
	}
	return strings.Join(parts, "\n\n")
}
