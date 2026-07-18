import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  EffectiveResource,
  ProviderDefinition,
} from "@/lib/api/facade/aiResource";
import { WorkerRuntimeStep } from "../WorkerRuntimeStep";
import {
  completeDraft,
  createOptions,
  modelProvider,
  modelResource,
} from "./test-utils";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("WorkerRuntimeStep", () => {
  it("explains why creation is blocked when no compatible model resource exists", () => {
    render(
      <WorkerRuntimeStep
        draft={completeDraft()}
        modelResources={{ status: "ready", data: [] }}
        toolModelResources={{ status: "ready", data: [] }}
        options={{ status: "ready", data: createOptions() }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={(key) => key}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "ide.createPod.noModelResourcesAvailableHint",
    );
  });

  it("does not render a model selector for cursor", () => {
    const draft = completeDraft();
    draft.worker_type_slug = "cursor-cli";
    draft.model_resource_id = 0;
    const options = createOptions();
    options.worker_types[0] = {
      ...options.worker_types[0],
      slug: "cursor-cli",
      requires_model_resource: false,
    };

    render(
      <WorkerRuntimeStep
        draft={draft}
        modelResources={{ status: "ready", data: [] }}
        toolModelResources={{ status: "ready", data: [] }}
        options={{ status: "ready", data: options }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={(key) => key}
      />,
    );

    expect(screen.queryByTestId("worker-runtime-field-model")).not.toBeInTheDocument();
  });

  it("renders the required Seedance video resource picker", () => {
    const draft = completeDraft();
    draft.worker_type_slug = "seedance-expert";
    draft.tool_model_resource_ids = {};
    const options = createOptions();
    options.worker_types[0] = {
      ...options.worker_types[0],
      slug: "seedance-expert",
      tool_model_requirements: [{
        role: "seedance-video",
        provider_keys: ["doubao"],
        protocol_adapters: ["openai-compatible"],
        modality: "video",
        capability: "video-generation",
      }],
    };

    render(
      <WorkerRuntimeStep
        draft={draft}
        modelResources={{ status: "ready", data: [] }}
        toolModelResources={{ status: "ready", data: [seedanceResource()] }}
        options={{ status: "ready", data: options }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={(key) => key}
      />,
    );

    expect(screen.getByTestId("worker-runtime-field-seedance-video")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Seedance video" })).toBeInTheDocument();
  });
});

function seedanceResource() {
  return {
    selectable: true,
    blockingReason: "",
    connection: {
      id: 2,
      ownerScope: "organization" as const,
      identifier: "doubao",
      providerKey: "doubao",
      name: "Doubao",
      baseUrl: "https://ark.cn-beijing.volces.com/api/v3",
      configuredFields: ["api_key"],
      status: "valid" as const,
      isEnabled: true,
      validationError: "",
      canManage: true,
      resources: [],
    },
    resource: {
      id: 77,
      providerConnectionId: 2,
      identifier: "seedance-2",
      modelId: "doubao-seedance-2-0-260128",
      displayName: "Seedance 2.0",
      modalities: ["video"],
      capabilities: ["video-generation"],
      defaultModalities: ["video"],
      status: "valid" as const,
      isEnabled: true,
      validationError: "",
    },
  };
}
