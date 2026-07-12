import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "./api-fetch";
import { createSession, getSessionByPodKey, listAgents } from "./sessions-api";

vi.mock("./api-fetch", () => ({
  apiFetch: vi.fn(),
}));

const apiFetchMock = vi.mocked(apiFetch);

describe("createSession model resources", () => {
  beforeEach(() => {
    apiFetchMock.mockReset();
  });

  it("submits the exact default model resource for model-backed agents", async () => {
    apiFetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            object: "list",
            data: [
              { id: 41, name: "Other", provider_key: "openai", model: "gpt-5", is_default: false },
              { id: 42, name: "Codex", provider_key: "openai", model: "gpt-5.5", is_default: true },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            id: "session-1",
            agent_id: "codex-cli",
            status: "launching",
          }),
          { status: 200 },
        ),
      );

    await createSession({ id: "codex-cli", requiresModelResource: true }, "Fix CI");

    expect(apiFetchMock).toHaveBeenNthCalledWith(1, "/v1/model-resources");
    const createInit = apiFetchMock.mock.calls[1][1] as RequestInit;
    expect(JSON.parse(createInit.body as string)).toMatchObject({
      agent_id: "codex-cli",
      model_resource_id: 42,
    });
  });

  it("uses the Worker model requirement instead of a slug allowlist", async () => {
    apiFetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            object: "list",
            data: [{ id: 42, is_default: true }],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            id: "session-minimax",
            agent_id: "minimax-cli",
            status: "launching",
          }),
          { status: 200 },
        ),
      );

    await createSession({ id: "minimax-cli", requiresModelResource: true }, "Run task");

    expect(apiFetchMock).toHaveBeenNthCalledWith(1, "/v1/model-resources");
    const createInit = apiFetchMock.mock.calls[1][1] as RequestInit;
    expect(JSON.parse(createInit.body as string)).toMatchObject({
      agent_id: "minimax-cli",
      model_resource_id: 42,
    });
  });

  it("fails before creation when no default model resource exists", async () => {
    apiFetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          object: "list",
          data: [
            { id: 41, name: "Other", provider_key: "openai", model: "gpt-5", is_default: false },
          ],
        }),
        { status: 200 },
      ),
    );

    await expect(
      createSession({ id: "codex-cli", requiresModelResource: true }, "Fix CI"),
    ).rejects.toThrow("No default model resource is configured");
    expect(apiFetchMock).toHaveBeenCalledTimes(1);
  });

  it("does not load a model resource for agents that do not require one", async () => {
    apiFetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          id: "session-2",
          agent_id: "custom-agent",
          status: "launching",
        }),
        { status: 200 },
      ),
    );

    await createSession({ id: "custom-agent", requiresModelResource: false }, "Run task");

    expect(apiFetchMock).toHaveBeenCalledTimes(1);
    const createInit = apiFetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(createInit.body as string)).not.toHaveProperty("model_resource_id");
  });

  it("marks command-line workers as PTY-only at creation", async () => {
    apiFetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            object: "list",
            data: [
              { id: 42, name: "Codex", provider_key: "openai", model: "gpt-5.5", is_default: true },
            ],
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({ id: "session-pty", agent_id: "codex-cli", status: "launching" }),
          { status: 200 },
        ),
      );

    await createSession(
      { id: "codex-cli", requiresModelResource: true },
      "Run command",
      undefined,
      {
        mode: "pty",
      },
    );

    const createInit = apiFetchMock.mock.calls[1][1] as RequestInit;
    expect(JSON.parse(createInit.body as string)).toMatchObject({ pty_only: true });
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
    expect(apiFetchMock).toHaveBeenCalledWith("/v1/sessions/by-pod/mobile-pod");
  });

  it("requires every Worker to declare its supported interaction modes", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: "codex-cli",
              name: "Codex",
              harness: "codex",
              supported_modes: ["acp", "pty"],
              requires_model_resource: true,
            },
            {
              id: "aider",
              name: "Aider",
              harness: "aider",
              supported_modes: ["pty"],
              requires_model_resource: false,
            },
          ],
        }),
        { status: 200 },
      ),
    );

    await expect(listAgents()).resolves.toEqual([
      {
        id: "codex-cli",
        name: "Codex",
        harness: "codex",
        supportedModes: ["acp", "pty"],
        requiresModelResource: true,
      },
      {
        id: "aider",
        name: "Aider",
        harness: "aider",
        supportedModes: ["pty"],
        requiresModelResource: false,
      },
    ]);
  });

  it("rejects a Worker catalog without a mode contract", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [{ id: "codex-cli", name: "Codex", requires_model_resource: true }],
        }),
        { status: 200 },
      ),
    );

    await expect(listAgents()).rejects.toThrow("未声明支持的交互模式");
  });

  it("rejects a Worker catalog without a model resource requirement", async () => {
    apiFetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [{ id: "codex-cli", name: "Codex", supported_modes: ["acp", "pty"] }],
        }),
        { status: 200 },
      ),
    );

    await expect(listAgents()).rejects.toThrow("未声明模型资源要求");
  });
});
