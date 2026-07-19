import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api-fetch";
import { createMobileWorkerSession } from "./mobile-session-creation";
import { createSession, getSessionByPodKey, listAgents } from "./sessions-api";

vi.mock("./api-fetch", () => ({ apiFetch: vi.fn() }));
vi.mock("./mobile-session-creation", () => ({ createMobileWorkerSession: vi.fn() }));

const apiFetchMock = vi.mocked(apiFetch);
const createWorkerSessionMock = vi.mocked(createMobileWorkerSession);

beforeEach(() => {
  apiFetchMock.mockReset();
  createWorkerSessionMock.mockReset();
});

describe("mobile sessions API", () => {
  it("delegates fresh creation to the authoritative Worker-plan builder", async () => {
    createWorkerSessionMock.mockResolvedValue({
      id: "session-1",
      agent_id: "agent_catalog_1",
      status: "launching",
    });

    await expect(
      createSession(
        {
          id: "agent_catalog_1",
          workerTypeSlug: "codex-cli",
          supportedModes: ["acp", "pty"],
          requiresModelResource: true,
        },
        "Fix CI",
      ),
    ).resolves.toMatchObject({ id: "session-1", agentId: "agent_catalog_1" });

    expect(createWorkerSessionMock).toHaveBeenCalledWith(
      expect.objectContaining({ workerTypeSlug: "codex-cli" }),
      "Fix CI",
      undefined,
      "acp",
    );
  });

  it("maps a builtin catalog agent to its authoritative worker type", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [{
            id: "codex-cli",
            name: "Codex",
            harness: "codex",
            builtin: true,
            supported_modes: ["acp", "pty"],
            requires_model_resource: true,
          }],
        }),
        { status: 200 },
      ),
    );

    await expect(listAgents()).resolves.toEqual([{
      id: "codex-cli",
      workerTypeSlug: "codex-cli",
      name: "Codex",
      harness: "codex",
      supportedModes: ["acp", "pty"],
      requiresModelResource: true,
    }]);
  });

  it("leaves a custom catalog agent uncreatable without worker type metadata", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [{
            id: "agent_123",
            name: "Custom",
            supported_modes: ["acp"],
            requires_model_resource: false,
          }],
        }),
        { status: 200 },
      ),
    );

    await expect(listAgents()).resolves.toMatchObject([{ workerTypeSlug: undefined }]);
  });

  it("resolves a mobile Worker link by its Pod key", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "session-1",
          pod_key: "mobile-pod",
          agent_id: "codex-cli",
          interaction_mode: "pty",
          status: "running",
        }),
        { status: 200 },
      ),
    );

    await expect(getSessionByPodKey("mobile-pod")).resolves.toMatchObject({
      id: "session-1",
      podKey: "mobile-pod",
      interactionMode: "pty",
    });
  });
});
