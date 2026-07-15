package goalloop

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode/utf8"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func verificationFingerprints(result *runnerv1.VerificationResultEvent) (string, string) {
	normalizedOutput := strings.Join(strings.Fields(result.GetOutput()), " ")
	progress := fingerprint(fmt.Sprintf(
		"%d\n%s\n%s", result.GetExitCode(), result.GetError(), normalizedOutput,
	))
	errorBasis := strings.TrimSpace(result.GetError())
	if errorBasis == "" {
		errorBasis = firstNonEmptyLine(result.GetOutput())
	}
	return progress, fingerprint(fmt.Sprintf("%d\n%s", result.GetExitCode(), errorBasis))
}

func consecutiveFingerprintCount(previous *string, current string, count int) int {
	if previous != nil && *previous == current {
		return count + 1
	}
	return 1
}

func fingerprint(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func firstNonEmptyLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

func buildVerificationRetryPrompt(
	iteration, maxIterations int,
	result *runnerv1.VerificationResultEvent,
) string {
	output := truncatePromptEvidence(result.GetOutput())
	return fmt.Sprintf(
		"GoalLoop verification failed. Continue the same objective in iteration %d of %d.\n"+
			"Verifier exit code: %d\nVerifier output:\n%s\n"+
			"Fix the verified failure, then stop and wait. External verification decides completion.",
		iteration, effectiveLimit(maxIterations, 10), result.GetExitCode(), output,
	)
}

func truncatePromptEvidence(output string) string {
	const maxBytes = 8 << 10
	end := min(len(output), maxBytes)
	for end > 0 && !utf8.ValidString(output[:end]) {
		end--
	}
	return output[:end]
}

func effectiveLimit(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
