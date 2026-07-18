import { render, screen, within } from "@testing-library/react";

import { ToolActivityCard } from "./ToolActivityCard";

describe("ToolActivityCard", () => {
  it("summarizes completed file changes and keeps raw evidence collapsed", () => {
    render(
      <ToolActivityCard
        item={{
          id: "file-1",
          kind: "tool",
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
        }}
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
        item={{
          id: "shell-1",
          kind: "tool",
          title: "shell",
          input: JSON.stringify({
            command: "/bin/bash -lc 'cat boundary.txt'",
            cwd: "/workspace/project",
          }),
          output: "BOUNDARY_OK\n",
          status: "completed",
        }}
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
        item={{
          id: "shell-2",
          kind: "tool",
          title: "shell",
          input: "pnpm test",
          output: "exit status 1",
          status: "failed",
        }}
      />,
    );

    expect(screen.getByText("Details").closest("details")).toHaveAttribute("open");
  });
});
