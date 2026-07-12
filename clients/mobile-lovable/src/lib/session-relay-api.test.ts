import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api-fetch";
import { getSessionRelayConnection } from "./session-relay-api";

vi.mock("./api-fetch", () => ({
  apiFetch: vi.fn(),
}));

const apiFetchMock = vi.mocked(apiFetch);

describe("getSessionRelayConnection", () => {
  beforeEach(() => {
    apiFetchMock.mockReset();
  });

  it("reads only the server-issued browser connection", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          relay_url: "wss://relay.example",
          token: "browser-token",
          pod_key: "mobile-pod",
        }),
        { status: 200 },
      ),
    );

    await expect(getSessionRelayConnection("session-1")).resolves.toEqual({
      relayUrl: "wss://relay.example",
      token: "browser-token",
      podKey: "mobile-pod",
    });
    expect(apiFetchMock).toHaveBeenCalledWith("/v1/sessions/session-1/relay-connection");
  });

  it("rejects an unusable response instead of inventing a relay URL", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(JSON.stringify({ relay_url: "" }), { status: 200 }),
    );

    await expect(getSessionRelayConnection("session-1")).rejects.toThrow(
      "Invalid relay connection",
    );
  });
});
