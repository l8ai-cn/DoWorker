package goalloop

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/loopscript"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/secretguard"
)

const (
	maxDraftPromptBytes = 16 << 10
	maxDraftSourceBytes = 128 << 10
	maxDraftLocaleBytes = 32
)

var (
	ErrDraftGenerationUnavailable  = errors.New("loop draft generation is unavailable")
	ErrInvalidDraftGenerationInput = errors.New("invalid Loop draft generation input")
	ErrDraftContainsSecret         = errors.New("loop AI input contains secret-like data")
	ErrDraftSourceInvalid          = errors.New("current Loop source is invalid")
	ErrGeneratedDraftInvalid       = errors.New("generated Loop draft is invalid")
	ErrDraftProviderUnavailable    = errors.New("loop draft provider is unavailable")
)

var supportedDraftAdapters = []string{
	"openai-compatible",
	"anthropic",
	"gemini",
}

type DraftResourceResolver interface {
	ResolveExact(
		context.Context,
		airesourceservice.Actor,
		int64,
		int64,
		airesourceservice.ResolutionRequirements,
	) (*airesourceservice.ResolvedResource, error)
}

type DraftJSONGenerator interface {
	Generate(
		context.Context,
		*airesourceservice.ResolvedResource,
		string,
		string,
	) ([]byte, error)
}

type DraftGenerator struct {
	resources DraftResourceResolver
	provider  DraftJSONGenerator
}

type DraftGenerationScope struct {
	OrganizationID int64
	UserID         int64
}

type DraftGenerationInput struct {
	Prompt          string
	CurrentSource   string
	ModelResourceID int64
	Locale          string
}

type DraftProposal struct {
	Program         *loopscript.Program
	CanonicalSource string
}

func NewDraftGenerator(
	resources DraftResourceResolver,
	provider DraftJSONGenerator,
) *DraftGenerator {
	return &DraftGenerator{resources: resources, provider: provider}
}

func (generator *DraftGenerator) Generate(
	ctx context.Context,
	scope DraftGenerationScope,
	input DraftGenerationInput,
) (DraftProposal, error) {
	if generator == nil || generator.resources == nil || generator.provider == nil {
		return DraftProposal{}, ErrDraftGenerationUnavailable
	}
	normalized, err := validateDraftGenerationInput(scope, input)
	if err != nil {
		return DraftProposal{}, err
	}
	currentProgram, err := parseCurrentDraft(normalized.CurrentSource)
	if err != nil {
		return DraftProposal{}, err
	}
	resource, err := generator.resolveResource(ctx, scope, normalized.ModelResourceID)
	if err != nil {
		return DraftProposal{}, err
	}
	systemPrompt, userPrompt := buildLoopGenerationPrompts(normalized)
	raw, err := generator.provider.Generate(ctx, resource, systemPrompt, userPrompt)
	if err != nil {
		return DraftProposal{}, ErrDraftProviderUnavailable
	}
	source, err := decodeLoopGeneration(raw)
	if err != nil {
		return DraftProposal{}, fmt.Errorf("%w: response envelope", ErrGeneratedDraftInvalid)
	}
	proposal, err := compileDraftProposal(source)
	if err != nil {
		return DraftProposal{}, err
	}
	if err := enforceDraftMutationPolicy(currentProgram, proposal.Program); err != nil {
		return DraftProposal{}, err
	}
	return proposal, nil
}

func validateDraftGenerationInput(
	scope DraftGenerationScope,
	input DraftGenerationInput,
) (DraftGenerationInput, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.CurrentSource = strings.TrimSpace(input.CurrentSource)
	input.Locale = strings.TrimSpace(input.Locale)
	if scope.OrganizationID <= 0 || scope.UserID <= 0 ||
		input.ModelResourceID <= 0 || input.Prompt == "" || input.Locale == "" {
		return DraftGenerationInput{}, ErrInvalidDraftGenerationInput
	}
	if len(input.Prompt) > maxDraftPromptBytes ||
		len(input.CurrentSource) > maxDraftSourceBytes ||
		len(input.Locale) > maxDraftLocaleBytes {
		return DraftGenerationInput{}, ErrInvalidDraftGenerationInput
	}
	if secretguard.ContainsCredentialLiteral(input.Prompt) {
		return DraftGenerationInput{}, ErrDraftContainsSecret
	}
	return input, nil
}

func parseCurrentDraft(source string) (*loopscript.Program, error) {
	if source == "" {
		return nil, nil
	}
	program, diagnostics := loopscript.Parse(source)
	if len(diagnostics) != 0 || program == nil {
		for _, diagnostic := range diagnostics {
			if diagnostic.Code == "loop.secret.literal-forbidden" {
				return nil, ErrDraftContainsSecret
			}
		}
		return nil, ErrDraftSourceInvalid
	}
	return program, nil
}

func compileDraftProposal(source string) (DraftProposal, error) {
	program, diagnostics := loopscript.Parse(source)
	if len(diagnostics) != 0 {
		return DraftProposal{}, generatedDraftDiagnostic(diagnostics[0])
	}
	canonical, diagnostics := loopscript.Format(program)
	if len(diagnostics) != 0 {
		return DraftProposal{}, generatedDraftDiagnostic(diagnostics[0])
	}
	if _, diagnostics = loopscript.CompileGoalLoopV1(program); len(diagnostics) != 0 {
		return DraftProposal{}, generatedDraftDiagnostic(diagnostics[0])
	}
	return DraftProposal{Program: program, CanonicalSource: canonical}, nil
}

func generatedDraftDiagnostic(diagnostic loopscript.Diagnostic) error {
	return fmt.Errorf("%w: %s", ErrGeneratedDraftInvalid, diagnostic.Code)
}
