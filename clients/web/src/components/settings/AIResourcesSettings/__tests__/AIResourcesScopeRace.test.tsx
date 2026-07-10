import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
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
  rotateConnectionCredentials: vi.fn(),
  setConnectionEnabled: vi.fn(),
  setDefaultResource: vi.fn(),
  setResourceEnabled: vi.fn(),
  updateConnection: vi.fn(),
  updateResource: vi.fn(),
  validateConnection: vi.fn(),
}));

vi.mock("@/lib/api", () => api);
vi.mock("next-intl", () => ({ useTranslations: () => (key: string) => key }));

import { AIResourcesSettings } from "../AIResourcesSettings";

const catalog = [{ key: "openai", displayName: "OpenAI", modalities: ["chat"], credentialFields: [], defaultBaseUrl: "", protocolAdapter: "openai", supportsCustomEndpoint: false, supportsModelDiscovery: false }];
const personalConnection = [{ id: 7, ownerScope: "personal", identifier: "personal-openai", providerKey: "openai", name: "Personal OpenAI", baseUrl: "", configuredFields: [], status: "valid", isEnabled: true, validationError: "", canManage: true, resources: [] }];
const organizationConnection = [{ ...personalConnection[0], ownerScope: "organization", name: "Organization OpenAI" }];

describe("AIResourcesSettings scope races", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.getCatalog.mockResolvedValue(catalog);
    api.listPersonalConnections.mockResolvedValue(personalConnection);
    api.listOrganizationConnections.mockResolvedValue(organizationConnection);
    api.listPersonalEffectiveResources.mockResolvedValue([]);
    api.listOrganizationEffectiveResources.mockResolvedValue([]);
  });

  it("does not let a completed personal request replace newer organization data", async () => {
    let resolvePersonal: (value: typeof personalConnection) => void = () => undefined;
    const personalRequest = new Promise<typeof personalConnection>((resolve) => { resolvePersonal = resolve; });
    api.listPersonalConnections.mockReturnValueOnce(personalRequest);
    const { rerender } = render(<AIResourcesSettings scope="personal" canManage />);

    await waitFor(() => expect(api.listPersonalConnections).toHaveBeenCalledOnce());
    rerender(<AIResourcesSettings scope="organization" organizationSlug="acme" canManage />);
    expect(await screen.findByText("Organization OpenAI")).toBeVisible();
    await act(async () => { resolvePersonal(personalConnection); await personalRequest; });

    expect(screen.queryByText("Personal OpenAI")).not.toBeInTheDocument();
  });

  it("does not reload the former scope when its mutation completes after a scope change", async () => {
    let resolveValidation: () => void = () => undefined;
    api.validateConnection.mockReturnValueOnce(new Promise<void>((resolve) => { resolveValidation = resolve; }));
    const { rerender } = render(<AIResourcesSettings scope="personal" canManage />);

    fireEvent.click(await screen.findByRole("button", { name: "settings.aiResources.validate" }));
    rerender(<AIResourcesSettings scope="organization" organizationSlug="acme" canManage />);
    expect(await screen.findByText("Organization OpenAI")).toBeVisible();
    await act(async () => { resolveValidation(); await Promise.resolve(); });

    await waitFor(() => expect(api.listPersonalConnections).toHaveBeenCalledOnce());
    expect(screen.queryByText("Personal OpenAI")).not.toBeInTheDocument();
  });
});
