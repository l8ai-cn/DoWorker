package runner

import "fmt"

func refreshRelayToken(pod *Pod, relayURL, token string) error {
	pod.LockRelay()
	defer pod.UnlockRelay()
	relayClient := pod.RelayClient
	if relayClient == nil || !relayClient.IsConnected() || relayClient.GetRelayURL() != relayURL {
		return fmt.Errorf("relay subscription completed without active client")
	}
	relayClient.UpdateToken(token)
	return nil
}
