package codex

import "strings"

type codexInputAdapter struct{}

const (
	bracketedPasteStart = "\x1b[200~"
	bracketedPasteEnd   = "\x1b[201~"
)

func (a *codexInputAdapter) Adapt(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	endsWithEnter := data[len(data)-1] == '\r' || data[len(data)-1] == '\n'

	s := string(data)
	if strings.HasSuffix(s, "\r\n") {
		s = strings.TrimSuffix(s, "\r\n")
	} else if strings.HasSuffix(s, "\r") || strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	adapted := []byte(bracketedPasteStart + s + bracketedPasteEnd)
	if endsWithEnter {
		adapted = append(adapted, '\r')
	}
	return adapted
}
