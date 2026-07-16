import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({
  getCatalog: vi.fn(),
  listPersonalConnections: vi.fn(),
  listOrganizationConnections: vi.fn(),
  listPersonalEffectiveResources: vi.fn(),
  listOrganizationEffectiveResources: vi.fn(),
  createPersonalConnection: vi.fn(),
  createOrganizationConnection: vi.fn(),
  createResource: vi.fn(),
  deleteConnection: vi.fn(),
  deleteResource: vi.fn(),
  setConnectionEnabled: vi.fn(),
  setDefaultResource: vi.fn(),
  setResourceEnabled: vi.fn(),
  rotateConnectionCredentials: vi.fn(),
  updateConnection: vi.fn(),
  updateResource: vi.fn(),
  validateConnection: vi.fn(),
}));

vi.mock("@/lib/api", () => api);
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

import { AIResourcesSettings } from "../AIResourcesSettings";

const catalog = [{
  key: "openai",
  displayName: "OpenAI",
  modalities: ["chat", "image"],
  credentialFields: [{ key: "api_key", label: "API Key", secret: true, required: true }],
  defaultBaseUrl: "https://api.openai.com",
  protocolAdapter: "openai",
  supportsCustomEndpoint: true,
  supportsModelDiscovery: false,
}];

const connections = [{
  id: 7,
  ownerScope: "organization",
  identifier: "openai-production",
  providerKey: "openai",
  name: "OpenAI production",
  baseUrl: "https://api.openai.com",
  configuredFields: ["api_key"],
  status: "valid",
  isEnabled: true,
  validationError: "",
  canManage: true,
  resources: [
    {
      id: 10,
      providerConnectionId: 7,
      identifier: "gpt-4-1",
      modelId: "gpt-4.1",
      displayName: "GPT-4.1",
      modalities: ["chat"],
      capabilities: ["tools"],
      defaultModalities: ["chat"],
      status: "valid",
      isEnabled: true,
      validationError: "",
    },
    {
      id: 11,
      providerConnectionId: 7,
      identifier: "gpt-image",
      modelId: "gpt-image-1",
      displayName: "GPT Image",
      modalities: ["image"],
      capabilities: [],
      defaultModalities: [],
      status: "invalid",
      isEnabled: false,
      validationError: "credentials rejected",
    },
  ],
}];

