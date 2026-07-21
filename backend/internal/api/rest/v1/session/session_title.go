package sessionapi

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
)

var attachmentMarkerLineRE = regexp.MustCompile(`(?m)^\[(?:Attached(?: file)?):[^\]]*\]\s*$`)

func sessionTitleEmpty(title *string) bool {
	return title == nil || strings.TrimSpace(*title) == ""
}

func deriveSessionTitleFromPrompt(prompt string) *string {
	s := strings.TrimSpace(prompt)
	if s == "" {
		return nil
	}
	s = attachmentMarkerLineRE.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ">") {
			continue
		}
		s = trimmed
		break
	}
	if s == "" {
		return nil
	}
	s = truncateSessionTitle(s, 60)
	if s == "" {
		return nil
	}
	return &s
}

func truncateSessionTitle(raw string, max int) string {
	if max <= 1 || utf8.RuneCountInString(raw) <= max {
		return strings.TrimSpace(raw)
	}
	runes := []rune(raw)
	slice := runes[:max-1]
	lastSpace := -1
	for i, r := range slice {
		if r == ' ' {
			lastSpace = i
		}
	}
	cut := len(slice)
	if lastSpace > max-10 {
		cut = lastSpace
	}
	return strings.TrimSpace(string(slice[:cut])) + "…"
}

func (d *Deps) maybeSeedSessionTitle(ctx context.Context, row *domain.Session, prompt string) {
	if d == nil || d.Sessions == nil || row == nil || !sessionTitleEmpty(row.Title) {
		return
	}
	title := deriveSessionTitleFromPrompt(prompt)
	if title == nil {
		return
	}
	if err := d.Sessions.UpdateTitle(ctx, row.ID, title); err != nil {
		return
	}
	row.Title = title
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
}
