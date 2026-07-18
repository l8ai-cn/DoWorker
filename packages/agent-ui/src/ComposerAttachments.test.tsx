import "@testing-library/jest-dom/vitest";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import { ComposerAttachments } from "./ComposerAttachments";
import type { AgentSessionRuntime } from "./contracts";

describe("ComposerAttachments", () => {
  it("uploads, displays, and removes a session attachment", async () => {
    const onChange = vi.fn();
    const uploadAttachment = vi.fn(async () => ({
      bytes: 4,
      id: "file-1",
      mediaType: "text/plain",
      name: "brief.txt",
    }));
    const runtime = { uploadAttachment } as unknown as AgentSessionRuntime;
    const view = render(
      <AgentWorkspaceLocaleProvider locale="zh-CN">
        <ComposerAttachments
          attachments={[]}
          disabled={false}
          onChange={onChange}
          runtime={runtime}
          sessionId="session-1"
        />
      </AgentWorkspaceLocaleProvider>,
    );

    fireEvent.change(screen.getByTestId("agent-attachment-input"), {
      target: { files: [new File(["test"], "brief.txt", { type: "text/plain" })] },
    });

    await waitFor(() => expect(uploadAttachment).toHaveBeenCalled());
    const attachment = await uploadAttachment.mock.results[0].value;
    view.rerender(
      <AgentWorkspaceLocaleProvider locale="zh-CN">
        <ComposerAttachments
          attachments={[attachment]}
          disabled={false}
          onChange={onChange}
          runtime={runtime}
          sessionId="session-1"
        />
      </AgentWorkspaceLocaleProvider>,
    );
    expect(screen.getByText("brief.txt")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "移除附件: brief.txt" }));
    expect(onChange).toHaveBeenLastCalledWith([]);
  });

  it("explains rejected attachment types in the composer", async () => {
    const runtime = {
      uploadAttachment: vi.fn(async () => {
        throw new Error("agent_attachment_upload_unsupported");
      }),
    } as unknown as AgentSessionRuntime;

    render(
      <AgentWorkspaceLocaleProvider locale="zh-CN">
        <ComposerAttachments
          attachments={[]}
          disabled={false}
          onChange={vi.fn()}
          runtime={runtime}
          sessionId="session-1"
        />
      </AgentWorkspaceLocaleProvider>,
    );

    fireEvent.change(screen.getByTestId("agent-attachment-input"), {
      target: { files: [new File(["test"], "unsupported.bin")] },
    });

    expect(
      await screen.findByText("当前会话不支持此附件类型。"),
    ).toBeVisible();
  });
});
