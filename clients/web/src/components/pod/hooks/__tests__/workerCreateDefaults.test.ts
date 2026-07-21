import { describe, expect, it } from "vitest";
import {
  defaultConfigDocumentPatch,
  defaultToolModelPatch,
} from "../workerCreateDefaults";
import { completeDraft, createOptions } from "../../CreatePodForm/__tests__/test-utils";

describe("workerCreateDefaults", () => {
  it("selects a compatible tool model for required tool roles", () => {
    const draft = completeDraft();
    const options = createOptions();
    options.worker_types[0].tool_model_requirements = [{
      role: "seedance-video",
      provider_keys: ["sub2api-seedance"],
      protocol_adapters: ["ark-seedance"],
      modality: "video",
      capability: "video-generation",
    }];

    const patch = defaultToolModelPatch(
      draft,
      options.worker_types[0],
      [seedanceResource()],
    );

    expect(patch.tool_model_resource_ids).toEqual({ "seedance-video": 77 });
  });

  it("binds required config documents without forcing optional documents", () => {
    const draft = completeDraft();
    const options = createOptions();
    options.worker_types[0].slug = "seedance-expert";
    options.worker_types[0].config_document_requirements = [
      {
        document_id: "settings",
        format: "json",
        target_path: "/workspace/settings.json",
        required: true,
      },
      {
        document_id: "optional",
        format: "json",
        target_path: "/workspace/optional.json",
        required: false,
      },
    ];

    const patch = defaultConfigDocumentPatch(
      draft,
      options.worker_types[0],
      [{
        id: 88,
        name: "seedance-settings",
        agent_slug: "seedance-expert",
        kind: "config",
        kind_primary: false,
      }],
    );

    expect(patch.config_document_bindings).toEqual([{
      document_id: "settings",
      config_bundle_id: 88,
    }]);
  });
});

function seedanceResource() {
  return {
    selectable: true,
    blockingReason: "",
    connection: {
      id: 2,
      ownerScope: "organization" as const,
      identifier: "seedance",
      providerKey: "sub2api-seedance",
      name: "Seedance",
      baseUrl: "https://example.com/v1",
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
      identifier: "seedance-video",
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
