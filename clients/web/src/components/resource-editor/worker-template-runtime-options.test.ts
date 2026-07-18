import { describe, expect, it } from "vitest";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { requiredCredentialReferenceFields } from "./worker-template-definition-bindings";
import {
  selectWorkerTemplateType,
  synchronizeWorkerTemplateRuntime,
} from "./worker-template-runtime-options";

describe("worker template runtime options", () => {
  it("preserves matching documents and replaces bindings for a new Worker type", () => {
    const draft = configuredDraft();
    const next = selectWorkerTemplateType(draft, workerOptions(), "minimax-cli");

    expect(next?.spec.workspace.configDocumentBindings).toEqual([
      binding("shared", "shared-bundle", 3),
    ]);
  });

  it("repairs stale binding order from the authoritative Worker catalog", () => {
    const draft = configuredDraft();
    draft.spec.workspace.configDocumentBindings = [
      binding("obsolete", "obsolete-bundle"),
      binding("codex", "codex-bundle"),
      binding("shared", "shared-bundle", 3),
    ];

    const next = synchronizeWorkerTemplateRuntime(draft, workerOptions());

    expect(next?.spec.workspace.configDocumentBindings).toEqual([
      binding("shared", "shared-bundle", 3),
      binding("codex", "codex-bundle"),
    ]);
  });

  it("removes credentials that are not declared by the selected Worker type", () => {
    const draft = configuredDraft();
    draft.spec.typeConfig.secretRefs = {
      CURSOR_API_KEY: {
        kind: "EnvironmentBundle",
        name: "cursor-credentials",
      },
      STALE_KEY: {
        kind: "EnvironmentBundle",
        name: "stale-credentials",
      },
    };

    const next = selectWorkerTemplateType(
      draft,
      workerOptions(),
      "cursor-cli",
    );

    expect(next?.spec.typeConfig.secretRefs).toEqual({
      CURSOR_API_KEY: {
        kind: "EnvironmentBundle",
        name: "cursor-credentials",
      },
    });
  });

  it("derives credential required state from the projected Worker schema", () => {
    expect([...requiredCredentialReferenceFields({
      fields: {
        CURSOR_API_KEY: { kind: "secret", required: false },
        SIGNING_KEY: { kind: "secret", required: true },
      },
    })]).toEqual(["SIGNING_KEY"]);
  });
});

function configuredDraft() {
  const draft = createWorkerTemplateDraft("acme");
  draft.spec.workerType = "codex-cli";
  draft.spec.optionsRevision = "catalog-old";
  draft.spec.runtime.runtimeImageId = 11;
  draft.spec.workspace.configDocumentBindings = [
    binding("shared", "shared-bundle", 3),
    binding("codex", "codex-bundle"),
    binding("obsolete", "obsolete-bundle"),
  ];
  return draft;
}

function binding(documentId: string, name: string, revision?: number) {
  return {
    documentId,
    configBundleRef: {
      kind: "EnvironmentBundle",
      name,
      revision,
    },
  };
}

function workerOptions(): WorkerCreateOptions {
  return {
    revision: "catalog-current",
    worker_types: [
      workerType("codex-cli", ["shared", "codex"]),
      workerType("minimax-cli", ["shared", "minimax"]),
      workerType("cursor-cli", [], [{
        id: "cursor-api-key",
        source_kind: "credential_bundle",
        source_ref: "cursor",
        target_kind: "env",
        target_name: "CURSOR_API_KEY",
      }]),
    ],
    runtime_images: [
      runtimeImage(11, "codex-cli"),
      runtimeImage(42, "minimax-cli"),
    ],
    compute_targets: [],
    deployment_modes: [],
    resource_profiles: [],
  };
}

function workerType(
  slug: string,
  documentIDs: string[],
  credentialRequirements: WorkerCreateOptions["worker_types"][number]["credential_requirements"] = [],
) {
  return {
    slug,
    name: slug,
    description: "",
    schema_version: 1,
    config_schema: {},
    supported_interaction_modes: ["pty"],
    requires_model_resource: false,
    model_protocol_adapters: [],
    tool_model_requirements: [],
    credential_requirements: credentialRequirements,
    config_document_requirements: documentIDs.map((documentID) => ({
      document_id: documentID,
      format: "json",
      target_path: `/workspace/${documentID}.json`,
      required: true,
    })),
    selectable: true,
    blocking_reason: "",
  };
}

function runtimeImage(id: number, workerTypeSlug: string) {
  return {
    id,
    slug: workerTypeSlug,
    name: workerTypeSlug,
    reference: "",
    digest: "",
    worker_type_slugs: [workerTypeSlug],
    selectable: true,
    blocking_reason: "",
  };
}
