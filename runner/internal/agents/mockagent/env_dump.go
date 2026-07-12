package mockagent

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// envDumpPrefixes matches the prefix set the e2e env-bundle regression relies on.
// Keep in sync with backend/migrations/000150_e2e_echo_print_env_on_startup.
var envDumpPrefixes = []string{"E2E_TEST_", "ANTHROPIC_", "CLAUDE_"}

const envDumpReadyMarker = "E2E_TEST_ENV_DUMP_READY=1"

// writeEnvDump replicates the bash `env | grep -E ... > /tmp/e2e-echo-env-dump-$$`
// the legacy e2e-echo agent ran at startup, so docker-exec-based env-bundle
// tests can keep asserting on the same file path. Failure is non-fatal: the
// dump is best-effort and the runtime continues even if /tmp is read-only.
func writeEnvDump(env []string) {
	pid := os.Getpid()
	path := fmt.Sprintf("/tmp/e2e-echo-env-dump-%d", pid)

	if err := os.WriteFile(path, []byte(envDumpContent(env)), 0o644); err != nil {
		// Stay silent in PTY (printing to stdout would pollute the echo
		// protocol consumers test against). The dump's absence is itself
		// the failure signal the env-bundle e2e watches for.
		return
	}
}

func envDumpContent(env []string) string {
	matched := append(filterEnvByPrefix(env, envDumpPrefixes), envDumpReadyMarker)
	sort.Strings(matched)
	return strings.Join(matched, "\n") + "\n"
}

func filterEnvByPrefix(env, prefixes []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		for _, p := range prefixes {
			if strings.HasPrefix(kv, p) {
				out = append(out, kv)
				break
			}
		}
	}
	return out
}
