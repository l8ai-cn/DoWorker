package runner

import "fmt"

func (h *RunnerMessageHandler) refreshRelayToken(podKey, relayURL, token string) error {
	pod, ok := h.podStore.Get(podKey)
	if !ok {
		return fmt.Errorf("pod not found: %s", podKey)
	}

	pod.LockRelay()
	defer pod.UnlockRelay()
	relayClient := pod.RelayClient
	if relayClient == nil || !relayClient.IsConnected() || relayClient.GetRelayURL() != relayURL {
		return fmt.Errorf("relay subscription completed without active client")
	}
	relayClient.UpdateToken(token)
	return nil
}
