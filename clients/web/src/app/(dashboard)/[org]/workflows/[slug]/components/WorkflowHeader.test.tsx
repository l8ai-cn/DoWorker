import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { WorkflowData } from "@/stores/workflow";
import { WorkflowHeader } from "./WorkflowHeader";

const workflow = {
  slug: "nightly-review",
  name: "Nightly review",
  status: "enabled",
  execution_mode: "direct",
  active_run_count: 0,
  max_concurrent_runs: 1,
} as WorkflowData;

describe("WorkflowHeader", () => {
  it("offers a resource revision and no direct edit or delete action", () => {
    const onRevise = vi.fn();

    render(
      <WorkflowHeader
        workflow={workflow}
        triggering={false}
        t={(key) => key}
        onTrigger={() => {}}
        onRevise={onRevise}
        onEnable={() => {}}
        onDisable={() => {}}
      />,
    );

    fireEvent.click(screen.getByRole("button", {
      name: "workflows.newRevision",
    }));

    expect(onRevise).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("button", { name: "common.edit" })).not.toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: "common.delete" })).not.toBeInTheDocument();
  });
});
