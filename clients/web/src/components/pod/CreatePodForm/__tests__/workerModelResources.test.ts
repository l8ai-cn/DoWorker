import { describe, expect, it } from "vitest";
import type { EffectiveResource, ProviderDefinition } from "@/lib/api/facade/aiResource";
import {
  agentRequiresModelResource,
  compatibleModelResources,
} from "../workerModelResources";

const geminiProvider: ProviderDefinition = {
  key: "gemini",
  displayName: "Gemini",
  modalities: ["chat"],
  credentialFields: [],
  defaultBaseUrl: "https://generativelanguage.googleapis.com",
  protocolAdapter: "gemini",
  supportsCustomEndpoint: false,
  supportsModelDiscovery: false,
};

const geminiResource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 1,
    ownerScope: "user",
    identifier: "gemini-main",
    providerKey: "gemini",
    name: "Gemini",
    baseUrl: "https://generativelanguage.googleapis.com",
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
    identifier: "gemini-pro",
    modelId: "gemini-pro",
    displayName: "Gemini Pro",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: ["chat"],
    status: "valid",
    isEnabled: true,
    validationError: "",
  },
};

describe("workerModelResources", () => {
  it("allows selectable Gemini resources when exact model injection is supported", () => {
    expect(agentRequiresModelResource("gemini-cli")).toBe(true);
    expect(compatibleModelResources("gemini-cli", [geminiResource], [geminiProvider])).toEqual([
      geminiResource,
    ]);
  });
});
