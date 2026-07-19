import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/identity", () => ({ authenticatedFetch: vi.fn() }));

import { authenticatedFetch } from "@/lib/identity";
import { fetchAvailableAgents } from "./availableAgentCatalogComposition";

const fetchMock = vi.mocked(authenticatedFetch);

function response(body: unknown): Response {
  return {
    ok: true,
    status: 200,
    statusText: "OK",
    json: async () => body,
  } as Response;
}

beforeEach(() => {
  fetchMock.mockReset();
  fetchMock.mockImplementation(async (input) => {
    switch (String(input)) {
      case "/v1/agents":
        return response({
          data: [{
            id: "agent_catalog_42",
            worker_type_slug: "codex-cli",
            supported_modes: ["acp", "pty"],
            requires_model_resource: true,
            name: "codex",
            builtin: true,
          }],
        });
      case "/v1/sessions?limit=100&kind=any":
        return response({
          data: [{
            id: "session_42",
            agent_id: "agent_session_opaque_7",
            agent_name: "custom-session-agent",
          }],
        });
      case "/v1/sessions/session_42/agent":
        return response({ id: "agent_session_opaque_7", name: "custom-session-agent" });
      default:
        throw new Error(`Unexpected request: ${String(input)}`);
    }
  });
});

describe("available Agent creation metadata", () => {
  it("keeps builtin agent id separate from its authoritative Worker type slug", async () => {
    const agent = (await fetchAvailableAgents()).find((item) => item.id === "agent_catalog_42");

    expect(agent).toMatchObject({
      id: "agent_catalog_42",
      workerTypeSlug: "codex-cli",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
    });
  });

  it("does not invent Worker creation metadata for a session-derived custom agent", async () => {
    const agent = (await fetchAvailableAgents()).find((item) => item.id === "agent_session_opaque_7");

    expect(agent).toBeDefined();
    expect(agent).not.toHaveProperty("workerTypeSlug");
    expect(agent).not.toHaveProperty("supportedModes");
    expect(agent).not.toHaveProperty("requiresModelResource");
  });
});
