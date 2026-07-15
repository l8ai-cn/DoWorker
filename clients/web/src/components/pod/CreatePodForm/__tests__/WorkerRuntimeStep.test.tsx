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
        modelProviders={{ status: "ready", data: [modelProvider()] }}
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

  it("renders tool model roles and filters resources by the server contract", () => {
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
    const compatible = resourceForTool(84, "volcengine", "video", "video-generation");
    const wrongProvider = resourceForTool(85, "openai", "video", "video-generation");

    render(
      <WorkerRuntimeStep
        draft={{
          ...completeDraft(),
          model_resource_id: 0,
          tool_model_resource_ids: { "video-generator": 84 },
        }}
        modelResources={{ status: "ready", data: [compatible, wrongProvider] }}
        modelProviders={{ status: "ready", data: providers() }}
        options={{ status: "ready", data: options }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={(key) => key}
      />,
    );

    expect(screen.queryByLabelText("workerCreate.runtime.model")).not.toBeInTheDocument();
    expect(
      screen.getByLabelText("workerCreate.runtime.toolModel · video-generator"),
    ).toHaveTextContent("Volcengine · Model 84");
    expect(screen.queryByText("OpenAI · Model 85")).not.toBeInTheDocument();
  });
});

function providers(): ProviderDefinition[] {
  return [
    provider("volcengine", "openai-compatible"),
    provider("openai", "openai-compatible"),
  ];
}

function provider(key: string, protocolAdapter: string): ProviderDefinition {
  return {
    key,
    displayName: key,
    modalities: ["chat", "video"],
    credentialFields: [],
    defaultBaseUrl: "https://provider.example",
    protocolAdapter,
    supportsCustomEndpoint: false,
    supportsModelDiscovery: false,
  };
}

function resourceForTool(
  id: number,
  providerKey: string,
  modality: string,
  capability: string,
): EffectiveResource {
  const resource = modelResource();
  return {
    ...resource,
    connection: {
      ...resource.connection!,
      id,
      name: providerKey === "volcengine" ? "Volcengine" : "OpenAI",
      providerKey,
    },
    resource: {
      ...resource.resource!,
      id,
      providerConnectionId: id,
      displayName: `Model ${id}`,
      modalities: [modality],
      capabilities: [capability],
    },
  };
}
