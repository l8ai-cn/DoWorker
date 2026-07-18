import "@testing-library/jest-dom/vitest";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import {
  agentWorkspaceRuntime,
  agentWorkspaceSnapshot,
} from "./AgentWorkspace.test-fixture";
import { ConversationComposer } from "./ConversationComposer";

describe("ConversationComposer", () => {
  it("sends the uploaded attachment with a regular message", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);
    agentRuntime.uploadAttachment = vi.fn(async () => ({
      bytes: 4,
      id: "file_1",
      mediaType: "text/csv",
      name: "sales.csv",
    }));

    render(
      <AgentWorkspaceLocaleProvider locale="zh-CN">
        <ConversationComposer
          onError={vi.fn()}
          presentation="developer"
          runtime={agentRuntime}
          snapshot={snapshot}
        />
      </AgentWorkspaceLocaleProvider>,
    );

    fireEvent.change(screen.getByTestId("agent-attachment-input"), {
      target: {
        files: [new File(["data"], "sales.csv", { type: "text/csv" })],
      },
    });
    expect(await screen.findByText("sales.csv")).toBeVisible();
    fireEvent.change(screen.getByLabelText("给智能体发送消息"), {
      target: { value: "分析这份数据" },
    });
    fireEvent.click(screen.getByRole("button", { name: "发送消息" }));

    await waitFor(() => {
      expect(agentRuntime.sendMessage).toHaveBeenCalledWith(
        snapshot.sessionId,
        expect.any(String),
        {
          attachments: [{
            bytes: 4,
            id: "file_1",
            mediaType: "text/csv",
            name: "sales.csv",
          }],
          text: "分析这份数据",
        },
      );
    });
  });

  it("keeps attachments when a slash command cannot accept them", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);
    agentRuntime.uploadAttachment = vi.fn(async () => ({
      bytes: 4,
      id: "file-1",
      mediaType: "text/plain",
      name: "brief.txt",
    }));
    const onError = vi.fn();

    render(
      <AgentWorkspaceLocaleProvider locale="zh-CN">
        <ConversationComposer
          onError={onError}
          presentation="developer"
          runtime={agentRuntime}
          snapshot={snapshot}
        />
      </AgentWorkspaceLocaleProvider>,
    );

    fireEvent.change(screen.getByTestId("agent-attachment-input"), {
      target: {
        files: [new File(["test"], "brief.txt", { type: "text/plain" })],
      },
    });
    expect(await screen.findByText("brief.txt")).toBeVisible();

    const input = screen.getByLabelText("给智能体发送消息");
    fireEvent.change(input, { target: { value: "/compact" } });
    fireEvent.keyDown(input, { key: "Enter" });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(
        expect.objectContaining({
          message: "斜杠命令不能携带附件。请移除附件，或改为发送普通消息。",
        }),
      );
    });
    expect(screen.getByText("brief.txt")).toBeVisible();
    expect(input).toHaveValue("/compact");
    expect(agentRuntime.sendSlashCommand).not.toHaveBeenCalled();
    expect(agentRuntime.sendMessage).not.toHaveBeenCalled();
  });
});
