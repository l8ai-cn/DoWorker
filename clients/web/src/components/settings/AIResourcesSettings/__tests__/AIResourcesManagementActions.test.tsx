import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({
  getCatalog: vi.fn(), listPersonalConnections: vi.fn(), listPersonalEffectiveResources: vi.fn(),
  createPersonalConnection: vi.fn(), createResource: vi.fn(), deleteConnection: vi.fn(), deleteResource: vi.fn(),
  setConnectionEnabled: vi.fn(), setDefaultResource: vi.fn(), setResourceEnabled: vi.fn(), validateConnection: vi.fn(),
  updateConnection: vi.fn(), rotateConnectionCredentials: vi.fn(), updateResource: vi.fn(),
}));

vi.mock("@/lib/api", () => api);
vi.mock("next-intl", () => ({ useTranslations: () => (key: string) => key }));

import { AIResourcesSettings } from "../AIResourcesSettings";

const connection = {
  id: 7, ownerScope: "personal", identifier: "openai-main", providerKey: "openai", name: "OpenAI",
  baseUrl: "https://api.openai.com/v1", configuredFields: ["api_key"], status: "valid", isEnabled: true,
  validationError: "", canManage: true,
  resources: [{
    id: 9, providerConnectionId: 7, identifier: "gpt-4-1", modelId: "gpt-4.1", displayName: "GPT-4.1",
    modalities: ["chat"], capabilities: ["text-generation"], defaultModalities: [], status: "valid",
    isEnabled: true, validationError: "",
  }],
};

describe("AI resource management actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.getCatalog.mockResolvedValue([{
      key: "openai", displayName: "OpenAI", modalities: ["chat"],
      credentialFields: [{ key: "api_key", label: "API Key", secret: true, required: true }],
      defaultBaseUrl: "https://api.openai.com/v1", protocolAdapter: "openai",
      supportsCustomEndpoint: true, supportsModelDiscovery: false,
    }]);
    api.listPersonalConnections.mockResolvedValue([connection]);
    api.listPersonalEffectiveResources.mockResolvedValue([]);
  });

  it("updates connection details, rotates credentials, and updates a model resource", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.edit: OpenAI" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.name"), { target: { value: "OpenAI team" } });
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.save" }));
    await waitFor(() => expect(api.updateConnection).toHaveBeenCalledWith(7, {
      name: "OpenAI team", baseUrl: "https://api.openai.com/v1",
    }));

    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.rotate: OpenAI" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.credentials.apiKey"), { target: { value: "sk-new" } });
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.rotate" }));
    await waitFor(() => expect(api.rotateConnectionCredentials).toHaveBeenCalledWith(7, { api_key: "sk-new" }));

    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.edit: GPT-4.1" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.name"), { target: { value: "GPT-4.1 team" } });
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.save" }));
    await waitFor(() => expect(api.updateResource).toHaveBeenCalledWith(9, {
      displayName: "GPT-4.1 team", modelId: "gpt-4.1", modalities: ["chat"], capabilities: ["text-generation"],
    }));
  });
});
