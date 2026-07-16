import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import { ApprovalDock } from "./ApprovalDock";
import type {
  AgentPermissionRequest,
  AgentSessionRuntime,
} from "./contracts";

function runtime(): AgentSessionRuntime {
  return {
    open: vi.fn(async () => undefined),
    close: vi.fn(),
    getSnapshot: vi.fn(),
    subscribe: vi.fn(() => () => undefined),
    sendMessage: vi.fn(async () => undefined),
    interrupt: vi.fn(async () => undefined),
    resolvePermission: vi.fn(async () => undefined),
    updateConfiguration: vi.fn(async () => undefined),
    loadOlder: vi.fn(async () => undefined),
  };
}

describe("ApprovalDock", () => {
  it("resolves approvals with the shared answers content shape", async () => {
    const agentRuntime = runtime();
    const permission = {
      id: "approval-1",
      kind: "approval",
      title: "Run release command",
      description: "pnpm release",
    } as AgentPermissionRequest;

    render(
      <ApprovalDock
        onError={vi.fn()}
        permissions={[permission]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Approve" }));

    await waitFor(() =>
      expect(agentRuntime.resolvePermission).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        "approval-1",
        { action: "accept", content: { answers: {} } },
      ),
    );
  });

  it("submits an approval at most once while resolution is pending", async () => {
    let complete!: () => void;
    const pending = new Promise<void>((resolve) => {
      complete = resolve;
    });
    const agentRuntime = runtime();
    agentRuntime.resolvePermission = vi.fn(() => pending);
    const permission = {
      id: "approval-pending",
      kind: "approval",
      title: "Run release command",
      description: "pnpm release",
    } as AgentPermissionRequest;

    render(
      <ApprovalDock
        onError={vi.fn()}
        permissions={[permission]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    const approve = screen.getByRole("button", { name: "Approve" });
    fireEvent.click(approve);
    fireEvent.click(approve);

    expect(agentRuntime.resolvePermission).toHaveBeenCalledTimes(1);
    expect(approve).toBeDisabled();
    expect(screen.getByRole("button", { name: "Reject" })).toBeDisabled();
    complete();
    await pending;
  });

  it("requires every structured answer before submitting arrays keyed by stable ids", async () => {
    const agentRuntime = runtime();
    const permission = {
      id: "question-request-1",
      kind: "question",
      title: "Agent needs input",
      questions: [
        {
          id: "framework",
          prompt: "Which framework?",
          header: "Framework",
          options: [
            { label: "React", description: "Use the existing React stack" },
            { label: "Vue", description: "Switch to Vue" },
          ],
          multiple: false,
          allowCustom: false,
          secret: false,
        },
        {
          id: "targets",
          prompt: "Which targets?",
          header: "Targets",
          options: [
            { label: "Browser", description: "Ship the browser client" },
            { label: "API", description: "Ship the API client" },
          ],
          multiple: true,
          allowCustom: true,
          secret: false,
        },
        {
          id: "token",
          prompt: "Enter the release token",
          header: "Credentials",
          options: [],
          multiple: false,
          allowCustom: true,
          secret: true,
        },
      ],
    } as AgentPermissionRequest;

    render(
      <ApprovalDock
        onError={vi.fn()}
        permissions={[permission]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    const submit = screen.getByRole("button", { name: "Submit answers" });
    expect(submit).toBeDisabled();
    expect(screen.getByText("Use the existing React stack")).toBeVisible();

    fireEvent.click(screen.getByLabelText("React"));
    fireEvent.click(screen.getByLabelText("Browser"));
    fireEvent.change(
      screen.getByLabelText("Custom answer for Which targets?"),
      { target: { value: "CLI" } },
    );

    const secret = screen.getByLabelText("Custom answer for Enter the release token");
    expect(secret).toHaveAttribute("type", "password");
    fireEvent.change(secret, { target: { value: "release-secret" } });

    expect(submit).toBeEnabled();
    fireEvent.click(submit);

    await waitFor(() =>
      expect(agentRuntime.resolvePermission).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        "question-request-1",
        {
          action: "accept",
          content: {
            answers: {
              framework: ["React"],
              targets: ["Browser", "CLI"],
              token: ["release-secret"],
            },
          },
        },
      ),
    );
  });

  it("declines structured questions without submitting partial answers", async () => {
    const agentRuntime = runtime();
    const permission = {
      id: "question-request-2",
      kind: "question",
      title: "Agent needs input",
      questions: [
        {
          id: "decision",
          prompt: "Proceed?",
          header: "Decision",
          options: [{ label: "Yes", description: "Continue" }],
          multiple: false,
          allowCustom: false,
          secret: false,
        },
      ],
    } as AgentPermissionRequest;

    render(
      <ApprovalDock
        onError={vi.fn()}
        permissions={[permission]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Reject" }));

    await waitFor(() =>
      expect(agentRuntime.resolvePermission).toHaveBeenCalledWith(
        "session-1",
        expect.any(String),
        "question-request-2",
        { action: "decline" },
      ),
    );
  });

  it("submits structured answers at most once while resolution is pending", async () => {
    let complete!: () => void;
    const pending = new Promise<void>((resolve) => {
      complete = resolve;
    });
    const agentRuntime = runtime();
    agentRuntime.resolvePermission = vi.fn(() => pending);
    const permission = {
      id: "question-pending",
      kind: "question",
      title: "Agent needs input",
      questions: [
        {
          id: "decision",
          prompt: "Proceed?",
          header: "Decision",
          options: [{ label: "Yes", description: "Continue" }],
          multiple: false,
          allowCustom: false,
          secret: false,
        },
      ],
    } as AgentPermissionRequest;

    render(
      <ApprovalDock
        onError={vi.fn()}
        permissions={[permission]}
        runtime={agentRuntime}
        sessionId="session-1"
      />,
    );

    fireEvent.click(screen.getByLabelText("Yes"));
    const submit = screen.getByRole("button", { name: "Submit answers" });
    fireEvent.click(submit);
    fireEvent.click(submit);

    expect(agentRuntime.resolvePermission).toHaveBeenCalledTimes(1);
    expect(submit).toBeDisabled();
    expect(screen.getByRole("button", { name: "Reject" })).toBeDisabled();
    complete();
    await pending;
  });

  it("clears answers when the active question request changes", () => {
    const agentRuntime = runtime();
    const first = {
      id: "question-request-1",
      kind: "question",
      title: "First question",
      questions: [
        {
          id: "decision",
          prompt: "Use the stable channel?",
          header: "Decision",
          options: [{ label: "Yes", description: "Use stable" }],
          multiple: false,
          allowCustom: false,
          secret: false,
        },
      ],
    } as AgentPermissionRequest;
    const second = {
      id: "question-request-2",
      kind: "question",
      title: "Second question",
      questions: [
        {
          id: "decision",
          prompt: "Use the preview channel?",
          header: "Decision",
          options: [{ label: "No", description: "Keep stable" }],
          multiple: false,
          allowCustom: false,
          secret: false,
        },
      ],
    } as AgentPermissionRequest;
    const props = {
      onError: vi.fn(),
      runtime: agentRuntime,
      sessionId: "session-1",
    };
    const { rerender } = render(
      <ApprovalDock {...props} permissions={[first]} />,
    );

    fireEvent.click(screen.getByLabelText("Yes"));
    expect(
      screen.getByRole("button", { name: "Submit answers" }),
    ).toBeEnabled();

    rerender(<ApprovalDock {...props} permissions={[second]} />);

    expect(screen.getByText("Use the preview channel?")).toBeVisible();
    expect(
      screen.getByRole("button", { name: "Submit answers" }),
    ).toBeDisabled();
  });
});
