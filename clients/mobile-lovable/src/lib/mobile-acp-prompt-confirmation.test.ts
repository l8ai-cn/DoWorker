import { describe, expect, it, vi } from "vitest";
import { createMobileAcpPromptConfirmation } from "./mobile-acp-prompt-confirmation";

describe("mobile ACP prompt confirmation", () => {
  it("resolves only when Runner echoes the matching accepted user prompt", async () => {
    const confirmation = createMobileAcpPromptConfirmation();
    const accepted = confirmation.waitFor("request-1");

    expect(confirmation.consume({ type: "contentChunk", role: "user", requestId: "other" })).toBe(
      false,
    );
    expect(
      confirmation.consume({ type: "contentChunk", role: "user", requestId: "request-1" }),
    ).toBe(true);
    await expect(accepted).resolves.toBeUndefined();
  });

  it("rejects the matching prompt when Runner reports command failure", async () => {
    const confirmation = createMobileAcpPromptConfirmation();
    const accepted = confirmation.waitFor("request-1");

    confirmation.consume({
      type: "commandFailed",
      requestId: "request-1",
      message: "ACP 初始化超时",
    });

    await expect(accepted).rejects.toThrow("ACP 初始化超时");
  });

  it("rejects pending prompts when the Relay connection closes", async () => {
    const confirmation = createMobileAcpPromptConfirmation();
    const accepted = confirmation.waitFor("request-1");

    confirmation.rejectAll("Worker 连接已关闭");

    await expect(accepted).rejects.toThrow("Worker 连接已关闭");
  });

  it("times out when Runner never acknowledges a prompt", async () => {
    vi.useFakeTimers();
    const confirmation = createMobileAcpPromptConfirmation(100);
    const accepted = confirmation.waitFor("request-1");
    const assertion = expect(accepted).rejects.toThrow("Worker 未确认接收消息，请重试");

    await vi.advanceTimersByTimeAsync(100);

    await assertion;
    vi.useRealTimers();
  });
});
