package goalloop

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/loopscript"
	"github.com/anthropics/agentsmesh/backend/pkg/secretguard"
)

var (
	ErrDraftRepairUnsupported = errors.New("loop diagnostic repair is unsupported")
	ErrDraftRepairTargetStale = errors.New("loop diagnostic repair target is stale")
)

type DraftRepairInput struct {
	Source          string
	ModelResourceID int64
	Locale          string
	DiagnosticCode  string
	NodeID          string
	FieldPath       string
	Prompt          string
}

type DraftIntegerPatch struct {
	NodeID    string
	FieldPath string
	OldValue  int64
	NewValue  int64
}

type DraftRepairProposal struct {
	Program         *loopscript.Program
	CanonicalSource string
	Patch           DraftIntegerPatch
}

type loopRepairEnvelope struct {
	Value int64 `json:"value"`
}

func (generator *DraftGenerator) Repair(
	ctx context.Context,
	scope DraftGenerationScope,
	input DraftRepairInput,
) (DraftRepairProposal, error) {
	if generator == nil || generator.resources == nil || generator.provider == nil {
		return DraftRepairProposal{}, ErrDraftGenerationUnavailable
	}
	normalized, err := validateDraftRepairInput(scope, input)
	if err != nil {
		return DraftRepairProposal{}, err
	}
	program, diagnostics := loopscript.Analyze(normalized.Source)
	if program == nil {
		return DraftRepairProposal{}, ErrDraftRepairUnsupported
	}
	if hasSecretDiagnostic(diagnostics) {
		return DraftRepairProposal{}, ErrDraftContainsSecret
	}
	diagnostic, ok := exactRepairDiagnostic(diagnostics, normalized)
	if !ok {
		return DraftRepairProposal{}, ErrDraftRepairTargetStale
	}
	target, ok := integerRepairTarget(program, diagnostic)
	if !ok {
		return DraftRepairProposal{}, ErrDraftRepairUnsupported
	}
	resource, err := generator.resolveResource(ctx, scope, normalized.ModelResourceID)
	if err != nil {
		return DraftRepairProposal{}, err
	}
	systemPrompt, userPrompt := buildLoopRepairPrompts(normalized, target)
	raw, err := generator.provider.Generate(ctx, resource, systemPrompt, userPrompt)
	if err != nil {
		return DraftRepairProposal{}, ErrDraftProviderUnavailable
	}
	value, err := decodeLoopRepair(raw)
	if err != nil || value < target.minimum || value > target.maximum ||
		value == target.current {
		return DraftRepairProposal{}, fmt.Errorf("%w: invalid repair value", ErrGeneratedDraftInvalid)
	}
	updated := *program
	target.apply(&updated, value)
	canonical, formatDiagnostics := loopscript.Format(&updated)
	if len(formatDiagnostics) != 0 {
		return DraftRepairProposal{}, fmt.Errorf("%w: repair did not compile", ErrGeneratedDraftInvalid)
	}
	if _, compileDiagnostics := loopscript.CompileGoalLoopV1(&updated); len(compileDiagnostics) != 0 {
		return DraftRepairProposal{}, fmt.Errorf("%w: repair did not compile", ErrGeneratedDraftInvalid)
	}
	return DraftRepairProposal{
		Program:         &updated,
		CanonicalSource: canonical,
		Patch: DraftIntegerPatch{
			NodeID: normalized.NodeID, FieldPath: normalized.FieldPath,
			OldValue: target.current, NewValue: value,
		},
	}, nil
}

func validateDraftRepairInput(
	scope DraftGenerationScope,
	input DraftRepairInput,
) (DraftRepairInput, error) {
	input.Source = strings.TrimSpace(input.Source)
	input.Locale = strings.TrimSpace(input.Locale)
	input.DiagnosticCode = strings.TrimSpace(input.DiagnosticCode)
	input.NodeID = strings.TrimSpace(input.NodeID)
	input.FieldPath = strings.TrimSpace(input.FieldPath)
	input.Prompt = strings.TrimSpace(input.Prompt)
	if scope.OrganizationID <= 0 || scope.UserID <= 0 || input.ModelResourceID <= 0 ||
		input.Source == "" || input.Locale == "" || input.DiagnosticCode == "" ||
		input.NodeID == "" || input.FieldPath == "" {
		return DraftRepairInput{}, ErrInvalidDraftGenerationInput
	}
	if len(input.Source) > maxDraftSourceBytes || len(input.Locale) > maxDraftLocaleBytes ||
		len(input.Prompt) > maxDraftPromptBytes || len(input.DiagnosticCode) > 128 ||
		len(input.NodeID) > 128 || len(input.FieldPath) > 128 {
		return DraftRepairInput{}, ErrInvalidDraftGenerationInput
	}
	if secretguard.ContainsCredentialLiteral(input.Source) ||
		secretguard.ContainsCredentialLiteral(input.Prompt) {
		return DraftRepairInput{}, ErrDraftContainsSecret
	}
	return input, nil
}

func exactRepairDiagnostic(
	diagnostics []loopscript.Diagnostic,
	input DraftRepairInput,
) (loopscript.Diagnostic, bool) {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == input.DiagnosticCode &&
			diagnostic.NodeID == input.NodeID &&
			diagnostic.FieldPath == input.FieldPath {
			return diagnostic, true
		}
	}
	return loopscript.Diagnostic{}, false
}

func hasSecretDiagnostic(diagnostics []loopscript.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "loop.secret.literal-forbidden" {
			return true
		}
	}
	return false
}

func buildLoopRepairPrompts(
	input DraftRepairInput,
	target draftIntegerTarget,
) (string, string) {
	system := strings.TrimSpace(`
Choose one integer value that repairs one validated LoopScript diagnostic.
Return exactly one JSON object with exactly one integer field named "value".
The value must stay inside the supplied inclusive range.
Do not return source code, Markdown, explanations, extra fields, or secrets.
`)
	user := fmt.Sprintf(
		"Interface locale: %s\nDiagnostic: %s\nNode: %s\nField: %s\nCurrent: %d\nAllowed: %d..%d\nUser intent: %s",
		input.Locale, input.DiagnosticCode, input.NodeID, input.FieldPath,
		target.current, target.minimum, target.maximum, input.Prompt,
	)
	return system, user
}

func decodeLoopRepair(raw []byte) (int64, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var envelope loopRepairEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return 0, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return 0, errors.New("trailing JSON")
	}
	return envelope.Value, nil
}
