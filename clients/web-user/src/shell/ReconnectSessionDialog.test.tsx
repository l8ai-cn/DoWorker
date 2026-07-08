// Unit tests for ReconnectSessionDialog — the affordance shown when the
// open session is unreachable (host offline, or not host-bound with the
// runner down).

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ReconnectSessionDialog, buildReconnectCommand } from "./ReconnectSessionDialog";

vi.mock("./ForkSessionDialog", () => ({
  ForkSessionForm: (props: {
    sourceSessionId: string;
    sourceTitle?: string | null;
    sourceWorkspace?: string | null;
    sourceHostId?: string | null;
    sourceGitBranch?: string | null;
    onClose: () => void;
  }) => (
    <div
      data-testid="fork-session-form-stub"
      data-source-session-id={props.sourceSessionId}
      data-source-title={props.sourceTitle ?? ""}
      data-source-workspace={props.sourceWorkspace ?? ""}
      data-source-host-id={props.sourceHostId ?? ""}
      data-source-git-branch={props.sourceGitBranch ?? ""}
    >
      <button type="button" data-testid="fork-session-form-close" onClick={props.onClose}>
        close
      </button>
    </div>
  ),
}));

afterEach(() => {
  cleanup();
});

describe("buildReconnectCommand", () => {
  it("emits runner run for host_offline", () => {
    const cmd = buildReconnectCommand({
      conversationId: "conv_host1",
      serverUrl: "https://example.databricksapps.com",
      state: "host_offline",
    });
    expect(cmd).toContain("runner run");
    expect(cmd).toContain("https://example.databricksapps.com");
    expect(cmd).not.toContain("--resume");
  });

  it("emits runner run for host_offline claude-native session", () => {
    const cmd = buildReconnectCommand({
      conversationId: "conv_host_claude",
      serverUrl: "https://x.databricksapps.com",
      wrapper: "claude-code-native-ui",
      state: "host_offline",
    });
    expect(cmd).toContain("runner run");
  });

  it("emits runner run for local_stranded session", () => {
    const cmd = buildReconnectCommand({
      conversationId: "conv_abc123",
      serverUrl: "https://example.databricksapps.com",
      state: "local_stranded",
    });
    expect(cmd).toContain("runner run");
    expect(cmd).toContain("conv_abc123");
    expect(cmd).toContain("https://example.databricksapps.com");
  });

  it("emits runner run for claude-native local_stranded session", () => {
    const cmd = buildReconnectCommand({
      conversationId: "conv_claude1",
      serverUrl: "https://x.databricksapps.com",
      wrapper: "claude-code-native-ui",
      state: "local_stranded",
    });
    expect(cmd).toContain("runner run");
    expect(cmd).toContain("conv_claude1");
  });
});

describe("ReconnectSessionDialog", () => {
  it("shows runner run on the Reconnect tab for host_offline owner", () => {
    render(
      <ReconnectSessionDialog
        open
        onOpenChange={() => {}}
        conversationId="conv_host1"
        serverUrl="https://example.databricksapps.com"
        state="host_offline"
        isOwner
      />,
    );
    const block = screen.getByTestId("reconnect-session-command");
    expect(block.textContent).toContain("runner run");
  });

  it("hides the command for host_offline non-owner", () => {
    render(
      <ReconnectSessionDialog
        open
        onOpenChange={() => {}}
        conversationId="conv_host1"
        serverUrl="https://example.databricksapps.com"
        state="host_offline"
        isOwner={false}
      />,
    );
    expect(screen.queryByTestId("reconnect-session-command")).toBeNull();
  });

  it("passes fork props on the Clone tab", () => {
    render(
      <ReconnectSessionDialog
        open
        onOpenChange={() => {}}
        conversationId="conv_abc"
        serverUrl="https://example.databricksapps.com"
        state="local_stranded"
        isOwner
        sourceTitle="My session"
        sourceWorkspace="/workspace"
        sourceHostId="host_1"
        sourceGitBranch="main"
      />,
    );
    fireEvent.click(screen.getByRole("tab", { name: /clone/i }));
    const stub = screen.getByTestId("fork-session-form-stub");
    expect(stub.getAttribute("data-source-session-id")).toBe("conv_abc");
    expect(stub.getAttribute("data-source-title")).toBe("My session");
  });

  it("renders nothing when closed", () => {
    const { container } = render(
      <ReconnectSessionDialog
        open={false}
        onOpenChange={() => {}}
        conversationId="conv_abc"
        serverUrl="https://example.databricksapps.com"
        state="local_stranded"
        isOwner
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });
});
