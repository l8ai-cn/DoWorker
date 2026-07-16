package acp

import (
	"bufio"
	"io"
	"strings"
)

func (c *ACPClient) readStderr(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		level, message := classifyStderrLine(line)
		c.addLog(LogEntry{Level: level, Message: message})
		if c.cfg.Callbacks.OnLog != nil {
			c.cfg.Callbacks.OnLog(level, message)
		}
	}
}

func classifyStderrLine(line string) (level, message string) {
	trimmed := strings.TrimLeft(line, " \t")
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "error:"):
		return "error", strings.TrimSpace(trimmed[len("error:"):])
	case strings.HasPrefix(lower, "warn:"), strings.HasPrefix(lower, "warning:"):
		idx := strings.IndexByte(trimmed, ':')
		return "warn", strings.TrimSpace(trimmed[idx+1:])
	}
	return "stderr", line
}

func (c *ACPClient) addLog(entry LogEntry) {
	if entry.Level != "warn" && entry.Level != "error" {
		return
	}
	c.logsMu.Lock()
	defer c.logsMu.Unlock()
	c.logs = append(c.logs, entry)
	if len(c.logs) > c.maxLogs {
		c.logs = c.logs[len(c.logs)-c.maxLogs:]
	}
}
