import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@/test/test-utils";
import { AcpPromptInput } from "@/components/workspace/acp/AcpPromptInput";
import {
  __seedAcpSessionForTests,
  __resetAcpSessionsForTests,
} from "@/stores/acpSession";
import { EMPTY_SESSION } from "@/stores/acpSessionTypes";
import { relayPool } from "@/stores/relayConnection";

vi.mock("@/stores/relayConnection", () => ({
  relayPool: {
    sendAcpCommand: vi.fn().mockResolvedValue(undefined),
    isConnected: vi.fn().mockReturnValue(true),
  },
}));

describe("AcpPromptInput", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(relayPool.isConnected).mockReturnValue(true);
    __resetAcpSessionsForTests();
  });

  it("renders input with correct placeholder", () => {
    render(<AcpPromptInput podKey="pod-1" />);
    expect(screen.getByPlaceholderText("Send instruction...")).toBeInTheDocument();
  });

  it("sends prompt on Enter", async () => {
    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...");

    fireEvent.change(textarea, { target: { value: "create hello world" } });
    fireEvent.keyDown(textarea, { key: "Enter" });

    await waitFor(() =>
      expect(relayPool.sendAcpCommand).toHaveBeenCalledWith("pod-1", {
        type: "prompt",
        prompt: "create hello world",
      }),
    );
  });

  it("does not send on Shift+Enter", () => {
    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...");

    fireEvent.change(textarea, { target: { value: "multiline" } });
    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: true });

    expect(relayPool.sendAcpCommand).not.toHaveBeenCalled();
  });

  it("clears input after sending", async () => {
    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...") as HTMLTextAreaElement;

    fireEvent.change(textarea, { target: { value: "test" } });
    fireEvent.keyDown(textarea, { key: "Enter" });

    await waitFor(() => expect(textarea.value).toBe(""));
  });

  it("does not send empty prompt", () => {
    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...");

    fireEvent.keyDown(textarea, { key: "Enter" });
    expect(relayPool.sendAcpCommand).not.toHaveBeenCalled();
  });

  it("shows error when not connected", () => {
    vi.mocked(relayPool.isConnected).mockReturnValue(false);

    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...");

    fireEvent.change(textarea, { target: { value: "test" } });
    fireEvent.keyDown(textarea, { key: "Enter" });

    expect(screen.getByText("Not connected")).toBeInTheDocument();
    expect(relayPool.sendAcpCommand).not.toHaveBeenCalled();
  });

  it("keeps the prompt when the relay rejects a send", async () => {
    vi.mocked(relayPool.sendAcpCommand).mockRejectedValueOnce(new Error("relay disconnected"));

    render(<AcpPromptInput podKey="pod-1" />);
    const textarea = screen.getByPlaceholderText("Send instruction...") as HTMLTextAreaElement;

    fireEvent.change(textarea, { target: { value: "keep this instruction" } });
    fireEvent.keyDown(textarea, { key: "Enter" });

    await waitFor(() => expect(screen.getByText("Not connected")).toBeInTheDocument());
    expect(textarea.value).toBe("keep this instruction");
  });

  it("sends interrupt command when cancel button is clicked during processing", () => {
    __seedAcpSessionForTests("pod-1", { ...EMPTY_SESSION, state: "processing" });

    const { container } = render(<AcpPromptInput podKey="pod-1" />);
    const cancelBtn = container.querySelector("button[title='Cancel']");
    expect(cancelBtn).toBeTruthy();
    fireEvent.click(cancelBtn!);

    expect(relayPool.sendAcpCommand).toHaveBeenCalledWith("pod-1", {
      type: "interrupt",
    });
  });
});
