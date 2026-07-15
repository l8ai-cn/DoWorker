import { beforeEach, describe, expect, it, vi } from "vitest";

const { createGoalLoopFromPlan } = vi.hoisted(() => ({
  createGoalLoopFromPlan: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  applyBindingResourcePlan: vi.fn(),
  applyExpertPlan: vi.fn(),
  applyPromptPlan: vi.fn(),
  applyWorkerTemplatePlan: vi.fn(),
  applyWorkflowPlan: vi.fn(),
  createGoalLoopFromPlan,
  createWorkerFromPlan: vi.fn(),
}));

import { applyResourcePlan } from "./apply-resource-plan";
import type { ResourceEditorKind } from "./resource-editor-types";

describe("applyResourcePlan", () => {
  beforeEach(() => {
    createGoalLoopFromPlan.mockReset();
  });

  it("uses the typed GoalLoop creation endpoint", async () => {
    createGoalLoopFromPlan.mockResolvedValue({ goalLoopId: 83n });

    await expect(applyResourcePlan("acme", "GoalLoop", "plan-1"))
      .resolves.toEqual({ goalLoopId: 83n });
    expect(createGoalLoopFromPlan).toHaveBeenCalledWith("acme", "plan-1");
  });

  it("fails explicitly for an unsupported runtime kind", () => {
    expect(() =>
      applyResourcePlan("acme", "Unsupported" as ResourceEditorKind, "plan-1"),
    ).toThrow("Unsupported resource kind: Unsupported");
  });
});
