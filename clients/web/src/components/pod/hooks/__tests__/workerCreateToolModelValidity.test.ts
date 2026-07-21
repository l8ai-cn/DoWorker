import { describe, expect, it } from "vitest";
import { workerCreateValidity } from "../workerCreateValidity";
import { completeDraft, createOptions } from "../../CreatePodForm/__tests__/test-utils";

describe("workerCreateValidity tool models", () => {
  it("requires every declared tool model role", () => {
    const draft = completeDraft();
    const options = createOptions();
    options.worker_types[0].tool_model_requirements = [{
      role: "seedance-video",
      provider_keys: ["doubao", "sub2api-seedance"],
      protocol_adapters: ["openai-compatible", "ark-seedance"],
      modality: "video",
      capability: "video-generation",
    }];

    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).runtime).toBe(false);

    draft.tool_model_resource_ids = { "seedance-video": 77 };
    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).runtime).toBe(true);
  });

  it("only requires config documents marked as required", () => {
    const draft = completeDraft();
    const options = createOptions();
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

    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).workspace).toBe(false);

    draft.config_document_bindings = [{
      document_id: "settings",
      config_bundle_id: 88,
    }];
    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).workspace).toBe(true);
  });
});
