const RESET_POLL_INTERVAL_MS = 25;

export interface MobileRelayResetConnection {
  disconnect(podKey: string): Promise<void>;
  get_status(podKey: string): Promise<string>;
}

type Pause = () => Promise<void>;

export async function resetMobileRelayConnection(
  relay: MobileRelayResetConnection,
  podKey: string,
  pause: Pause = () => new Promise((resolve) => window.setTimeout(resolve, RESET_POLL_INTERVAL_MS)),
) {
  await relay.disconnect(podKey);
  while ((await relay.get_status(podKey)) !== "disconnected") {
    await pause();
  }
}
