package goalloop

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type loopGenerationEnvelope struct {
	Source string `json:"source"`
}

func decodeLoopGeneration(raw []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	var envelope loopGenerationEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return "", err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", errors.New("trailing JSON")
		}
		return "", err
	}
	source := strings.TrimSpace(envelope.Source)
	if source == "" {
		return "", errors.New("empty source")
	}
	return source, nil
}

func buildLoopGenerationPrompts(input DraftGenerationInput) (string, string) {
	system := strings.TrimSpace(`
You generate LoopScript schema version 1 for an AI Agent execution loop.
Return exactly one JSON object with exactly one string field named "source".
Do not return Markdown, code fences, Blockly JSON, explanations, or extra fields.
The source must contain one Loop, limits, repeat, one agent task, one verification step, and an on_failure policy.
Preserve verification and budget boundaries unless the user explicitly requests stricter valid values.
Worker selection and runtime configuration are outside LoopScript. Never declare or choose a Worker, model, runner, snapshot, or runtime.
Never include credentials, API keys, tokens, or secret literal values.
Do not execute or start the Loop. Only propose valid source code.
`)
	if input.CurrentSource == "" {
		system += "\nFor a new Loop, set the verification command exactly to \"false\" so the draft cannot execute successfully before explicit user review."
	}
	user := fmt.Sprintf(
		"Interface locale: %s\nUser request:\n%s\n\nCurrent LoopScript:\n%s",
		input.Locale,
		input.Prompt,
		input.CurrentSource,
	)
	return system, user
}
