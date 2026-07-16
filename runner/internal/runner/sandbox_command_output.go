package runner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type boundedCommandBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (buffer *boundedCommandBuffer) Write(data []byte) (int, error) {
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining > 0 {
		writeBytes := len(data)
		if writeBytes > remaining {
			writeBytes = remaining
		}
		_, _ = buffer.buffer.Write(data[:writeBytes])
	}
	if len(data) > remaining {
		buffer.exceeded = true
	}
	return len(data), nil
}

func runBoundedCommand(command *exec.Cmd, outputLimit int) (string, error) {
	stdout := &boundedCommandBuffer{limit: outputLimit}
	stderr := &boundedCommandBuffer{limit: outputLimit}
	command.Stdout = stdout
	command.Stderr = stderr
	runErr := command.Run()
	if stdout.exceeded {
		return "", fmt.Errorf("git output exceeds %d byte limit", outputLimit)
	}
	if stderr.exceeded {
		return "", fmt.Errorf("git error output exceeds %d byte limit", outputLimit)
	}
	if runErr != nil {
		message := strings.TrimSpace(stderr.buffer.String())
		if message == "" {
			return "", runErr
		}
		return "", fmt.Errorf("%w: %s", runErr, message)
	}
	return stdout.buffer.String(), nil
}
