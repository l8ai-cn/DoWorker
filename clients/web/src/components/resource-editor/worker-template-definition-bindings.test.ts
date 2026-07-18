import { describe, expect, it } from "vitest";
import {
  missingRequiredConfigDocumentReference,
  synchronizeConfigDocumentBindings,
  updateConfigDocumentBinding,
} from "./worker-template-definition-bindings";

const requirements = [
  {
    document_id: "settings",
    format: "json",
    target_path: "DO_AGENT_SETTINGS",
    required: true,
  },
  {
    document_id: "openclaw-json",
    format: "json",
    target_path: "openclaw-home/.openclaw/openclaw.json",
    required: false,
  },
];

describe("Worker configuration document bindings", () => {
  it("keeps only required unbound document placeholders", () => {
    expect(synchronizeConfigDocumentBindings(requirements, [])).toEqual([{
      documentId: "settings",
      configBundleRef: { kind: "EnvironmentBundle", name: "" },
    }]);
  });

  it("removes a blank optional document binding", () => {
    expect(updateConfigDocumentBinding(
      requirements,
      [{
        documentId: "settings",
        configBundleRef: { kind: "EnvironmentBundle", name: "" },
      }],
      "openclaw-json",
      undefined,
    )).toEqual([{
      documentId: "settings",
      configBundleRef: { kind: "EnvironmentBundle", name: "" },
    }]);
  });

  it("does not block planning for an unbound optional document", () => {
    expect(missingRequiredConfigDocumentReference(
      requirements,
      [{
        documentId: "openclaw-json",
        configBundleRef: { kind: "EnvironmentBundle", name: "" },
      }],
    )).toBe("settings");
    expect(missingRequiredConfigDocumentReference(
      requirements.filter((requirement) => !requirement.required),
      [],
    )).toBeNull();
  });
});
