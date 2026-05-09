import { vi } from "vitest";
import type { RepositoryData, WebhookStatus, WebhookSecretResponse } from "@/lib/api";
import { getRepositoryService, getApiClient } from "@/lib/wasm-core";

export const mockGetWebhookStatus = vi.fn();
export const mockGetWebhookSecret = vi.fn();
export const mockRegisterWebhook = vi.fn();
export const mockDeleteWebhook = vi.fn();
export const mockMarkWebhookConfigured = vi.fn();

export const mockClipboardWriteText = vi.fn();
Object.assign(navigator, {
  clipboard: {
    writeText: mockClipboardWriteText,
  },
});

export const mockRepository: RepositoryData = {
  id: 1,
  organization_id: 100,
  provider_type: "gitlab",
  provider_base_url: "https://gitlab.com",
  http_clone_url: "https://gitlab.com/org/repo.git",
  external_id: "123",
  name: "test-repo",
  slug: "org/test-repo",
  default_branch: "main",
  visibility: "organization",
  is_active: true,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

export const registeredStatus: WebhookStatus = {
  registered: true,
  webhook_id: "wh_123",
  webhook_url: "https://example.com/webhooks/org/gitlab/1",
  events: ["merge_request", "pipeline"],
  is_active: true,
  needs_manual_setup: false,
  registered_at: "2025-01-01T00:00:00Z",
};

export const manualSetupStatus: WebhookStatus = {
  registered: true,
  webhook_url: "https://example.com/webhooks/org/gitlab/1",
  events: ["merge_request", "pipeline"],
  is_active: false,
  needs_manual_setup: true,
  last_error: "OAuth token not available",
};

export const notRegisteredStatus: WebhookStatus = {
  registered: false,
  is_active: false,
  needs_manual_setup: false,
};

export const secretResponse: WebhookSecretResponse = {
  webhook_url: "https://example.com/webhooks/org/gitlab/1",
  webhook_secret: "super_secret_value",
  events: ["merge_request", "pipeline"],
};

const stableRepoSvc = {
  get_webhook_status: mockGetWebhookStatus,
  get_webhook_secret: mockGetWebhookSecret,
  delete_webhook: mockDeleteWebhook,
  register_webhook: mockRegisterWebhook,
  mark_webhook_configured: mockMarkWebhookConfigured,
  list: vi.fn().mockResolvedValue('{"repositories":[]}'),
  get: vi.fn().mockResolvedValue('{}'),
  create: vi.fn().mockResolvedValue('{}'),
  update: vi.fn().mockResolvedValue('{}'),
  delete: vi.fn().mockResolvedValue(undefined),
  list_branches: vi.fn().mockResolvedValue('{"branches":[]}'),
  sync_branches: vi.fn().mockResolvedValue('{"branches":[]}'),
  get_webhook_secret_for_setup: vi.fn().mockResolvedValue('{}'),
  list_merge_requests: vi.fn().mockResolvedValue('{"merge_requests":[]}'),
};

const stableClient = {
  get: vi.fn().mockResolvedValue('{}'),
  post: vi.fn().mockResolvedValue('{}'),
  put: vi.fn().mockResolvedValue('{}'),
  delete: vi.fn().mockResolvedValue('{}'),
  patch: vi.fn().mockResolvedValue('{}'),
  org_path: vi.fn((p: string) => `/api/v1/orgs/test-org${p}`),
};

export function setupWebhookMocks() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  vi.mocked(getRepositoryService).mockReturnValue(stableRepoSvc as any);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  vi.mocked(getApiClient).mockReturnValue(stableClient as any);
}

export function resetAllMocks() {
  mockGetWebhookStatus.mockReset();
  mockGetWebhookSecret.mockReset();
  mockRegisterWebhook.mockReset();
  mockDeleteWebhook.mockReset();
  mockMarkWebhookConfigured.mockReset();
  mockClipboardWriteText.mockReset();
  mockClipboardWriteText.mockResolvedValue(undefined);
  setupWebhookMocks();
}