describe("AIResourcesSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.getCatalog.mockResolvedValue(catalog);
    api.listPersonalConnections.mockResolvedValue(connections);
    api.listOrganizationConnections.mockResolvedValue(connections);
    api.listPersonalEffectiveResources.mockResolvedValue([]);
    api.listOrganizationEffectiveResources.mockResolvedValue([]);
  });

  it("loads organization resources and lets an owner open provider onboarding", async () => {
    render(<AIResourcesSettings scope="organization" organizationSlug="acme" canManage />);

    await waitFor(() => expect(api.listOrganizationConnections).toHaveBeenCalledWith("acme"));
    expect(screen.getByRole("button", { name: "settings.aiResources.addConnection" })).toBeVisible();

    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addConnection" }));
    expect(screen.getByRole("dialog", { name: "settings.aiResources.connection.createTitle" })).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.provider" }));
    fireEvent.click(screen.getByRole("option", { name: "OpenAI" }));
    expect(screen.getByLabelText("settings.aiResources.connection.credentials.apiKey")).toHaveAttribute("type", "password");
    const create = screen.getByRole("button", { name: "settings.aiResources.connection.create" });
    expect(create).toBeDisabled();
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.name"), { target: { value: "OpenAI" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.identifier"), { target: { value: "openai" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.credentials.apiKey"), { target: { value: "sk-test" } });
    expect(create).toBeEnabled();
  });

  it("keeps model resource creation disabled until the required fields are complete", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    expect(screen.getByText("settings.aiResources.default")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    const create = screen.getByRole("button", { name: "settings.aiResources.resource.create" });
    expect(create).toBeDisabled();
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.name"), { target: { value: "GPT" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.modelId"), { target: { value: "gpt-4.1" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.identifier"), { target: { value: "gpt-4-1" } });
    expect(create).toBeDisabled();
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.capability.textGeneration" }));
    expect(create).toBeEnabled();
    fireEvent.click(create);
    await waitFor(() => expect(api.createResource).toHaveBeenCalledWith(7, {
      displayName: "GPT",
      identifier: "gpt-4-1",
      modelId: "gpt-4.1",
      modalities: ["chat"],
      capabilities: ["text-generation"],
    }));
  });

  it("only presents the selected provider's supported modalities", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    expect(screen.getByRole("checkbox", { name: "settings.aiResources.modality.chat" })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: "settings.aiResources.modality.image" })).toBeInTheDocument();
    expect(screen.queryByRole("checkbox", { name: "settings.aiResources.modality.audio" })).not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox", { name: "settings.aiResources.modality.multimodal" })).not.toBeInTheDocument();
  });

  it("keeps audio capability selection explicit", async () => {
    api.getCatalog.mockResolvedValueOnce([{ ...catalog[0], key: "elevenlabs", modalities: ["audio"] }]);
    api.listPersonalConnections.mockResolvedValueOnce([{ ...connections[0], providerKey: "elevenlabs" }]);
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    expect(screen.getByRole("checkbox", { name: "settings.aiResources.modality.audio" })).toBeChecked();
    expect(screen.queryByRole("checkbox", { name: "settings.aiResources.modality.chat" })).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.name"), { target: { value: "Voice" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.modelId"), { target: { value: "eleven-multilingual-v2" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.identifier"), { target: { value: "eleven-voice" } });
    const create = screen.getByRole("button", { name: "settings.aiResources.resource.create" });
    expect(create).toBeDisabled();
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.capability.textToSpeech" }));
    expect(create).toBeEnabled();
    fireEvent.click(create);
    await waitFor(() => expect(api.createResource).toHaveBeenCalledWith(7, {
      displayName: "Voice",
      identifier: "eleven-voice",
      modelId: "eleven-multilingual-v2",
      modalities: ["audio"],
      capabilities: ["text-to-speech"],
    }));
  });

  it("does not submit a capability that belonged only to a removed modality", async () => {
    api.getCatalog.mockResolvedValueOnce([{ ...catalog[0], modalities: ["chat", "audio"] }]);
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.modality.audio" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.capability.textGeneration" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.capability.textToSpeech" }));
    fireEvent.click(screen.getByRole("checkbox", { name: "settings.aiResources.modality.audio" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.name"), { target: { value: "GPT" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.modelId"), { target: { value: "gpt-4.1" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.identifier"), { target: { value: "gpt-4-1" } });
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.create" }));

    await waitFor(() => expect(api.createResource).toHaveBeenCalledWith(7, expect.objectContaining({
      modalities: ["chat"],
      capabilities: ["text-generation"],
    })));
  });

  it("clears connection credentials when the dialog closes", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addConnection" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.provider" }));
    fireEvent.click(screen.getByRole("option", { name: "OpenAI" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.name"), { target: { value: "OpenAI" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.identifier"), { target: { value: "openai" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.connection.credentials.apiKey"), { target: { value: "sk-test" } });
    fireEvent.click(screen.getByRole("button", { name: "common.cancel" }));

    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addConnection" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.provider" }));
    fireEvent.click(screen.getByRole("option", { name: "OpenAI" }));
    expect(screen.getByLabelText("settings.aiResources.connection.name")).toHaveValue("");
    expect(screen.getByLabelText("settings.aiResources.connection.credentials.apiKey")).toHaveValue("");
  });

  it("clears model resource fields when the dialog closes", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.name"), { target: { value: "GPT" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.modelId"), { target: { value: "gpt-4.1" } });
    fireEvent.change(screen.getByLabelText("settings.aiResources.resource.identifier"), { target: { value: "gpt-4-1" } });
    fireEvent.click(screen.getByRole("button", { name: "common.cancel" }));

    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.addResource" }));
    expect(screen.getByLabelText("settings.aiResources.resource.name")).toHaveValue("");
    expect(screen.getByLabelText("settings.aiResources.resource.modelId")).toHaveValue("");
    expect(screen.getByLabelText("settings.aiResources.resource.identifier")).toHaveValue("");
  });

  it("requires confirmation before deleting a resource", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.delete: GPT-4.1" }));
    expect(screen.getByRole("alertdialog", { name: "settings.aiResources.deleteConfirm.title" })).toBeVisible();
    expect(screen.getByRole("heading", { name: "settings.aiResources.deleteConfirm.title" })).toBeVisible();
    expect(api.deleteResource).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.deleteConfirm.action" }));
    await waitFor(() => expect(api.deleteResource).toHaveBeenCalledWith(10));
  });

  it("requires confirmation before deleting a connection", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("OpenAI production");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.connection.delete: OpenAI production" }));
    expect(screen.getByRole("heading", { name: "settings.aiResources.deleteConfirm.title" })).toBeVisible();
    expect(api.deleteConnection).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.deleteConfirm.action" }));
    await waitFor(() => expect(api.deleteConnection).toHaveBeenCalledWith(7));
  });

  it("closes a successful deletion confirmation when the refresh fails", async () => {
    api.listPersonalConnections.mockResolvedValueOnce(connections).mockRejectedValueOnce(new Error("refresh failed"));
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.delete: GPT-4.1" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.deleteConfirm.action" }));
    await waitFor(() => expect(screen.queryByRole("heading", { name: "settings.aiResources.deleteConfirm.title" })).not.toBeInTheDocument());
    expect(await screen.findByRole("alert")).toHaveTextContent("settings.aiResources.loadError");
  });

  it("clears the action error when retry succeeds after a failed refresh", async () => {
    api.listPersonalConnections.mockResolvedValueOnce(connections).mockRejectedValueOnce(new Error("refresh failed"));
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.delete: GPT-4.1" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.deleteConfirm.action" }));
    const retry = await screen.findByRole("button", { name: "settings.aiResources.retry" });
    fireEvent.click(retry);

    await waitFor(() => expect(screen.queryByText("settings.aiResources.operationError")).not.toBeInTheDocument());
    expect(screen.queryByText("settings.aiResources.loadError")).not.toBeInTheDocument();
  });

  it("reports a failed deletion inside its confirmation dialog", async () => {
    api.deleteResource.mockRejectedValueOnce(new Error("delete rejected"));
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.resource.delete: GPT-4.1" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.deleteConfirm.action" }));

    expect(await screen.findByText("settings.aiResources.deleteConfirm.error")).toBeVisible();
    expect(screen.getByRole("heading", { name: "settings.aiResources.deleteConfirm.title" })).toBeVisible();
  });

  it("renders the localized loading state", () => {
    api.getCatalog.mockReturnValue(new Promise(() => undefined));
    render(<AIResourcesSettings scope="personal" canManage />);

    expect(screen.getByText("settings.aiResources.loading")).toBeVisible();
  });

  it("filters by modality and keeps invalid or disabled resources visible with their state", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    expect(screen.getByText("GPT Image")).toBeVisible();
    expect(screen.getByText("settings.aiResources.status.invalid")).toBeVisible();
    expect(screen.getByText("settings.aiResources.status.disabled")).toBeVisible();
    expect(screen.getByText("settings.aiResources.validation.credentialsRejected")).toBeVisible();

    fireEvent.click(screen.getByRole("tab", { name: "settings.aiResources.modality.chat" }));
    expect(screen.getByText("GPT-4.1")).toBeVisible();
    expect(screen.queryByText("GPT Image")).not.toBeInTheDocument();
  });

  it("sets a multi-modal resource as the default for the active filter modality", async () => {
    api.listPersonalConnections.mockResolvedValueOnce([{
      ...connections[0],
      resources: [{ ...connections[0].resources[0], modalities: ["chat", "image"], defaultModalities: [] }],
    }]);
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("tab", { name: "settings.aiResources.modality.image" }));
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.setDefault" }));

    await waitFor(() => expect(api.setDefaultResource).toHaveBeenCalledWith(10, "image"));
  });

  it("requires an explicit default modality while viewing all resource types", async () => {
    api.listPersonalConnections.mockResolvedValueOnce([{
      ...connections[0],
      resources: [{ ...connections[0].resources[0], modalities: ["chat", "image"], defaultModalities: [] }],
    }]);
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    const setDefault = screen.getByRole("button", { name: "settings.aiResources.setDefault" });
    expect(setDefault).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "settings.aiResources.defaultModality" }));
    fireEvent.click(screen.getByRole("option", { name: "settings.aiResources.modality.image" }));
    fireEvent.click(setDefault);

    await waitFor(() => expect(api.setDefaultResource).toHaveBeenCalledWith(10, "image"));
  });

  it("shows a clear empty state when the selected capability has no resources", async () => {
    render(<AIResourcesSettings scope="personal" canManage />);

    await screen.findByText("GPT-4.1");
    fireEvent.click(screen.getByRole("tab", { name: "settings.aiResources.modality.video" }));

    expect(screen.getByText("settings.aiResources.emptyResources")).toBeVisible();
    expect(screen.queryByText("OpenAI production")).not.toBeInTheDocument();
  });

  it("shows unconnected usage honestly and does not expose management actions to an organization member", async () => {
    render(<AIResourcesSettings scope="organization" organizationSlug="acme" canManage={false} />);

    await screen.findByText("OpenAI production");
    expect(screen.getByText("settings.aiResources.usageNotConnected")).toBeVisible();
    expect(screen.queryByRole("button", { name: "settings.aiResources.addConnection" })).not.toBeInTheDocument();
  });

  it("renders a recoverable load error", async () => {
    api.listPersonalConnections.mockRejectedValueOnce(new Error("offline"));
    render(<AIResourcesSettings scope="personal" canManage />);

    expect(await screen.findByRole("alert")).toHaveTextContent("settings.aiResources.loadError");
    expect(screen.getByRole("button", { name: "settings.aiResources.retry" })).toBeVisible();
  });

  it("keeps the page visible and reports a failed management action", async () => {
    api.validateConnection.mockRejectedValueOnce(new Error("validation rejected"));
    render(<AIResourcesSettings scope="personal" canManage />);

    fireEvent.click(await screen.findByRole("button", { name: "settings.aiResources.validate" }));

    expect(await screen.findByText("settings.aiResources.operationError")).toBeVisible();
    expect(screen.getByText("OpenAI production")).toBeVisible();
  });
});
