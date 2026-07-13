package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

// SkillInfo holds parsed information about a discovered skill.
type SkillInfo struct {
	Slug          string
	DisplayName   string
	Description   string
	License       string
	Compatibility string
	AllowedTools  string
	Category      string
	Tags          []string
	DirPath       string // absolute path to the skill directory
}

func detectRepoType(repoDir string) string {
	if fileExists(filepath.Join(repoDir, "SKILL.md")) {
		return "single"
	}
	return "collection"
}

func scanCollectionSkills(repoDir string) ([]SkillInfo, error) {
	var skills []SkillInfo

	skillsDir := filepath.Join(repoDir, "skills")
	if dirExists(skillsDir) {
		entries, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() || shouldIgnoreDir(entry.Name()) {
					continue
				}
				dirPath := filepath.Join(skillsDir, entry.Name())
				if fileExists(filepath.Join(dirPath, "SKILL.md")) {
					info, err := parseSkillDir(dirPath)
					if err != nil {
						slog.Warn("Failed to parse skill", "dir", dirPath, "error", err)
						continue
					}
					skills = append(skills, *info)
				}
			}
		}
		if len(skills) > 0 {
			return skills, nil
		}
	}

	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || shouldIgnoreDir(entry.Name()) {
			continue
		}
		dirPath := filepath.Join(repoDir, entry.Name())
		if fileExists(filepath.Join(dirPath, "SKILL.md")) {
			info, err := parseSkillDir(dirPath)
			if err != nil {
				slog.Warn("Failed to parse skill", "dir", dirPath, "error", err)
				continue
			}
			skills = append(skills, *info)
		}
	}

	return skills, nil
}

var ignoredDirs = map[string]bool{
	".git": true, ".github": true, ".vscode": true,
	"spec": true, "template": true, "templates": true, ".claude-plugin": true,
	"node_modules": true, "__pycache__": true, "vendor": true,
}

func shouldIgnoreDir(name string) bool {
	return strings.HasPrefix(name, ".") || ignoredDirs[name]
}

func shouldIgnoreFile(name string) bool {
	return strings.HasPrefix(name, "._")
}

func parseSkillDir(dirPath string) (*SkillInfo, error) {
	skillMdPath := filepath.Join(dirPath, "SKILL.md")
	content, err := os.ReadFile(skillMdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	fm := parseFrontmatter(string(content))
	tags, err := readSkillConfigTags(dirPath)
	if err != nil {
		return nil, err
	}

	slug := fm["name"]
	if slug == "" {
		slug = filepath.Base(dirPath)
	}

	return &SkillInfo{
		Slug:          slug,
		DisplayName:   fm["name"],
		Description:   fm["description"],
		License:       fm["license"],
		Compatibility: fm["compatibility"],
		AllowedTools:  fm["allowed-tools"],
		Category:      fm["category"],
		Tags:          tags,
		DirPath:       dirPath,
	}, nil
}

func readSkillConfigTags(dirPath string) ([]string, error) {
	path := filepath.Join(dirPath, "skill.json")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to inspect skill.json: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("skill.json must be a regular file")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill.json: %w", err)
	}
	var config struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse skill.json: %w", err)
	}
	return []string(skilldom.NormalizeTags(config.Tags)), nil
}

func parseFrontmatter(content string) map[string]string {
	fm := make(map[string]string)

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return fm
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"'`)
			fm[key] = value
		}
	}

	return fm
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
