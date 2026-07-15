import { describe, expect, it } from "vitest";
import { workerCreateValidity } from "../workerCreateValidity";
import { completeDraft, createOptions } from "../../CreatePodForm/__tests__/test-utils";

describe("workerCreateValidity tool models", () => {
  it("requires every declared tool model role", () => {
    const draft = completeDraft();
    const options = createOptions();
    options.worker_types[0].tool_model_requirements = [{
      role: "seedance-video",
      provider_keys: ["doubao"],
      protocol_adapters: ["openai-compatible"],
      modality: "video",
      capability: "video-generation",
    }];

    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).runtime).toBe(false);

    draft.tool_model_resource_ids = { "seedance-video": 77 };
    expect(workerCreateValidity(draft, { status: "ready", data: options }, true).runtime).toBe(true);
  });
});
