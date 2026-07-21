import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { AgentSessionRuntime } from "./contracts";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import { ActivityTimeline } from "./ActivityTimeline";
import { AgentWorkspaceLocaleProvider } from "./AgentWorkspaceLocaleContext";
import { createBuiltinToolRenderers } from "./builtinToolRenderers";

describe("ActivityTimeline", () => {
  it("collapses consecutive tool activity and reveals individual steps on demand", () => {
    render(
      <ActivityTimeline
        items={[
          {
            ...toolContract("shell"),
            id: "tool-1",
            kind: "tool",
            title: "shell",
            input: JSON.stringify({ command: "pnpm test" }),
            output: "12 tests passed",
            status: "completed",
          },
          {
            ...toolContract("shell"),
            id: "tool-2",
            kind: "tool",
            title: "shell",
            input: JSON.stringify({ command: "pnpm lint" }),
            output: "No warnings",
            status: "completed",
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
      />,
    );

    const summary = screen.getByText("Used shell 2 times");
    const group = summary.closest("details");

    expect(group).not.toHaveAttribute("open");
    expect(screen.getByText(/pnpm test/)).not.toBeVisible();
    for (const output of screen.getAllByText("12 tests passed")) {
      expect(output).not.toBeVisible();
    }

    fireEvent.click(within(group!).getByText("Used shell 2 times"));

    expect(group).toHaveAttribute("open");
    expect(screen.getByText(/pnpm test/)).not.toBeVisible();
    fireEvent.click(within(group!).getAllByText("Details")[0]!);
    expect(screen.getByText(/pnpm test/)).toBeVisible();
    expect(screen.getByText(/pnpm lint/)).not.toBeVisible();
  });

  it("keeps failed tool groups collapsed while surfacing the failed status", () => {
    render(
      <ActivityTimeline
        items={[
          {
            ...toolContract("shell"),
            id: "tool-failed",
            kind: "tool",
            title: "shell",
            input: JSON.stringify({ command: "pnpm test" }),
            output: "exit status 1",
            status: "failed",
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
      />,
    );

    const group = screen.getByText("Used shell 1 time").closest("details");
    const groupSummary = group!.querySelector("summary")!;

    expect(group).not.toHaveAttribute("open");
    expect(within(groupSummary).getByText("Failed")).toBeVisible();
    expect(screen.getByText(/pnpm test/)).not.toBeVisible();
  });

  it("surfaces both running and failed states in a mixed tool group", () => {
    render(
      <ActivityTimeline
        items={[
          {
            ...toolContract("shell"),
            id: "tool-failed",
            kind: "tool",
            title: "shell",
            status: "failed",
          },
          {
            ...toolContract("shell"),
            id: "tool-running",
            kind: "tool",
            title: "shell",
            status: "running",
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
      />,
    );

    const group = screen.getByText("Used shell 2 times").closest("details");
    const groupSummary = group!.querySelector("summary")!;

    expect(within(groupSummary).getByText("Failed")).toBeVisible();
    expect(within(groupSummary).getByText("Running")).toBeVisible();
  });

  it("presents pending tool activity as actively running", () => {
    render(
      <ActivityTimeline
        items={[
          {
            ...toolContract("shell"),
            id: "tool-pending",
            kind: "tool",
            title: "shell",
            status: "pending",
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
      />,
    );

    const group = screen.getByText("Used shell 1 time").closest("details");
    const groupSummary = group!.querySelector("summary")!;

    expect(within(groupSummary).getByText("Running")).toBeVisible();
  });

  it("uses a renderer presentation instead of the reported title", () => {
    const renderers = new ToolRendererRegistry<AgentToolRendererRegistration>();
    renderers.register(
      toolContract("shell").identity,
      { presentation: { label: "Exact shell" } },
      "test.shell",
    );

    render(
      <ActivityTimeline
        items={[
          {
            ...toolContract("shell"),
            id: "tool-exact",
            kind: "tool",
            title: "generate_image",
            status: "completed",
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
        toolRenderers={renderers}
      />,
    );

    expect(screen.getByText("Used Exact shell 1 time")).toBeVisible();
  });

  it.each([
    ["en-US", "Changed 1 file"],
    ["zh-CN", "修改了 1 个文件"],
  ] as const)("localizes builtin file changes in %s", (locale, summary) => {
    render(
      <AgentWorkspaceLocaleProvider locale={locale}>
        <ActivityTimeline
          items={[{
            ...toolContract("filesystem.edit"),
            id: "tool-edit",
            kind: "tool",
            status: "completed",
            title: "Edit",
          }]}
          runtime={{} as AgentSessionRuntime}
          sessionId="session-1"
          toolRenderers={createBuiltinToolRenderers()}
        />
      </AgentWorkspaceLocaleProvider>,
    );

    expect(screen.getByText(summary)).toBeVisible();
  });

  it("shows failed builtin file-change output after expanding the group", () => {
    render(
      <ActivityTimeline
        items={[{
          ...toolContract("filesystem.edit"),
          id: "tool-edit-failed",
          kind: "tool",
          output: "file not found",
          status: "failed",
          title: "Edit",
        }]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
        toolRenderers={createBuiltinToolRenderers()}
      />,
    );

    const group = screen.getByText("Changed 1 file").closest("details");
    fireEvent.click(within(group!).getByText("Changed 1 file"));

    expect(screen.getByText("file not found")).toBeVisible();
  });

  it("wraps long user paths inside narrow embedded workspaces", () => {
    const text =
      "Copy /workspace/repos/sandboxes/standalone-session/deliverables/agentic-showcase";

    render(
      <ActivityTimeline
        items={[
          {
            id: "user-1",
            kind: "message",
            role: "user",
            status: "completed",
            text,
          },
        ]}
        runtime={{} as AgentSessionRuntime}
        sessionId="session-1"
      />,
    );

    expect(screen.getByText(text)).toHaveClass("break-words");
  });
});

function toolContract(semanticKey: string) {
  return {
    identity: {
      namespace: "agentcloud.acp",
      schemaVersion: "1",
      semanticKey,
    },
    results: [],
  } as const;
}
