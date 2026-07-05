package capability

import "strings"

// ScanDeclarations extracts CAPABILITY axis/value pairs from Agentfile source
// without running the full parser (list endpoints must stay sub-millisecond).
func ScanDeclarations(src string) map[string]string {
	if src == "" {
		return nil
	}
	caps := make(map[string]string)
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "CAPABILITY ") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		caps[parts[1]] = parts[2]
	}
	if len(caps) == 0 {
		return nil
	}
	return caps
}
