import { create } from "@bufbuild/protobuf";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  CreateGoalLoopFromPlanResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_apply_pb";
import {
  PlanResourceResponseSchema,
  ValidateResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_queries_pb";
import {
  PlanStatus,
  ResourceOperation,
} from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { act, fireEvent, render, screen, waitFor } from "@/test/test-utils";
import { createResourceDraft } from "./resource-draft-factory";

const api = vi.hoisted(() => ({
  validateResource: vi.fn(),
  planResource: vi.fn(),
  applyBindingResourcePlan: vi.fn(),
  applyExpertPlan: vi.fn(),
  applyPromptPlan: vi.fn(),
  applyWorkerTemplatePlan: vi.fn(),
  applyWorkflowPlan: vi.fn(),
  createGoalLoopFromPlan: vi.fn(),
  createWorkerFromPlan: vi.fn(),
  listResources: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({ ...api }));

import { ResourceEditorShell } from "./ResourceEditorShell";

describe("ResourceEditorShell concurrency", () => {
  beforeEach(() => {
    Object.values(api).forEach((method) => method.mockReset());
    api.listResources.mockResolvedValue({ items: [] });
  });

  it("does not switch to a stale plan after the form changes", async () => {
    const user = userEvent.setup();
    const pending = deferred<ReturnType<typeof readyPlan>>();
    api.planResource.mockReturnValue(pending.promise);
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    const name = screen.getByLabelText(/Resource name/);
    await user.type(name, "old-name");
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await user.clear(name);
    await user.type(name, "new-name");
    pending.resolve(readyPlan(
      "stale-plan",
      "2099-07-14T16:00:00Z",
      "Prompt",
    ));

    await waitFor(() => expect(api.planResource).toHaveBeenCalledOnce());
    expect(screen.getByLabelText(/Resource name/)).toHaveValue("new-name");
    expect(screen.queryByText("Plan ready")).not.toBeInTheDocument();
  });

  it("keeps newer YAML when an older validation response returns", async () => {
    const user = userEvent.setup();
    const pending = deferred<ReturnType<typeof validResponse>>();
    api.validateResource.mockReturnValue(pending.promise);
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.type(screen.getByLabelText(/Resource name/), "old-name");
    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    const oldSource = (editor as HTMLTextAreaElement).value;
    await user.click(screen.getByRole("tab", { name: "Configuration" }));
    const newSource = oldSource.replace("name: old-name", "name: new-name");
    fireEvent.change(editor, { target: { value: newSource } });
    pending.resolve(validResponse("old-name"));

    await waitFor(() => expect(api.validateResource).toHaveBeenCalledOnce());
    expect(screen.getByTestId("resource-yaml-editor")).toHaveValue(newSource);
  });

  it("does not close over a newer draft when an old apply completes", async () => {
    const user = userEvent.setup();
    const pending = deferred<ReturnType<typeof goalLoopResult>>();
    const onApplied = vi.fn();
    api.planResource.mockResolvedValue(readyPlan(
      "goal-loop-plan",
      "2099-07-14T16:00:00Z",
      "GoalLoop",
    ));
    api.createGoalLoopFromPlan.mockReturnValue(pending.promise);
    render(
      <ResourceEditorShell
        orgSlug="acme"
        kind="GoalLoop"
        onApplied={onApplied}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("tab", { name: "Configuration" }));
    const objective = screen.getByLabelText(/Objective/);
    await user.click(screen.getByRole("button", { name: "Apply resource" }));
    await waitFor(() => {
      expect(api.createGoalLoopFromPlan).toHaveBeenCalledOnce();
    });
    fireEvent.change(objective, { target: { value: "new objective" } });
    pending.resolve(goalLoopResult());

    await waitFor(() => expect(objective).toHaveValue("new objective"));
    expect(onApplied).not.toHaveBeenCalled();
  });

  it("accepts a successful apply response after the local plan expiry time", async () => {
    const user = userEvent.setup();
    const pending = deferred<ReturnType<typeof goalLoopResult>>();
    const onApplied = vi.fn();
    const expiresAt = new Date(Date.now() + 1_000).toISOString();
    api.planResource.mockResolvedValue(readyPlan(
      "expiring-goal-loop-plan",
      expiresAt,
      "GoalLoop",
    ));
    api.createGoalLoopFromPlan.mockReturnValue(pending.promise);
    render(
      <ResourceEditorShell
        orgSlug="acme"
        kind="GoalLoop"
        onApplied={onApplied}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await act(async () => {});
    expect(screen.getByText("Plan ready")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Apply resource" }));
    await new Promise((resolve) => window.setTimeout(resolve, 1_050));
    await act(async () => {
      pending.resolve(goalLoopResult());
      await Promise.resolve();
    });

    expect(onApplied).toHaveBeenCalledOnce();
  });
});

function readyPlan(
  planId: string,
  expiresAt = "2099-07-14T16:00:00Z",
  kind: Parameters<typeof createResourceDraft>[0] = "WorkerTemplate",
) {
  const draft = createResourceDraft(kind, "acme");
  draft.metadata.name = "planned-resource";
  return create(PlanResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    canonicalJson: new TextEncoder().encode(JSON.stringify(draft)),
    plan: {
      planId,
      operation: ResourceOperation.CREATE,
      expiresAt,
      status: PlanStatus.PENDING,
    },
  });
}

function validResponse(name: string) {
  const draft = createResourceDraft("WorkerTemplate", "acme");
  draft.metadata.name = name;
  return create(ValidateResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    canonicalJson: new TextEncoder().encode(JSON.stringify(draft)),
  });
}

function goalLoopResult() {
  return create(CreateGoalLoopFromPlanResponseSchema, {
    resource: { revision: 1n },
    goalLoopId: 7n,
    workerSpecSnapshotId: 8n,
    resourceRevision: 1n,
  });
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((fulfill) => {
    resolve = fulfill;
  });
  return { promise, resolve };
}
