import { describe, expect, it, vi } from "vitest";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { assertWorkerTemplatePlanReady } from "./worker-template-plan-readiness";

const mocks = vi.hoisted(() => ({
  options: undefined as WorkerCreateOptions | undefined,
}));

vi.mock("@/components/pod/hooks/useWorkerCreateOptions", () => ({
  loadWorkerCreateOptions: async () => mocks.options,
}));

describe("assertWorkerTemplatePlanReady", () => {
  it("rejects an unbound required configuration document before planning", async () => {
    mocks.options = workerOptions();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "do-agent";
    draft.spec.optionsRevision = "catalog-current";
    draft.spec.runtime.runtimeImageId = 1;
    draft.spec.workspace.configDocumentBindings = [{
      documentId: "settings",
      configBundleRef: { kind: "EnvironmentBundle", name: "" },
    }];

    await expect(assertWorkerTemplatePlanReady("acme", draft)).rejects.toThrow(
      'Configuration document "settings" requires an EnvironmentBundle reference.',
    );
  });
});

function workerOptions(): WorkerCreateOptions {
  return {
    revision: "catalog-current",
    worker_types: [{
      slug: "do-agent",
      name: "Do Agent",
      description: "",
      schema_version: 1,
      config_schema: {},
      supported_interaction_modes: ["acp"],
      requires_model_resource: false,
      model_protocol_adapters: [],
      tool_model_requirements: [],
      credential_requirements: [],
      config_document_requirements: [{
        document_id: "settings",
        format: "json",
        target_path: "DO_AGENT_SETTINGS",
        required: true,
      }],
      selectable: true,
      blocking_reason: "",
    }],
    runtime_images: [{
      id: 1,
      slug: "do-agent",
      name: "Do Agent",
      reference: "registry.example/do-agent:latest",
      digest: "sha256:test",
      worker_type_slugs: ["do-agent"],
      selectable: true,
      blocking_reason: "",
    }],
    compute_targets: [],
    deployment_modes: [],
    resource_profiles: [],
  };
}
