import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MobileAcpPanel } from "./mobile-acp-panel";
import { useMobileAcpRelay } from "@/hooks/use-mobile-acp-relay";

vi.mock("@/hooks/use-mobile-acp-relay", () => ({ useMobileAcpRelay: vi.fn() }));

const hookState = {
  connection: "connected" as const,
  control: { acquire: vi.fn(), release: vi.fn(), acquiring: false, error: null },
  error: null,
  lease: { status: "granted" as const },
  reconnect: vi.fn(),
  respondPermission: vi.fn().mockResolvedValue(undefined),
  sendPrompt: vi.fn().mockResolvedValue(undefined),
  session: {
    state: "idle",
    messages: [{ role: "assistant", text: "How can I help?", complete: true }],
    pendingPermissions: [
      {
        requestId: "perm-1",
        toolName: "shell",
        argumentsJson: "{\"command\":\"ls\"}",
        description: "List files",
      },
    ],
  },
  interrupt: vi.fn().mockResolvedValue(undefined),
};

describe("MobileAcpPanel", () => {
  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("sends a prompt through the direct Pod ACP connection", async () => {
    vi.mocked(useMobileAcpRelay).mockReturnValue(hookState);
    render(<MobileAcpPanel podKey="pod-1" />);

    fireEvent.change(screen.getByRole("textbox"), { target: { value: "检查项目状态" } });
    fireEvent.click(screen.getByRole("button", { name: "发送消息" }));

    expect(hookState.sendPrompt).toHaveBeenCalledWith("检查项目状态");
  });

  it("shows and resolves pending permission requests", () => {
    vi.mocked(useMobileAcpRelay).mockReturnValue(hookState);
    render(<MobileAcpPanel podKey="pod-1" />);

    expect(screen.getByText("List files")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "允许 shell" }));
    expect(hookState.respondPermission).toHaveBeenCalledWith("perm-1", true);
  });
});
