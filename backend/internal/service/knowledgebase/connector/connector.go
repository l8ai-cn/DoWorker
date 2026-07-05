// Package connector materializes external knowledge sources (Feishu wiki,
// DingTalk wiki, Google Drive) as markdown files. The sync worker commits
// fetched docs under raw/{source_type}/ in the KB git repository, keeping
// the pod mount pipeline git-only (single-direction: external → raw/).
package connector

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var ErrBadConfig = errors.New("connector: invalid source_config")

// DocRef identifies one document in the external source. Path is the
// markdown file path relative to raw/{source_type}/ in the KB repo.
type DocRef struct {
	ID    string
	Title string
	Path  string
}

type Doc struct {
	Ref      DocRef
	Markdown string
}

type Connector interface {
	SourceType() string
	ListDocs(ctx context.Context, config json.RawMessage) ([]DocRef, error)
	FetchDoc(ctx context.Context, config json.RawMessage, ref DocRef) (*Doc, error)
}

// NewRegistry returns all built-in connectors keyed by source_type.
func NewRegistry() map[string]Connector {
	hc := &http.Client{Timeout: 30 * time.Second}
	return map[string]Connector{
		"feishu":   &FeishuConnector{HTTP: hc},
		"dingtalk": &DingTalkConnector{HTTP: hc},
		"google":   &GoogleDriveConnector{HTTP: hc},
	}
}

var docPathSanitizer = regexp.MustCompile(`[^\p{L}\p{N}._-]+`)

// DocPath derives a stable markdown path from a doc title + provider id.
// The id suffix keeps paths unique when titles collide or get renamed.
func DocPath(title, id string) string {
	base := docPathSanitizer.ReplaceAllString(strings.TrimSpace(title), "-")
	base = strings.Trim(base, "-.")
	if base == "" {
		base = "untitled"
	}
	if len(base) > 80 {
		base = base[:80]
	}
	shortID := id
	if len(shortID) > 12 {
		shortID = shortID[len(shortID)-12:]
	}
	return base + "-" + shortID + ".md"
}

func decodeConfig(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		return ErrBadConfig
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return errors.Join(ErrBadConfig, err)
	}
	return nil
}
