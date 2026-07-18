package workercreation

import "strings"

func digestFromSHA(value string) string {
	if strings.HasPrefix(value, "sha256:") {
		return value
	}
	return "sha256:" + value
}
