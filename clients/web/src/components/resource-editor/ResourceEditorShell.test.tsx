import { create } from "@bufbuild/protobuf";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  PlanStatus,
  PlanResourceResponseSchema,
  ResourceOperation,
  ValidateResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { fireEvent, render, screen } from "@/test/test-utils";
import type { AsyncState } from "@/components/pod/hooks/workerCreateDraft";
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
const options = vi.hoisted(() => ({
  state: { status: "loading" } as AsyncState<WorkerCreateOptions>,
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({
  ...api,
}));

vi.mock("@/components/pod/hooks/useWorkerCreateOptions", () => ({
  useWorkerCreateOptions: () => options.state,
  loadWorkerCreateOptions: async () => {
    if (options.state.status !== "ready") {
      throw new Error("Worker options are unavailable.");
    }
    return options.state.data;
  },
}));

import { ResourceEditorShell } from "./ResourceEditorShell";

describe("ResourceEditorShell", () => {
  beforeEach(() => {
    Object.values(api).forEach((method) => method.mockReset());
    api.listResources.mockResolvedValue({ items: [] });
    options.state = { status: "ready", data: workerOptions() };
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

  it("adopts edited YAML before planning from the Plan tab", async () => {
    const user = userEvent.setup();
    const canonical = createResourceDraft("Prompt", "acme");
    canonical.metadata.name = "yaml-prompt";
    canonical.spec.content = "Content from YAML";
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      canonicalJson: new TextEncoder().encode(JSON.stringify(canonical)),
      plan: {
        planId: "yaml-plan",
        operation: ResourceOperation.CREATE,
        expiresAt: "2099-07-14T16:00:00Z",
        status: PlanStatus.PENDING,
      },
    }));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.type(screen.getByLabelText(/Resource name/), "yaml-prompt");
    await user.type(screen.getByLabelText(/Prompt content/), "Form content");
    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          "Form content",
          "Content from YAML",
        ),
      },
    });
    await user.click(screen.getByRole("tab", { name: "Plan & diff" }));
    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    await screen.findByText("Plan ready");
    const document = api.planResource.mock.calls[0][1];
    expect(JSON.parse(document.content).spec.content).toBe("Content from YAML");
  });

  it("blocks Plan while Worker options are unavailable", async () => {
    const user = userEvent.setup();
    options.state = { status: "loading" };
    render(<ResourceEditorShell orgSlug="acme" />);

    const plan = screen.getByRole("button", { name: "Generate plan" });
    expect(plan).toBeDisabled();
    await user.click(plan);
    expect(api.planResource).not.toHaveBeenCalled();
  });

  it("does not plan YAML with a Worker type outside the organization catalog", async () => {
    const user = userEvent.setup();
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.type(screen.getByLabelText(/Resource name/), "reviewer");
    await user.click(screen.getByRole("button", { name: "Worker type" }));
    await user.click(screen.getByRole("option", { name: "Codex CLI" }));
    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          "workerType: codex-cli",
          "workerType: unavailable-worker",
        ),
      },
    });
    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText(
      "The selected Worker type is unavailable.",
    )).toBeInTheDocument();
    expect(api.planResource).not.toHaveBeenCalled();
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

  it("adopts the Plan canonical draft before returning to the form", async () => {
    const user = userEvent.setup();
    const canonical = createResourceDraft("WorkerTemplate", "acme");
    canonical.metadata.name = "canonical-reviewer";
    canonical.metadata.displayName = "Canonical reviewer";
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      canonicalJson: new TextEncoder().encode(JSON.stringify(canonical)),
      plan: {
        planId: "canonical-plan",
        operation: ResourceOperation.CREATE,
        expiresAt: "2099-07-14T16:00:00Z",
        status: PlanStatus.PENDING,
      },
    }));
    render(<ResourceEditorShell orgSlug="acme" />);

    await user.click(screen.getByRole("tab", { name: "YAML" }));
    const editor = await screen.findByTestId("resource-yaml-editor");
    fireEvent.change(editor, {
      target: {
        value: (editor as HTMLTextAreaElement).value.replace(
          'name: ""',
          "name: yaml-reviewer",
        ),
      },
    });
    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await screen.findByText("Plan ready");
    await user.click(screen.getByRole("tab", { name: "Configuration" }));

    expect(screen.getByLabelText(/Resource name/))
      .toHaveValue("canonical-reviewer");
    expect(screen.getByLabelText(/Display name/))
      .toHaveValue("Canonical reviewer");
  });

  it("resets the local draft when the organization identity changes", async () => {
    const user = userEvent.setup();
    const view = render(
      <ResourceEditorShell orgSlug="acme" kind="Prompt" />,
    );
    await user.type(screen.getByLabelText(/Resource name/), "acme-prompt");

    view.rerender(
      <ResourceEditorShell orgSlug="globex" kind="Prompt" />,
    );

    expect(screen.getByLabelText(/Resource name/)).toHaveValue("");
    await user.click(screen.getByRole("tab", { name: "YAML" }));
    expect(
      (await screen.findByTestId(
        "resource-yaml-editor",
      ) as HTMLTextAreaElement).value,
    ).toContain("namespace: globex");
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
});

function readyPlan(planId: string) {
  const draft = createResourceDraft("GoalLoop", "acme");
  draft.metadata.name = "goal-loop";
  return create(PlanResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    canonicalJson: new TextEncoder().encode(JSON.stringify(draft)),
    plan: {
      planId,
      operation: ResourceOperation.CREATE,
      expiresAt: "2099-07-14T16:00:00Z",
      status: PlanStatus.PENDING,
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
      model_protocol_adapters: [],
      tool_model_requirements: [],
      credential_requirements: [],
      config_document_requirements: [],
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
