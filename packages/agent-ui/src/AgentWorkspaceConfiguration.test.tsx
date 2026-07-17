import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AgentWorkspace } from "./AgentWorkspace";
import {
  agentWorkspaceRuntime,
  agentWorkspaceSnapshot,
} from "./AgentWorkspace.test-fixture";

describe("AgentWorkspace configuration", () => {
  it("updates configuration and sends supported slash commands", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);

    render(<AgentWorkspace runtime={agentRuntime} sessionId={snapshot.sessionId} />);

    fireEvent.click(await screen.findByRole("combobox", { name: "Permissions" }));
    fireEvent.click(screen.getByRole("option", { name: "Accept edits" }));
    fireEvent.click(screen.getByRole("combobox", { name: "Model" }));
    fireEvent.click(screen.getByRole("option", { name: "GPT-5.5" }));

    const input = screen.getByLabelText("Message the agent");
    fireEvent.change(input, { target: { value: "/comp" } });
    fireEvent.click(screen.getByRole("button", { name: "/compact" }));
    fireEvent.keyDown(input, { key: "Enter" });

    await waitFor(() => {
      expect(agentRuntime.updateConfiguration).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { permissionMode: "acceptEdits" },
      );
      expect(agentRuntime.updateConfiguration).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { model: "gpt-5.5" },
      );
      expect(agentRuntime.sendSlashCommand).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { name: "compact", arguments: "" },
      );
    });
  });

  it("renders Chinese chrome with custom configuration pickers", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);
    const { container } = render(
      <AgentWorkspace
        locale="zh-CN"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByRole("heading", {
        name: "Codex，我能为你做什么？",
      }),
    ).toBeVisible();
    expect(screen.getByRole("tab", { name: "对话" })).toBeVisible();
    expect(screen.getByText("智能体模式")).toBeVisible();
    expect(screen.getByText("请求审批")).toBeVisible();
    expect(screen.getByText("就绪")).toBeVisible();
    expect(screen.getByLabelText("给智能体发送消息")).toHaveAttribute(
      "placeholder",
      "让 Codex 帮你完成任务…",
    );
    expect(container.querySelector("select")).toBeNull();

    const permissionPicker = screen.getByRole("combobox", { name: "权限" });
    fireEvent.click(permissionPicker);
    expect(screen.getByRole("listbox", { name: "权限选项" })).toBeVisible();
    fireEvent.click(screen.getByRole("option", { name: "自动接受编辑" }));

    await waitFor(() => {
      expect(agentRuntime.updateConfiguration).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { permissionMode: "acceptEdits" },
      );
      expect(permissionPicker).toHaveFocus();
    });
  });

  it("requires an explicit permission choice when no mode is active", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.configuration = [
      {
        id: "permissionMode",
        label: "Permissions",
        value: "",
        options: [
          { value: "bypass", label: "bypass" },
          { value: "ask_dangerous", label: "ask_dangerous" },
          { value: "ask_any_write", label: "ask_any_write" },
        ],
      },
    ];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    const picker = await screen.findByRole("combobox", { name: "权限" });
    expect(picker).toHaveTextContent("权限");
    expect(picker).not.toHaveTextContent("无需确认");

    fireEvent.click(picker);
    screen.getAllByRole("option").forEach((option) => {
      expect(option).toHaveAttribute("aria-selected", "false");
    });
    fireEvent.click(screen.getByRole("option", { name: "无需确认" }));

    await waitFor(() => {
      expect(agentRuntime.updateConfiguration).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { permissionMode: "bypass" },
      );
    });
  });

  it("renders generated permission controls with protocol keys and English labels", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.configuration = [
      {
        id: "permission_mode",
        label: "Permissions",
        value: "",
        options: [
          { value: "bypass", label: "bypass" },
          { value: "ask_dangerous", label: "ask_dangerous" },
          { value: "ask_any_write", label: "ask_any_write" },
        ],
      },
    ];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);

    render(<AgentWorkspace runtime={agentRuntime} sessionId={snapshot.sessionId} />);

    fireEvent.click(
      await screen.findByRole("combobox", { name: "Permissions" }),
    );
    fireEvent.click(screen.getByRole("option", { name: "Full access" }));

    await waitFor(() => {
      expect(agentRuntime.updateConfiguration).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        { permission_mode: "bypass" },
      );
    });
  });

  it("shows synchronized configuration without allowing read-only observers to edit it", async () => {
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.configuration = [
      {
        id: "permission_mode",
        label: "Permissions",
        value: "bypass",
        options: [
          { value: "bypass", label: "bypass" },
          { value: "ask_dangerous", label: "ask_dangerous" },
        ],
      },
    ];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);

    render(
      <AgentWorkspace
        readOnly
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    const picker = await screen.findByRole("combobox", {
      name: "Permissions",
    });
    expect(picker).toHaveTextContent("Full access");
    expect(picker).toBeDisabled();
    expect(picker).not.toHaveAttribute("aria-busy");
    expect(picker.querySelector(".animate-spin")).toBeNull();
    expect(agentRuntime.updateConfiguration).not.toHaveBeenCalled();
  });

  it("closes the picker when keyboard focus leaves it", async () => {
    const user = userEvent.setup();
    const snapshot = agentWorkspaceSnapshot();
    snapshot.status = "idle";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = agentWorkspaceRuntime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    fireEvent.click(await screen.findByRole("combobox", { name: "权限" }));
    expect(screen.getByRole("listbox", { name: "权限选项" })).toBeVisible();
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "更改前询问" })).toHaveFocus();
    });

    await user.tab();

    await waitFor(() => {
      expect(
        screen.queryByRole("listbox", { name: "权限选项" }),
      ).not.toBeInTheDocument();
    });
  });
});
