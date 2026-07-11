package agentpod

import "context"

func (s *PodService) DeleteTerminalPod(ctx context.Context, podKey string) error {
	pod, err := s.GetPod(ctx, podKey)
	if err != nil {
		return err
	}
	if !pod.IsTerminal() {
		return ErrPodNotTerminal
	}
	deleted, err := s.repo.DeleteTerminalByKey(ctx, podKey)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrPodNotFound
	}
	return nil
}
