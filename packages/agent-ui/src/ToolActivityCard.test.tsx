import { render, screen, within } from "@testing-library/react";

import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import { ToolActivityCard } from "./ToolActivityCard";

describe("ToolActivityCard", () => {
  it("uses a renderer only for the exact tool identity", () => {
    const registry =
      new ToolRendererRegistry<AgentToolRendererRegistration>();
    registry.register(
      {
        namespace: "agentsmesh.acp",
        semanticKey: "generate_image",
        schemaVersion: "1",
      },
      {
        summary: ({ item }) => (
          <div>图片任务：{String(item.inputValue)}</div>
        ),
      },
      "builtin.image",
    );

    const { rerender } = render(
      <ToolActivityCard
        item={tool({
          identity: {
            namespace: "agentsmesh.acp",
            semanticKey: "generate_image",
            schemaVersion: "1",
          },
          inputValue: "生成产品主图",
          title: "generate_image",
        })}
        renderers={registry}
      />,
    );

    expect(screen.getByText("图片任务：生成产品主图")).toBeVisible();

    rerender(
      <ToolActivityCard
        item={tool({
          identity: {
            namespace: "agentsmesh.acp",
            semanticKey: "generate_image",
            schemaVersion: "2",
          },
          input: "raw input",
          title: "generate_image",
        })}
        renderers={registry}
      />,
    );

    expect(screen.queryByText("图片任务：生成产品主图")).not.toBeInTheDocument();
    expect(screen.getAllByText("raw input")[0]).toBeVisible();
  });

  it("summarizes completed file changes and keeps raw evidence collapsed", () => {
    render(
      <ToolActivityCard
        item={tool({
          id: "file-1",
          title: "fileChange",
          input: JSON.stringify({
            changes: [
              {
                path: "/workspace/project/boundary.txt",
                kind: { type: "add" },
                diff: "BOUNDARY_OK\n",
              },
            ],
          }),
          output: "Add /workspace/project/boundary.txt",
          status: "completed",
        })}
      />,
    );

    expect(screen.getByText("Added boundary.txt")).toBeVisible();
    expect(
      within(screen.getByTestId("tool-summary")).getByText(
        "Add /workspace/project/boundary.txt",
      ),
    ).toBeVisible();
    expect(screen.getByText("Details").closest("details")).not.toHaveAttribute("open");
  });

  it("shows command and output summaries without expanding raw JSON", () => {
    render(
      <ToolActivityCard
        item={tool({
          id: "shell-1",
          title: "shell",
          input: JSON.stringify({
            command: "/bin/bash -lc 'cat boundary.txt'",
            cwd: "/workspace/project",
          }),
          output: "BOUNDARY_OK\n",
          status: "completed",
        })}
      />,
    );

    expect(screen.getByText("/bin/bash -lc 'cat boundary.txt'")).toBeVisible();
    expect(
      within(screen.getByTestId("tool-summary")).getByText("BOUNDARY_OK"),
    ).toBeVisible();
    expect(screen.getByText("Details").closest("details")).not.toHaveAttribute("open");
  });

  it("opens failed tool evidence so the error is immediately inspectable", () => {
    render(
      <ToolActivityCard
        item={tool({
          id: "shell-2",
          title: "shell",
          input: "pnpm test",
          output: "exit status 1",
          status: "failed",
        })}
      />,
    );

    expect(screen.getByText("Details").closest("details")).toHaveAttribute("open");
  });
});

function tool(
  patch: Partial<Parameters<typeof ToolActivityCard>[0]["item"]> = {},
): Parameters<typeof ToolActivityCard>[0]["item"] {
  return {
    id: "tool-1",
    identity: {
      namespace: "agentsmesh.acp",
      semanticKey: "shell",
      schemaVersion: "1",
    },
    kind: "tool",
    results: [],
    status: "completed",
    title: "shell",
    ...patch,
  };
}
