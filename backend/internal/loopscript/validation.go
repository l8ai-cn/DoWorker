package loopscript

import (
	"fmt"
	"regexp"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var customBlockDigestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type nodeDescriptor struct {
	nodeID   string
	localID  string
	position sourcePosition
}

func validateProgram(
	program *Program,
	positions *programPositions,
	redactions textRedactions,
) []Diagnostic {
	if program == nil {
		return []Diagnostic{newDiagnostic("loop.program.nil", "program is required", "", sourcePosition{})}
	}
	var diagnostics []Diagnostic
	if program.SchemaVersion != 1 {
		diagnostics = append(diagnostics, diagnosticFor(
			"loop.schema-version.unsupported", "schema version must be 1", program.Loop.NodeID, loopPosition(positions),
		))
	}

	nodes := programNodes(program, positions)
	seenNodeIDs := make(map[string]struct{}, len(nodes))
	seenLocalIDs := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if node.nodeID == "" {
			diagnostics = append(diagnostics, diagnosticFor(
				"loop.node-id.missing", "every node requires @id(...)", "", node.position,
			))
		} else {
			diagnostics = appendInvalidIdentifier(diagnostics, node.nodeID, node.nodeID, node.position)
			if _, exists := seenNodeIDs[node.nodeID]; exists {
				diagnostics = append(diagnostics, diagnosticFor(
					"loop.node-id.duplicate", fmt.Sprintf("duplicate node id %q", node.nodeID), node.nodeID, node.position,
				))
			}
			seenNodeIDs[node.nodeID] = struct{}{}
		}

		diagnostics = appendInvalidIdentifier(diagnostics, node.localID, node.nodeID, node.position)
		if _, exists := seenLocalIDs[node.localID]; exists {
			diagnostics = append(diagnostics, diagnosticFor(
				"loop.local-id.duplicate", fmt.Sprintf("duplicate local id %q", node.localID), node.nodeID, node.position,
			))
		}
		seenLocalIDs[node.localID] = struct{}{}
	}
	diagnostics = appendCustomBlockDiagnostics(diagnostics, program.Loop.Repeat.CustomBlock, seenNodeIDs, positions)

	loop := program.Loop
	diagnostics = appendRangeDiagnostics(diagnostics, loop, positions)
	if loop.Repeat.Until.LocalID != loop.Repeat.Verifier.LocalID || loop.Repeat.Until.Field != "passed" {
		diagnostics = append(diagnostics, diagnosticFor(
			"loop.reference.until-invalid",
			"repeat until must equal <verify-local-id>.passed",
			loop.Repeat.NodeID, repeatPosition(positions),
		))
	}
	if loop.Repeat.Max > loop.Limits.Iterations {
		diagnostics = append(diagnostics, diagnosticForField(
			"loop.repeat.max-exceeds-limit",
			"repeat max must not exceed limits.iterations",
			loop.Repeat.NodeID, "repeat.max", repeatPosition(positions),
		))
	}
	if loop.FailurePolicy != FailurePause && loop.FailurePolicy != FailureFail {
		diagnostics = append(diagnostics, diagnosticFor(
			"loop.failure-policy.invalid", "failure policy must be pause or fail", loop.NodeID, failurePosition(positions),
		))
	}
	diagnostics = appendTextDiagnostics(diagnostics, loop, positions, redactions)
	return diagnostics
}

func appendCustomBlockDiagnostics(
	diagnostics []Diagnostic,
	custom *CustomBlockRef,
	seenNodeIDs map[string]struct{},
	positions *programPositions,
) []Diagnostic {
	if custom == nil {
		return diagnostics
	}
	position := customBlockPosition(positions)
	if custom.NodeID == "" {
		diagnostics = append(diagnostics, diagnosticFor(
			"loop.custom-block.node-id.required",
			"custom block node id is required",
			"",
			position,
		))
	} else {
		diagnostics = appendInvalidIdentifier(diagnostics, custom.NodeID, custom.NodeID, position)
		if _, exists := seenNodeIDs[custom.NodeID]; exists {
			diagnostics = append(diagnostics, diagnosticFor(
				"loop.node-id.duplicate",
				fmt.Sprintf("duplicate node id %q", custom.NodeID),
				custom.NodeID,
				position,
			))
		}
		seenNodeIDs[custom.NodeID] = struct{}{}
	}
	if custom.DefinitionID == "" {
		diagnostics = append(diagnostics, diagnosticFor(
			"loop.custom-block.definition-id.required",
			"custom block definition id is required",
			custom.NodeID,
			position,
		))
	}
	diagnostics = appendInvalidIdentifier(diagnostics, custom.Slug, custom.NodeID, position)
	if custom.Version < 1 {
		diagnostics = append(diagnostics, diagnosticForField(
			"loop.custom-block.version.invalid",
			"custom block version must be positive",
			custom.NodeID,
			"repeat.custom_block.version",
			position,
		))
	}
	if !customBlockDigestPattern.MatchString(custom.DefinitionDigest) {
		diagnostics = append(diagnostics, diagnosticForField(
			"loop.custom-block.digest.invalid",
			"custom block definition digest must be a lowercase SHA-256 hex value",
			custom.NodeID,
			"repeat.custom_block.definition_digest",
			position,
		))
	}
	return diagnostics
}

func programNodes(program *Program, positions *programPositions) []nodeDescriptor {
	loop := program.Loop
	return []nodeDescriptor{
		{loop.NodeID, loop.LocalID, loopPosition(positions)},
		{loop.Repeat.NodeID, loop.Repeat.LocalID, repeatPosition(positions)},
		{loop.Repeat.Agent.NodeID, loop.Repeat.Agent.LocalID, agentPosition(positions)},
		{loop.Repeat.Verifier.NodeID, loop.Repeat.Verifier.LocalID, verifierPosition(positions)},
	}
}

func appendInvalidIdentifier(diagnostics []Diagnostic, value, nodeID string, position sourcePosition) []Diagnostic {
	if err := slugkit.Validate(value); err != nil {
		return append(diagnostics, diagnosticFor(
			"loop.identifier.invalid",
			fmt.Sprintf("identifier %q is invalid: %v", value, err),
			nodeID, position,
		))
	}
	return diagnostics
}

func diagnosticFor(code, message, nodeID string, position sourcePosition) Diagnostic {
	return newDiagnostic(code, message, nodeID, position)
}

func diagnosticForField(
	code, message, nodeID, fieldPath string,
	position sourcePosition,
) Diagnostic {
	diagnostic := diagnosticFor(code, message, nodeID, position)
	diagnostic.FieldPath = fieldPath
	return diagnostic
}
