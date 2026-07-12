import { describe, expect, it, vi } from "vitest";
import { resetMobileRelayConnection } from "./mobile-relay-reset";

describe("resetMobileRelayConnection", () => {
  it("waits for the old driver to leave the pool before reuse", async () => {
    const relay = {
      disconnect: vi.fn().mockResolvedValue(undefined),
      get_status: vi
        .fn()
        .mockResolvedValueOnce("reconnecting")
        .mockResolvedValueOnce("disconnected"),
    };
    const pause = vi.fn().mockResolvedValue(undefined);

    await resetMobileRelayConnection(relay, "pod-1", pause);

    expect(relay.disconnect).toHaveBeenCalledWith("pod-1");
    expect(relay.get_status).toHaveBeenCalledTimes(2);
    expect(pause).toHaveBeenCalledTimes(1);
  });
});
