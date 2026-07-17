import { create } from "@bufbuild/protobuf";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  ApplyExpertPlanResponseSchema,
  ApplyWorkerTemplatePlanResponseSchema,
  ApplyWorkflowPlanResponseSchema,
  CreateWorkerFromPlanResponseSchema,
  PlanResourceResponseSchema,
  ResourceSchema,
  ResourceOperation,
  ValidateResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { fireEvent, render, screen, waitFor } from "@/test/test-utils";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { createResourceDraft } from "./resource-draft-factory";

const api = vi.hoisted(() => ({
  validateResource: vi.fn(),
  planResource: vi.fn(),
  applyBindingResourcePlan: vi.fn(),
  applyExpertPlan: vi.fn(),
  applyPromptPlan: vi.fn(),
  applyWorkerTemplatePlan: vi.fn(),
  applyWorkflowPlan: vi.fn(),
  createWorkerFromPlan: vi.fn(),
  listResources: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  ...api,
}));

vi.mock("@/components/pod/hooks/useWorkerCreateOptions", () => ({
  useWorkerCreateOptions: () => ({
    status: "ready",
    data: workerOptions(),
  }),
}));

import { ResourceEditorShell } from "./ResourceEditorShell";

describe("ResourceEditorShell", () => {
  beforeEach(() => {
    Object.values(api).forEach((method) => method.mockReset());
    api.listResources.mockResolvedValue({ items: [] });
    api.validateResource.mockResolvedValue(create(
      ValidateResourceResponseSchema,
      {
        operation: ResourceOperation.CREATE,
        canonicalJson: new TextEncoder().encode(validManifestJson()),
      },
    ));
  });

  it("shows form edits in the YAML view from the same draft", async () => {
    const user = userEvent.setup();
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.type(screen.getByLabelText(/Resource name/), "code-reviewer");
    await user.click(screen.getByRole("button", { name: "Worker type" }));
    await user.click(screen.getByRole("option", { name: "Codex CLI" }));
    await user.click(screen.getByRole("tab", { name: "YAML" }));

    const editor = await screen.findByTestId("resource-yaml-editor");
    expect((editor as HTMLTextAreaElement).value).toContain(
      "name: code-reviewer",
    );
    expect((editor as HTMLTextAreaElement).value).toContain(
      "workerType: codex-cli",
    );
  });

  it("keeps invalid YAML visible and blocks switching back to the form", async () => {
    const user = userEvent.setup();
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, { target: { value: "kind: [" } });
    await user.click(screen.getByRole("tab", { name: "Configuration" }));

    expect(screen.getByTestId("resource-yaml-editor")).toHaveValue("kind: [");
    expect(screen.getByText(
      "Fix YAML before returning to the form or applying.",
    )).toBeInTheDocument();
    expect(api.validateResource).not.toHaveBeenCalled();
  });

  it("returns valid YAML to the form using the canonical draft", async () => {
    const user = userEvent.setup();
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          'name: ""',
          "name: yaml-name",
        ),
      },
    });
    await user.click(screen.getByRole("tab", { name: "Configuration" }));

    expect(await screen.findByLabelText(/Resource name/))
      .toHaveValue("code-reviewer");
  });

  it("does not plan GoalLoop YAML with invalid integer fields", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(readyPlan("invalid-goal-loop-plan"));
    render(<ResourceEditorShell orgSlug="acme" kind="GoalLoop" />);

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    const invalidSource = (editor as HTMLTextAreaElement).value.replace(
      "maxIterations: 10",
      "maxIterations: 101",
    );
    fireEvent.change(editor, { target: { value: invalidSource } });
    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findAllByText(
      "GoalLoop YAML contains invalid integer fields.",
    )).not.toHaveLength(0);
    expect(api.planResource).not.toHaveBeenCalled();
    expect(screen.getByTestId("resource-yaml-editor")).toHaveValue(invalidSource);
  });

  it("does not validate invalid GoalLoop integers when returning to the form", async () => {
    const user = userEvent.setup();
    render(<ResourceEditorShell orgSlug="acme" kind="GoalLoop" />);

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          "maxIterations: 10",
          "maxIterations: 1.5",
        ),
      },
    });
    await user.click(screen.getByRole("tab", { name: "Configuration" }));

    expect(await screen.findAllByText(
      "GoalLoop YAML contains invalid integer fields.",
    )).not.toHaveLength(0);
    expect(api.validateResource).not.toHaveBeenCalled();
  });

  it("applies only the current WorkerTemplate plan", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      plan: {
        planId: "11111111-1111-4111-8111-111111111111",
        operation: ResourceOperation.CREATE,
        expiresAt: "2099-07-14T16:00:00Z",
      },
    }));
    api.applyWorkerTemplatePlan.mockResolvedValue(create(
      ApplyWorkerTemplatePlanResponseSchema,
      {
        resource: { revision: 1n },
        workerSpecSnapshotId: 9n,
      },
    ));
    render(<ResourceEditorShell orgSlug="acme" />);

    const apply = screen.getByRole("button", { name: "Apply template" });
    expect(apply).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    await screen.findByText("Plan ready");
    expect(apply).toBeEnabled();
    await user.click(apply);

    await waitFor(() => {
      expect(api.applyWorkerTemplatePlan).toHaveBeenCalledWith(
        "acme",
        "11111111-1111-4111-8111-111111111111",
      );
    });
    expect(await screen.findByText(/WorkerSpec snapshot 9/)).toBeInTheDocument();
  });

  it("creates a Worker only through its typed plan apply", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      plan: {
        planId: "22222222-2222-4222-8222-222222222222",
        operation: ResourceOperation.CREATE,
        expiresAt: "2099-07-14T16:00:00Z",
      },
    }));
    api.createWorkerFromPlan.mockResolvedValue(create(
      CreateWorkerFromPlanResponseSchema,
      {
        resource: { revision: 1n },
        launchId: 7n,
        podId: 8n,
        podKey: "worker-abcd",
        workerSpecSnapshotId: 9n,
        resourceRevision: 1n,
        runnerId: 10n,
      },
    ));
    render(<ResourceEditorShell orgSlug="acme" kind="Worker" />);

    await user.type(screen.getByLabelText(/Resource name/), "review-run");
    await user.type(
      screen.getByRole("combobox", { name: /Worker template/ }),
      "code-reviewer",
    );
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("button", { name: "Create Worker" }));

    await waitFor(() => {
      expect(api.createWorkerFromPlan).toHaveBeenCalledWith(
        "acme",
        "22222222-2222-4222-8222-222222222222",
      );
    });
    expect(api.applyWorkerTemplatePlan).not.toHaveBeenCalled();
    expect(await screen.findByText(/WorkerSpec snapshot 9/)).toBeInTheDocument();
  });

  it("applies an Expert through the Expert typed apply", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(readyPlan("expert-plan"));
    api.applyExpertPlan.mockResolvedValue(create(
      ApplyExpertPlanResponseSchema,
      {
        resource: { revision: 2n },
        expertId: 12n,
        workerSpecSnapshotId: 13n,
        resourceRevision: 2n,
      },
    ));
    render(<ResourceEditorShell orgSlug="acme" kind="Expert" />);

    await user.type(screen.getByLabelText(/Resource name/), "reviewer");
    await user.type(
      screen.getByRole("combobox", { name: /Worker template/ }),
      "code-reviewer",
    );
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("button", { name: "Apply resource" }));

    await waitFor(() => {
      expect(api.applyExpertPlan).toHaveBeenCalledWith("acme", "expert-plan");
    });
    expect(await screen.findByText(/WorkerSpec snapshot 13/)).toBeInTheDocument();
  });

  it("applies a Workflow through the Workflow typed apply", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(readyPlan("workflow-plan"));
    api.applyWorkflowPlan.mockResolvedValue(create(
      ApplyWorkflowPlanResponseSchema,
      {
        resource: { revision: 3n },
        workflowId: 14n,
        workerSpecSnapshotId: 15n,
        resourceRevision: 3n,
      },
    ));
    render(<ResourceEditorShell orgSlug="acme" kind="Workflow" />);

    await user.type(screen.getByLabelText(/Resource name/), "nightly-review");
    await user.type(
      screen.getByRole("combobox", { name: /Worker template/ }),
      "code-reviewer",
    );
    await user.type(
      screen.getByRole("combobox", { name: /^Prompt/ }),
      "review-prompt",
    );
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("button", { name: "Apply resource" }));

    await waitFor(() => {
      expect(api.applyWorkflowPlan).toHaveBeenCalledWith(
        "acme",
        "workflow-plan",
      );
    });
  });

  it("applies Prompt and binding resources through their typed apply RPCs", async () => {
    const user = userEvent.setup();
    api.planResource
      .mockResolvedValueOnce(readyPlan("prompt-plan"))
      .mockResolvedValueOnce(readyPlan("binding-plan"));
    api.applyPromptPlan.mockResolvedValue(create(ResourceSchema, {
      revision: 4n,
    }));
    api.applyBindingResourcePlan.mockResolvedValue(create(ResourceSchema, {
      revision: 5n,
    }));
    const prompt = render(
      <ResourceEditorShell orgSlug="acme" kind="Prompt" />,
    );

    await user.type(screen.getByLabelText(/Resource name/), "review-prompt");
    await user.type(screen.getByLabelText(/Prompt content/), "Review changes");
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("button", { name: "Apply resource" }));
    await waitFor(() => {
      expect(api.applyPromptPlan).toHaveBeenCalledWith("acme", "prompt-plan");
    });
    prompt.unmount();

    render(<ResourceEditorShell orgSlug="acme" kind="ModelBinding" />);
    await user.type(screen.getByLabelText(/Resource name/), "coding-primary");
    await user.type(screen.getByLabelText(/Model API resource ID/), "101");
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("button", { name: "Apply resource" }));
    await waitFor(() => {
      expect(api.applyBindingResourcePlan).toHaveBeenCalledWith(
        "acme",
        "binding-plan",
      );
    });
  });
});

function readyPlan(planId: string) {
  return create(PlanResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    plan: {
      planId,
      operation: ResourceOperation.CREATE,
      expiresAt: "2099-07-14T16:00:00Z",
    },
  });
}

function validManifestJson(): string {
  const draft = createResourceDraft("WorkerTemplate", "acme");
  return JSON.stringify({
    ...draft,
    metadata: { ...draft.metadata, name: "code-reviewer" },
  });
}

function workerOptions(): WorkerCreateOptions {
  return {
    revision: "catalog-current",
    worker_types: [{
      slug: "codex-cli",
      name: "Codex CLI",
      description: "",
      schema_version: 1,
      config_schema: {},
      supported_interaction_modes: ["pty", "acp"],
      requires_model_resource: false,
      tool_model_requirements: [],
      selectable: true,
      blocking_reason: "",
    }],
    runtime_images: [{
      id: 11,
      slug: "codex",
      name: "Codex stable",
      reference: "",
      digest: "",
      worker_type_slugs: ["codex-cli"],
      selectable: true,
      blocking_reason: "",
    }],
    compute_targets: [],
    deployment_modes: [],
    resource_profiles: [],
  };
}
