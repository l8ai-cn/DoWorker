package knowledgebase

import (
	"context"
	"crypto/sha1" //nolint:gosec // git blob ids are SHA-1 by protocol, not a security boundary
	"encoding/hex"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
	"github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase/connector"
)

// SyncFromConnector pulls the external source and commits changed docs under
// raw/{source_type}/ in one batch commit. raw/ keeps append/overwrite
// semantics — deletions upstream are NOT propagated (history stays in git).
func (s *Service) SyncFromConnector(ctx context.Context, kb *knowledgebase.KnowledgeBase, conn connector.Connector) error {
	s.setSyncStatus(ctx, kb, knowledgebase.SyncStatusSyncing, nil)
	if err := s.syncDocs(ctx, kb, conn); err != nil {
		msg := err.Error()
		s.setSyncStatus(ctx, kb, knowledgebase.SyncStatusFailed, &msg)
		return err
	}
	s.setSyncStatus(ctx, kb, knowledgebase.SyncStatusSynced, nil)
	return nil
}

func (s *Service) syncDocs(ctx context.Context, kb *knowledgebase.KnowledgeBase, conn connector.Connector) error {
	sourceConfig, err := s.decryptSourceSecrets(kb.SourceConfig)
	if err != nil {
		return fmt.Errorf("decrypt source_config: %w", err)
	}
	refs, err := conn.ListDocs(ctx, sourceConfig)
	if err != nil {
		return fmt.Errorf("list docs: %w", err)
	}

	repoName := repoNameFromPath(kb.GitRepoPath)
	existing, err := s.existingBlobSHAs(ctx, repoName, kb.DefaultBranch)
	if err != nil {
		return fmt.Errorf("list repo tree: %w", err)
	}

	var changes []gitea.FileChange
	isUpdate := map[string]string{}
	for _, ref := range refs {
		doc, err := conn.FetchDoc(ctx, sourceConfig, ref)
		if err != nil {
			return fmt.Errorf("fetch doc %q: %w", ref.Title, err)
		}
		path := "raw/" + conn.SourceType() + "/" + ref.Path
		if sha, ok := existing[path]; ok {
			if sha == gitBlobSHA(doc.Markdown) {
				continue
			}
			isUpdate[path] = sha
		}
		changes = append(changes, gitea.FileChange{Path: path, Content: doc.Markdown})
	}
	if len(changes) == 0 {
		return nil
	}

	message := fmt.Sprintf("sync(%s): %d document(s)", conn.SourceType(), len(changes))
	author := gitea.CommitAuthor{Name: "kb-sync", Email: "kb-sync@agentsmesh.local"}
	if err := s.git.CommitFiles(ctx, repoName, kb.DefaultBranch, message, author, changes, isUpdate); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	s.log.Info("knowledge base synced", "slug", kb.Slug, "source", conn.SourceType(), "changed", len(changes))
	return nil
}

func (s *Service) existingBlobSHAs(ctx context.Context, repoName, branch string) (map[string]string, error) {
	entries, err := s.git.ListTree(ctx, repoName, branch)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.Type == "blob" {
			out[e.Path] = e.SHA
		}
	}
	return out, nil
}

func (s *Service) setSyncStatus(ctx context.Context, kb *knowledgebase.KnowledgeBase, status string, syncErr *string) {
	updates := map[string]any{"sync_status": status, "sync_error": syncErr}
	if status == knowledgebase.SyncStatusSynced {
		updates["last_synced_at"] = time.Now()
	}
	if err := s.repo.Update(ctx, kb.OrganizationID, kb.ID, updates); err != nil {
		s.log.Warn("knowledge base sync status update failed", "slug", kb.Slug, "error", err)
	}
}

// gitBlobSHA computes the git object id for a blob, letting the sync skip
// commits for unchanged content without fetching the old file body.
func gitBlobSHA(content string) string {
	h := sha1.New() //nolint:gosec
	_, _ = fmt.Fprintf(h, "blob %d\x00", len(content))
	_, _ = h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
