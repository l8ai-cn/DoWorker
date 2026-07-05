package runner

import "strings"

const resumeExternalSessionEnvKey = "AGENTSMESH_RESUME_EXTERNAL_SESSION"

func resumeExternalSessionFromEnv(env []string) string {
	prefix := resumeExternalSessionEnvKey + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return e[len(prefix):]
		}
	}
	return ""
}
