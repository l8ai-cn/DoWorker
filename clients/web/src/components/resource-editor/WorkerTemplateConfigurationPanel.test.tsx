import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AsyncState } from "@/components/pod/hooks/workerCreateDraft";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { render, screen, waitFor } from "@/test/test-utils";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateConfigurationPanel } from "./WorkerTemplateConfigurationPanel";

const mocks = vi.hoisted(() => ({
  workerOptions: { status: "loading" } as AsyncState<WorkerCreateOptions>,
}));

vi.mock("@/components/pod/hooks/useWorkerCreateOptions", () => ({
  useWorkerCreateOptions: () => mocks.workerOptions,
}));

vi.mock("./use-resource-reference-options", () => ({
  useResourceReferenceOptions: () => ({
    loading: false,
    error: null,
    errorsByKind: {},
    byKind: {},
  }),
}));

vi.mock("./WorkerTemplateBindingsPanel", () => ({
  WorkerTemplateBindingsPanel: () => null,
}));

vi.mock("./WorkerTemplateIdentityPanel", () => ({
  WorkerTemplateIdentityPanel: () => null,
}));

vi.mock("./WorkerTemplateLifecyclePanel", () => ({
  WorkerTemplateLifecyclePanel: () => null,
}));

vi.mock("./WorkerTemplateRuntimePanel", () => ({
  WorkerTemplateRuntimePanel: () => null,
}));

describe("WorkerTemplateConfigurationPanel", () => {
  beforeEach(() => {
    mocks.workerOptions = { status: "loading" };
  });

  it.each([
    [
      "not ready",
      { status: "loading" } as const,
      "Worker options are loading.",
    ],
    [
      "load failed",
      { status: "error", error: "catalog unavailable" } as const,
      "catalog unavailable",
    ],
    [
      "worker type unresolved",
      unresolvedWorkerOptions(),
      "The selected Worker type is unavailable",
    ],
  ])("keeps references read-only when options are %s", async (
    _,
    workerOptions,
    expectedBlockReason,
  ) => {
    mocks.workerOptions = workerOptions;
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "missing-worker";
    draft.spec.typeConfig.secretRefs.API_KEY = {
      kind: "EnvironmentBundle",
      name: "existing-credentials",
      revision: 4,
    };
    draft.spec.workspace.configDocumentBindings = [{
      documentId: "settings",
      configBundleRef: {
        kind: "EnvironmentBundle",
        name: "existing-settings",
        revision: 3,
      },
    }];

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    expect(screen.getByRole("textbox", { name: "API_KEY" }))
      .toHaveValue("existing-credentials");
    expect(screen.getByRole("textbox", { name: "API_KEY" }))
      .toHaveAttribute("readonly");
    expect(screen.getByRole("textbox", { name: "settings" }))
      .toHaveValue("existing-settings");
    expect(screen.getByRole("textbox", { name: "settings" }))
      .toHaveAttribute("readonly");
    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(
        expect.stringContaining(expectedBlockReason),
      );
    });
  });

  it("clears the Plan block when the selected worker type is available", async () => {
    mocks.workerOptions = resolvedWorkerOptions();
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "codex-cli";

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(null);
    });
  });

  it("blocks the Plan when a required credential reference is missing", async () => {
    mocks.workerOptions = resolvedWorkerOptions(
      "cursor-cli",
      { fields: { CURSOR_API_KEY: { kind: "secret", required: true } } },
    );
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "cursor-cli";

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(
        expect.stringContaining("CURSOR_API_KEY"),
      );
    });
  });

  it("blocks the Plan when no credential group member is selected", async () => {
    mocks.workerOptions = resolvedWorkerOptions(
      "aider",
      {
        credential_requirement_groups: [{
          id: "provider-api-key",
          any_of: ["OPENAI_API_KEY", "ANTHROPIC_API_KEY"],
        }],
      },
    );
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "aider";

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(
        expect.stringContaining("OPENAI_API_KEY"),
      );
    });
  });

  it("blocks the Plan when a required configuration document is unbound", async () => {
    mocks.workerOptions = resolvedWorkerOptions(
      "do-agent",
      {},
      [{
        document_id: "settings",
        format: "json",
        target_path: "DO_AGENT_SETTINGS",
        required: true,
      }],
    );
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "do-agent";

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(
        expect.stringContaining("settings"),
      );
    });
  });

  it("does not block the Plan when an optional configuration document is unbound", async () => {
    mocks.workerOptions = resolvedWorkerOptions(
      "openclaw",
      {},
      [{
        document_id: "openclaw-json",
        format: "json",
        target_path: "OPENCLAW_CONFIG",
        required: false,
      }],
    );
    const onPlanBlockChange = vi.fn();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "openclaw";

    render(
      <WorkerTemplateConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={vi.fn()}
        onPlanBlockChange={onPlanBlockChange}
      />,
    );

    await waitFor(() => {
      expect(onPlanBlockChange).toHaveBeenLastCalledWith(null);
    });
  });
});

function unresolvedWorkerOptions(): AsyncState<WorkerCreateOptions> {
  return {
    status: "ready",
    data: {
      revision: "catalog-current",
      worker_types: [],
      runtime_images: [],
      compute_targets: [],
      deployment_modes: [],
      resource_profiles: [],
    },
  };
}

function resolvedWorkerOptions(
  slug = "codex-cli",
  configSchema: Record<string, unknown> = {},
  configDocumentRequirements: WorkerCreateOptions["worker_types"][number]["config_document_requirements"] = [],
): AsyncState<WorkerCreateOptions> {
  return {
    status: "ready",
    data: {
      revision: "catalog-current",
      worker_types: [{
        slug,
        name: slug,
        description: "",
        schema_version: 1,
        config_schema: configSchema,
        supported_interaction_modes: ["acp"],
        requires_model_resource: false,
        model_protocol_adapters: [],
        tool_model_requirements: [],
        credential_requirements: [],
        config_document_requirements: configDocumentRequirements,
        selectable: true,
        blocking_reason: "",
      }],
      runtime_images: [],
      compute_targets: [],
      deployment_modes: [],
      resource_profiles: [],
    },
  };
}
