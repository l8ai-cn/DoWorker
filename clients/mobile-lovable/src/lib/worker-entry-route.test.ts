import { describe, expect, it } from "vitest";
import { resolveWorkerEntryRoute } from "./worker-entry-route";

describe("resolveWorkerEntryRoute", () => {
  it("opens an ACP Worker through the Pod chat route", () => {
    expect(
      resolveWorkerEntryRoute({
        interactionMode: "acp",
        consoleAvailable: true,
      }),
    ).toBe("chat");
  });

  it("opens a PTY Worker through the Pod terminal route", () => {
    expect(
      resolveWorkerEntryRoute({
        interactionMode: "pty",
        consoleAvailable: true,
      }),
    ).toBe("terminal");
  });

  it("does not route an unavailable Worker to a console", () => {
    expect(
      resolveWorkerEntryRoute({
        interactionMode: "pty",
        consoleAvailable: false,
      }),
    ).toBeNull();
  });
});
