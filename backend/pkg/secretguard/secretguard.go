package secretguard

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	sensitiveName = regexp.MustCompile(
		`(?i)(?:^|[_-])(?:api[_-]?key|access[_-]?token|auth[_-]?token|refresh[_-]?token|token|secret|client[_-]?secret|password|passwd|passphrase|pwd|credential|private[_-]?key)$`,
	)
	knownCredential = regexp.MustCompile(
		`(?i)(?:^|[^a-z0-9_])(?:sk-ant-[a-z0-9_-]{16,}|sk-[a-z0-9][a-z0-9_-]{16,}|gh[pousr]_[a-z0-9_]{16,}|github_pat_[a-z0-9_]{20,}|glpat-[a-z0-9_-]{16,}|xox[abprs]-[a-z0-9-]{16,}|AKIA[0-9A-Z]{16}|ASIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{20,})`,
	)
	jwtCredential       = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	bearerCredential    = regexp.MustCompile(`(?i)\bbearer[ \t]+[A-Za-z0-9._~+/=-]{8,}`)
	privateKeyPEM       = regexp.MustCompile(`(?i)-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY-----`)
	sensitiveAssignment = regexp.MustCompile(
		`(?i)(?:^|[\s;,])(?:api[_-]?key|access[_-]?token|auth[_-]?token|refresh[_-]?token|token|secret|client[_-]?secret|password|passwd|passphrase|pwd|credential|private[_-]?key)\s*[:=]\s*["']?[^\s"']{4,}`,
	)
	connectionCredential = regexp.MustCompile(
		`(?i)\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis|amqp|https?)://[^/\s:@]+:[^@\s/]+@`,
	)
	base64Credential = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{32,}={0,2}|[A-Za-z0-9_-]{32,})$`)
	hexCredential    = regexp.MustCompile(`^[a-fA-F0-9]{32,}$`)
)

func IsSensitiveName(name string) bool {
	return sensitiveName.MatchString(name)
}

func ContainsCredentialLiteral(value string) bool {
	if knownCredential.MatchString(value) ||
		jwtCredential.MatchString(value) ||
		bearerCredential.MatchString(value) ||
		privateKeyPEM.MatchString(value) ||
		sensitiveAssignment.MatchString(value) ||
		connectionCredential.MatchString(value) {
		return true
	}

	for _, token := range strings.FieldsFunc(value, isOpaqueCredentialSeparator) {
		if len(token) >= 32 &&
			(hexCredential.MatchString(token) ||
				base64Credential.MatchString(token) && hasMixedBase64Alphabet(token)) {
			return true
		}
	}
	return false
}

func hasMixedBase64Alphabet(value string) bool {
	var upper, lower bool
	for _, r := range value {
		upper = upper || unicode.IsUpper(r)
		lower = lower || unicode.IsLower(r)
	}
	return upper && lower
}

func isOpaqueCredentialSeparator(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
		r != '+' && r != '/' && r != '_' && r != '-' && r != '='
}
