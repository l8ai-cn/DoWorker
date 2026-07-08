package codeximport

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Detect classifies a local path as a Codex rollout transcript or a workflow
// output directory and returns the concrete file/dir that should be read.
//
// Resolution order:
//   - a *.jsonl file             -> KindRollout (that file)
//   - a directory containing     -> KindRollout (newest rollout-*.jsonl)
//     rollout-*.jsonl file(s)
//   - a directory with           -> KindOutputDir (that directory)
//     conversation_input.json or run_manifest.json
//   - a directory with exactly    -> KindRollout (that file)
//     one *.jsonl file
func Detect(path string) (Kind, string, error) {
	if strings.TrimSpace(path) == "" {
		return "", "", fmt.Errorf("codeximport: empty source path")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("codeximport: cannot access %q: %w", path, err)
	}
	if !info.IsDir() {
		if strings.HasSuffix(strings.ToLower(path), ".jsonl") {
			return KindRollout, path, nil
		}
		return "", "", fmt.Errorf("codeximport: %q is not a .jsonl transcript", path)
	}

	if rollout := newestRollout(path); rollout != "" {
		return KindRollout, rollout, nil
	}
	if hasAny(path, "conversation_input.json", "run_manifest.json", "workflow_checkpoint.json") {
		return KindOutputDir, path, nil
	}
	if single := singleJSONL(path); single != "" {
		return KindRollout, single, nil
	}
	return "", "", fmt.Errorf("codeximport: %q is not a recognizable Codex source (no rollout jsonl or workflow manifest)", path)
}

// Convert detects the source kind and returns the normalized conversation.
func Convert(path string) (*Result, error) {
	kind, concrete, err := Detect(path)
	if err != nil {
		return nil, err
	}
	switch kind {
	case KindRollout:
		return convertRollout(concrete)
	case KindOutputDir:
		return convertOutputDir(concrete)
	default:
		return nil, fmt.Errorf("codeximport: unsupported source kind %q", kind)
	}
}

func hasAny(dir string, names ...string) bool {
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			return true
		}
	}
	return false
}

// newestRollout returns the most recently modified rollout-*.jsonl file
// directly inside dir, or "" when none exist. Rollout filenames are timestamp
// prefixed, but mtime is the robust ordering signal.
func newestRollout(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	type candidate struct {
		path    string
		modTime int64
	}
	var cands []candidate
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "rollout-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		cands = append(cands, candidate{path: filepath.Join(dir, name), modTime: info.ModTime().UnixNano()})
	}
	if len(cands) == 0 {
		return ""
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].modTime > cands[j].modTime })
	return cands[0].path
}

// singleJSONL returns the path of the only *.jsonl file in dir, or "" when
// there are zero or more than one.
func singleJSONL(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var found string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".jsonl") {
			continue
		}
		if found != "" {
			return "" // ambiguous
		}
		found = filepath.Join(dir, e.Name())
	}
	return found
}
