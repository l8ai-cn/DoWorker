package doagent

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
	"github.com/l8ai-cn/agentcloud/runner/internal/tokenusage"
)

type doagentParser struct{}

type doagentLogEntry struct {
	Event string `json:"event"`
	Model string `json:"model"`
	Usage *struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	} `json:"usage"`
}

func (p *doagentParser) Parse(sandboxPath string, podStartedAt time.Time) (*tokenusage.TokenUsage, error) {
	usage := tokenusage.NewTokenUsage()

	for _, logsDir := range doagentLogDirs(sandboxPath) {
		if _, err := os.Stat(logsDir); os.IsNotExist(err) {
			continue
		}
		parseDoAgentLogsDir(logsDir, podStartedAt, usage)
	}

	if usage.IsEmpty() {
		return nil, nil
	}
	return usage, nil
}

func doagentLogDirs(sandboxPath string) []string {
	if sandboxPath != "" {
		return []string{filepath.Join(sandboxPath, "do-agent-home", "logs")}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return []string{filepath.Join(home, ".agent", "logs")}
	}
	return nil
}

func parseDoAgentLogsDir(logsDir string, podStartedAt time.Time, usage *tokenusage.TokenUsage) {
	err := filepath.WalkDir(logsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		if !tokenusage.IsModifiedAfter(path, podStartedAt) {
			return nil
		}
		if parseErr := parseDoAgentLogFile(path, usage); parseErr != nil {
			logger.Pod().Warn("DoAgent parser: file parse error", "file", path, "error", parseErr)
		}
		return nil
	})
	if err != nil {
		logger.Pod().Warn("DoAgent parser: walk error", "dir", logsDir, "error", err)
	}
}

func parseDoAgentLogFile(path string, usage *tokenusage.TokenUsage) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry doagentLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.Event != "llm_response" || entry.Usage == nil {
			continue
		}

		in := entry.Usage.PromptTokens
		out := entry.Usage.CompletionTokens
		if in <= 0 && out <= 0 {
			continue
		}

		model := entry.Model
		if model == "" {
			model = "do-agent-unknown"
		}
		usage.Add(model, in, out, 0, 0)
	}

	return scanner.Err()
}
