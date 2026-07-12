import { describe, expect, it } from "vitest";
import { relayConnectionState } from "./relay-connection-state";

describe("relayConnectionState", () => {
  it("keeps the terminal locked until Relay confirms a connected state", () => {
    expect(relayConnectionState("connecting")).toBe("connecting");
    expect(relayConnectionState("disconnected")).toBe("reconnecting");
    expect(relayConnectionState("connected")).toBe("connected");
  });

  it("surfaces the Relay error state instead of treating it as disconnected", () => {
    expect(relayConnectionState("error")).toBe("failed");
  });
});
