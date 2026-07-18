package skill

import (
	"encoding/json"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

const skillConfigSchema = 2

type CreateSkillRequest struct {
	OrganizationID int64
	UserID         int64
	Slug           string
	Name           string
	Description    string
	License        string
	Instructions   string
	Tags           []string
}

type EnsurePlatformSkillRequest struct {
	UserID       int64
	Slug         string
	Name         string
	Description  string
	License      string
	Instructions string
	Tags         []string
}

type UpdateSkillRequest struct {
	OrganizationID int64
	SkillID        int64
	Name           *string
	Description    *string
	License        *string
	Instructions   *string
	Tags           *[]string
}

type skillConfig struct {
	Schema      int      `json:"schema"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	License     string   `json:"license,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func renderSkillFiles(slug, displayName, description, license, body string, tags []string) ([]gitops.FileChange, error) {
	var md strings.Builder
	md.WriteString("---\n")
	fmt.Fprintf(&md, "name: %s\n", slug)
	if d := strings.TrimSpace(description); d != "" {
		fmt.Fprintf(&md, "description: %s\n", sanitizeFrontmatterValue(d))
	}
	if l := strings.TrimSpace(license); l != "" {
		fmt.Fprintf(&md, "license: %s\n", sanitizeFrontmatterValue(l))
	}
	md.WriteString("---\n\n")
	md.WriteString(strings.TrimRight(body, "\n"))
	md.WriteString("\n")

	cfg := skillConfig{
		Schema:      skillConfigSchema,
		Slug:        slug,
		Name:        strings.TrimSpace(displayName),
		Description: strings.TrimSpace(description),
		License:     strings.TrimSpace(license),
		Tags:        []string(skilldom.NormalizeTags(tags)),
	}
	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("skill: render skill.json: %w", err)
	}

	return []gitops.FileChange{
		{Path: "SKILL.md", Content: []byte(md.String())},
		{Path: "skill.json", Content: append(cfgJSON, '\n')},
	}, nil
}

func sanitizeFrontmatterValue(v string) string {
	v = strings.ReplaceAll(v, "\r", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	return strings.TrimSpace(v)
}

func extractSkillBody(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n")
		}
	}
	return content
}
