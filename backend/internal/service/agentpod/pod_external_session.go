package agentpod

import "context"

func (s *PodService) UpdateExternalSessionID(ctx context.Context, podKey, externalID string) error {
	_, err := s.repo.UpdateByKey(ctx, podKey, map[string]interface{}{
		"external_session_id": externalID,
	})
	return err
}
