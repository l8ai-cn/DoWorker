import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@/test/test-utils";
import { GoalLoopPage } from "../GoalLoopPage";
import type { GoalLoopData } from "@/lib/viewModels/goal-loop";

const {
  mockCancelGoalLoop,
  mockListGoalLoops,
  mockStartGoalLoop,
  mockVerifyGoalLoop,
} = vi.hoisted(() => ({
  mockCancelGoalLoop: vi.fn(),
  mockListGoalLoops: vi.fn(),
  mockStartGoalLoop: vi.fn(),
  mockVerifyGoalLoop: vi.fn(),
}));

vi.mock("@/lib/api/facade/goalLoopConnect", () => ({
  cancelGoalLoop: mockCancelGoalLoop,
  listGoalLoops: mockListGoalLoops,
  startGoalLoop: mockStartGoalLoop,
  verifyGoalLoop: mockVerifyGoalLoop,
}));

vi.mock("@/components/resource-editor/ResourceEditorShell", () => ({
  ResourceEditorShell: ({
    kind,
    orgSlug,
    onApplied,
  }: {
    kind: string;
    orgSlug: string;
    onApplied: () => void;
  }) => (
    <button
      type="button"
      data-kind={kind}
      data-org={orgSlug}
      onClick={onApplied}
    >
      Apply GoalLoop
    </button>
  ),
}));

function goalLoop(name: string, status: GoalLoopData["status"]): GoalLoopData {
  return {
    id: status === "active" ? 2 : 1,
    slug: name.toLowerCase().replaceAll(" ", "-"),
    name,
    worker_spec_snapshot_id: 3,
    objective: `${name} objective`,
    acceptance_criteria: ["tests pass"],
    verification_command: "pnpm test",
    status,
    max_iterations: 5,
    timeout_minutes: 60,
    no_progress_limit: 2,
    same_error_limit: 2,
    escalation_policy: "pause",
    verification_output_truncated: false,
    created_at: "2026-07-12T00:00:00Z",
    updated_at: "2026-07-12T00:00:00Z",
  };
}

describe("GoalLoopPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListGoalLoops.mockResolvedValue([
      goalLoop("Draft migration", "draft"),
      goalLoop("Active release check", "active"),
    ]);
  });

  it("keeps creation on demand and prioritizes executing loops", async () => {
    render(<GoalLoopPage orgSlug="dev-org" />);

    await screen.findByRole("heading", { name: "Active release check" });

    expect(screen.queryByText("创建目标 Loop")).not.toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Active release check" })
        .compareDocumentPosition(screen.getByRole("heading", { name: "Draft migration" }))
        & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "新建 Loop" }));

    await waitFor(() => {
      expect(screen.getByRole("dialog", { name: "创建目标 Loop" })).toBeInTheDocument();
    });
    const apply = screen.getByRole("button", { name: "Apply GoalLoop" });
    expect(apply).toHaveAttribute("data-kind", "GoalLoop");
    expect(apply).toHaveAttribute("data-org", "dev-org");

    fireEvent.click(apply);

    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "创建目标 Loop" }))
        .not.toBeInTheDocument();
      expect(mockListGoalLoops).toHaveBeenCalledTimes(2);
    });
  });
});
