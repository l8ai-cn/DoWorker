package workspace

import "fmt"

func NormalizeCommitSHA(sha string) (string, error) {
	if sha == "" {
		return "", nil
	}
	if len(sha) != 40 && len(sha) != 64 {
		return "", fmt.Errorf("commit sha must be lowercase 40 or 64 hex characters")
	}
	for _, c := range sha {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return "", fmt.Errorf("commit sha must be lowercase 40 or 64 hex characters")
	}
	return sha, nil
}

func RequireCommitSHA(field, sha string) (string, error) {
	normalized, err := NormalizeCommitSHA(sha)
	if err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}
	if normalized == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	return normalized, nil
}
