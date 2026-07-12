import { describe, expect, it, vi } from "vitest";
import { sendGrantedTerminalResize } from "./relay-terminal-resize";

describe("sendGrantedTerminalResize", () => {
  it("does not resize before the browser holds the control lease", () => {
    const relay = { send_resize: vi.fn() };

    expect(sendGrantedTerminalResize(relay, "pod-1", { status: "observer" }, 100, 30)).toBe(false);
    expect(relay.send_resize).not.toHaveBeenCalled();
  });

  it("synchronizes the fitted size as soon as the lease is granted", () => {
    const relay = { send_resize: vi.fn().mockResolvedValue(undefined) };

    expect(
      sendGrantedTerminalResize(relay, "pod-1", { status: "granted", leaseId: "lease-1" }, 100, 30),
    ).toBe(true);
    expect(relay.send_resize).toHaveBeenCalledWith("pod-1", 100, 30);
  });
});
