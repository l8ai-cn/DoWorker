import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import type {
  WorkerCreateOptions,
  WorkerSpecDraft,
} from "@/lib/api/facade/podConnect";
import {
  createInitialWorkerDraftState,
  workerCreateDraftReducer,
} from "../../hooks/workerCreateDraft";
import { workerCreateValidity } from "../../hooks/workerCreateValidity";
import { WorkerCreateStepper } from "../WorkerCreateStepper";
import { WorkerRuntimeStep } from "../WorkerRuntimeStep";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const t = (key: string) => key;

describe("Worker create flow", () => {
  it("renders four named steps and prevents access to incomplete later steps", () => {
    const onChange = vi.fn();

    render(
      <WorkerCreateStepper
        current={1}
        onChange={onChange}
        steps={[
          step(1, "workerCreate.steps.runtime", true, true),
          step(2, "workerCreate.steps.typeConfig", false, false),
          step(3, "workerCreate.steps.workspace", false, false),
          step(4, "workerCreate.steps.preflight", false, false),
        ]}
      />,
    );

    expect(screen.getByText("workerCreate.steps.runtime")).toBeInTheDocument();
    expect(screen.getByText("workerCreate.steps.typeConfig")).toBeInTheDocument();
    expect(screen.getByText("workerCreate.steps.workspace")).toBeInTheDocument();
    expect(screen.getByText("workerCreate.steps.preflight")).toBeInTheDocument();

    fireEvent.click(screen.getByText("workerCreate.steps.preflight"));
    expect(onChange).not.toHaveBeenCalled();
  });

  it("renders runtime choices in the approved order and explains disabled options", () => {
    const { container } = render(
      <WorkerRuntimeStep
        draft={completeDraft()}
        modelResources={{ status: "ready", data: [modelResource()] }}
        toolModelResources={{ status: "ready", data: [] }}
        options={{ status: "ready", data: createOptions() }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={t}
      />,
    );

    expect(
      Array.from(container.querySelectorAll("[data-runtime-field]")).map(
        (node) => node.getAttribute("data-runtime-field"),
      ),
    ).toEqual([
      "model",
      "worker-type",
      "runtime-image",
      "compute-target",
      "deployment-mode",
      "resource-profile",
    ]);

    fireEvent.click(screen.getByLabelText("workerCreate.runtime.computeTarget"));
    expect(
      screen.getByText("workerCreate.runtime.options.dedicatedUnavailable"),
    ).toBeInTheDocument();
  });

  it("renders required option-loading failures as errors", () => {
    render(
      <WorkerRuntimeStep
        draft={completeDraft()}
        modelResources={{ status: "ready", data: [modelResource()] }}
        toolModelResources={{ status: "ready", data: [] }}
        options={{ status: "error", error: "catalog unavailable" }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={t}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("catalog unavailable");
  });

  it("requires confirmation before switching type and clears incompatible values", async () => {
    const state = createInitialWorkerDraftState({
      ...completeDraft(),
      model_resource_id: 42,
      runtime_image_id: 11,
      type_config_values: { approval_mode: "never" },
      secret_refs: [{ field: "SIGNING_KEY", kind: "env-bundle", id: 8 }],
      skill_ids: [13],
      env_bundle_ids: [8],
      config_document_bindings: [],
    });

    const next = workerCreateDraftReducer(state, {
      type: "change_worker_type",
      workerTypeSlug: "claude-code",
      schemaVersion: 2,
    });

    expect(next.draft.worker_type_slug).toBe("claude-code");
    expect(next.draft.type_schema_version).toBe(2);
    expect(next.draft.model_resource_id).toBe(0);
    expect(next.draft.runtime_image_id).toBe(0);
    expect(next.draft.type_config_values).toEqual({});
    expect(next.draft.secret_refs).toEqual([]);
    expect(next.draft.skill_ids).toEqual([]);
    expect(next.draft.env_bundle_ids).toEqual([]);
    expect(next.draft.tool_model_resource_ids).toEqual({});
    expect(next.draft.initial_task).toBe(state.draft.initial_task);

    const onWorkerTypeChange = vi.fn();
    render(
      <WorkerRuntimeStep
        draft={state.draft}
        modelResources={{ status: "ready", data: [modelResource()] }}
        toolModelResources={{ status: "ready", data: [] }}
        options={{ status: "ready", data: createOptions() }}
        onPatch={vi.fn()}
        onWorkerTypeChange={onWorkerTypeChange}
        t={t}
      />,
    );
    fireEvent.click(screen.getByLabelText("workerCreate.runtime.workerType"));
    fireEvent.click(screen.getByText("workerCreate.runtime.options.claude"));
    expect(screen.getByRole("alertdialog")).toBeInTheDocument();
    fireEvent.click(screen.getByText("workerCreate.typeChange.confirm"));
    expect(onWorkerTypeChange).toHaveBeenCalledWith("claude-code", 2);
  });

  it("keeps lifecycle values identical in the submitted draft", () => {
    const state = createInitialWorkerDraftState(completeDraft());
    const idle = workerCreateDraftReducer(state, {
      type: "set_lifecycle",
      terminationPolicy: "idle",
      idleTimeoutMinutes: 30,
    });
    expect(idle.draft.termination_policy).toBe("idle");
    expect(idle.draft.idle_timeout_minutes).toBe(30);

    const manual = workerCreateDraftReducer(idle, {
      type: "set_lifecycle",
      terminationPolicy: "manual",
      idleTimeoutMinutes: 30,
    });
    expect(manual.draft.termination_policy).toBe("manual");
    expect(manual.draft.idle_timeout_minutes).toBe(0);
  });

  it("accepts a Worker without a primary model when every tool role is bound", () => {
    const options = createOptions();
    options.worker_types[0] = {
      ...options.worker_types[0],
      requires_model_resource: false,
      model_protocol_adapters: [],
      tool_model_requirements: [{
        role: "video-generator",
        provider_keys: ["volcengine"],
        protocol_adapters: ["openai-compatible"],
        modality: "video",
        capability: "video-generation",
      }],
    };
    const draft = {
      ...completeDraft(),
      model_resource_id: 0,
      tool_model_resource_ids: { "video-generator": 84 },
    };

    expect(workerCreateValidity(
      draft,
      { status: "ready", data: options },
      true,
    ).runtime).toBe(true);
  });

});

function step(
  id: 1 | 2 | 3 | 4,
  label: string,
  complete: boolean,
  accessible: boolean,
) {
  return { id, label, complete, accessible };
}

function completeDraft(): WorkerSpecDraft {
  return {
    model_resource_id: 42,
    tool_model_resource_ids: {},
    worker_type_slug: "codex-cli",
    runtime_image_id: 11,
    placement_policy: "automatic",
    compute_target_id: 21,
    deployment_mode: "pooled",
    resource_profile_id: 31,
    type_schema_version: 1,
    type_config_values: {},
    secret_refs: [],
    interaction_mode: "acp",
    automation_level: "autonomous",
    repository_id: 51,
    branch: "main",
    skill_ids: [],
    knowledge_mounts: [],
    env_bundle_ids: [],
    config_document_bindings: [],
    instructions: "Review before editing.",
    initial_task: "Fix the failing test.",
    termination_policy: "manual",
    idle_timeout_minutes: 0,
    alias: "",
    options_revision: "runtime-catalog-1",
  };
}

function createOptions(): WorkerCreateOptions {
  return {
    revision: "runtime-catalog-1",
    worker_types: [
      {
        slug: "codex-cli",
        name: "Codex CLI",
        description: "",
        schema_version: 1,
        config_schema: { version: 1, fields: {} },
        supported_interaction_modes: ["acp"],
        requires_model_resource: true,
        model_protocol_adapters: [],
        tool_model_requirements: [],
        credential_requirements: [],
        config_document_requirements: [],
        selectable: true,
        blocking_reason: "",
        requires_model_resource: true,
        model_protocol_adapters: ["openai-compatible"],
        tool_model_requirements: [],
      },
      {
        slug: "claude-code",
        name: "Claude Code",
        description: "",
        schema_version: 2,
        config_schema: { version: 2, fields: {} },
        supported_interaction_modes: ["acp"],
        requires_model_resource: true,
        model_protocol_adapters: [],
        tool_model_requirements: [],
        credential_requirements: [],
        config_document_requirements: [],
        selectable: true,
        blocking_reason: "",
        requires_model_resource: true,
        model_protocol_adapters: ["anthropic"],
        tool_model_requirements: [],
      },
    ],
    runtime_images: [
      {
        id: 11,
        slug: "codex-stable",
        name: "Codex stable",
        reference: "registry/codex@sha256:test",
        digest: "sha256:test",
        worker_type_slugs: ["codex-cli"],
        selectable: true,
        blocking_reason: "",
      },
    ],
    compute_targets: [
      {
        id: 21,
        slug: "runner-pool",
        name: "Runner pool",
        kind: "runner-pool",
        supports_pooled: true,
        supports_dedicated: false,
        selectable: true,
        blocking_reason: "",
      },
      {
        id: 22,
        slug: "managed-kubernetes",
        name: "Managed Kubernetes",
        kind: "kubernetes",
        supports_pooled: false,
        supports_dedicated: true,
        selectable: false,
        blocking_reason: "Dedicated provisioning is disabled",
      },
    ],
    deployment_modes: [
      {
        value: "pooled",
        name: "Pooled",
        selectable: true,
        blocking_reason: "",
      },
      {
        value: "dedicated",
        name: "Dedicated",
        selectable: false,
        blocking_reason: "Selected target does not support dedicated mode",
      },
    ],
    resource_profiles: [
      {
        id: 31,
        slug: "standard",
        name: "Standard",
        cpu_request_millicpu: 200,
        cpu_limit_millicpu: 1000,
        memory_request_bytes: 268435456,
        memory_limit_bytes: 1073741824,
        selectable: true,
        blocking_reason: "",
      },
    ],
  };
}

function modelResource(): EffectiveResource {
  return {
    selectable: true,
    blockingReason: "",
    connection: {
      id: 1,
      ownerScope: "organization",
      identifier: "openai",
      providerKey: "openai",
      name: "OpenAI",
      baseUrl: "https://api.openai.com/v1",
      configuredFields: ["api_key"],
      status: "valid",
      isEnabled: true,
      validationError: "",
      canManage: true,
      resources: [],
    },
    resource: {
      id: 42,
      providerConnectionId: 1,
      identifier: "gpt-5",
      modelId: "gpt-5",
      displayName: "GPT-5",
      modalities: ["chat"],
      capabilities: ["text-generation"],
      defaultModalities: ["chat"],
      status: "valid",
      isEnabled: true,
      validationError: "",
    },
  };
}

function modelProvider(): ProviderDefinition {
  return {
    key: "openai",
    displayName: "OpenAI",
    modalities: ["chat"],
    credentialFields: [],
    defaultBaseUrl: "https://api.openai.com/v1",
    protocolAdapter: "openai-compatible",
    supportsCustomEndpoint: true,
    supportsModelDiscovery: true,
  };
}
