import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/identity", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/identity")>()), authenticatedFetch: vi.fn(),
}));
vi.mock("@/lib/workerSessionPlan", () => ({
  buildSessionWorkerPlan: vi.fn(async ({ mode, modelResourceId }: { mode: string; modelResourceId?: number }) => ({
    worker_spec: {
      options_revision: "catalog-test", runtime_image_id: 11, placement_policy: "automatic",
      compute_target_id: 21, deployment_mode: "pooled", resource_profile_id: 31,
    },
    automation_level: mode === "pty" ? "interactive" : "autonomous",
    ...(modelResourceId === undefined ? {} : { model_resource_id: modelResourceId }),
  })),
}));
vi.mock("@/hooks/useHosts", () => ({ useHosts: vi.fn() }));
vi.mock("@/hooks/useAvailableAgents", () => ({ useAvailableAgents: vi.fn() }));
vi.mock("@/hooks/useModelConfigs", () => ({ useModelConfigs: vi.fn(), defaultModelConfig: vi.fn() }));
vi.mock("@/hooks/useHostFilesystem", () => ({
  useHostFilesystem: vi.fn(), useCreateHostDirectory: vi.fn(() => ({ mutateAsync: vi.fn(), isPending: false })),
}));
vi.mock("@/hooks/useDirectorySessions", () => ({ useDirectorySessions: vi.fn() }));
vi.mock("@/hooks/RunnerHealthProvider", () => ({ useRunnerHealthRegistration: vi.fn() }));
vi.mock("@/hooks/useConversations", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/hooks/useConversations")>()), useProjects: () => ({ data: [] }),
}));
vi.mock("@/lib/agentLabels", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/agentLabels")>()), useBrainHarnessLabels: () => ({}),
}));
vi.mock("@/store/chatStore", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/store/chatStore")>()), setPendingInitialPrompt: vi.fn(),
}));

import { resetLandingDraft } from "./NewChatDialog";
import {
  authenticatedFetchMock,
  mockAgents,
  mockModelConfigError,
  mockModelConfigs,
  renderLanding,
  setupLandingMocks,
} from "./newChatLandingTestHarness";

function createBody(): Record<string, unknown> {
  return JSON.parse((authenticatedFetchMock.mock.calls[0][1] as RequestInit).body as string);
}

function expectWorkerPlan(body: Record<string, unknown>, level: "autonomous" | "interactive") {
  expect(body.worker_spec).toEqual({
    options_revision: "catalog-test",
    runtime_image_id: 11,
    placement_policy: "automatic",
    compute_target_id: 21,
    deployment_mode: "pooled",
    resource_profile_id: 31,
  });
  expect(body.automation_level).toBe(level);
  expect(body).not.toHaveProperty("agent_slug");
  expect(body).not.toHaveProperty("model_override");
  expect(body).not.toHaveProperty("terminal_launch_args");
}

beforeEach(setupLandingMocks);
afterEach(() => {
  cleanup();
  resetLandingDraft();
  localStorage.clear();
});

describe("NewChatLandingScreen authoritative create", () => {
  it("renders a usable composer and creates an autonomous Worker plan", async () => {
    authenticatedFetchMock.mockResolvedValue({ ok: true, json: async () => ({ id: "conv_new" }) } as Response);
    renderLanding();
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-workspace-chip").textContent).toContain("repo"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Inspect the repository" } });
    fireEvent.click(screen.getByTestId("new-chat-landing-submit"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    const body = createBody();
    expect(body.agent_id).toBe("a1");
    expect(body.host_id).toBe("host_1");
    expect(body.workspace).toBe("/Users/corey/repo");
    expectWorkerPlan(body, "interactive");
  });

  it("shows the agent list without legacy command-line configuration controls", () => {
    renderLanding();
    fireEvent.pointerDown(screen.getByTestId("new-chat-landing-agent-select"), { button: 0 });
    expect(screen.getByTestId("new-chat-landing-agent-a1")).toBeTruthy();
    expect(screen.getByTestId("new-chat-landing-agent-a2")).toBeTruthy();
    expect(screen.queryByTestId("new-chat-landing-approval-full-access")).toBeNull();
    expect(screen.queryByTestId("new-chat-landing-permission-plan")).toBeNull();
    expect(screen.queryByTestId("new-chat-landing-model-opus")).toBeNull();
  });

  it("creates an empty managed sandbox with a Worker plan and no legacy host fields", async () => {
    authenticatedFetchMock.mockResolvedValue({ ok: true, json: async () => ({ id: "conv_new" }) } as Response);
    renderLanding({ managed_sandboxes_enabled: true });
    fireEvent.pointerDown(screen.getByTestId("new-chat-landing-host-chip"), { button: 0 });
    fireEvent.click(screen.getByTestId("new-chat-landing-sandbox-option"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Start fresh" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    const body = createBody();
    expect(body.agent_id).toBe("a1");
    expect(body).not.toHaveProperty("host_id");
    expect(body).not.toHaveProperty("workspace");
    expect(body).not.toHaveProperty("host_type");
    expectWorkerPlan(body, "interactive");
  });

  it("requires a real model resource when the selected WorkerTemplate requires one", async () => {
    mockAgents([{
      id: "codex-cli", name: "codex-cli", display_name: "Codex", description: null, harness: "codex",
      skills: [], workerTypeSlug: "codex-cli", supportedModes: ["acp"], requiresModelResource: true,
    }]);
    mockModelConfigs([{ id: 4, name: "OpenAI", provider_key: "openai", model: "gpt-5.5", is_default: true }]);
    authenticatedFetchMock.mockResolvedValue({ ok: true, json: async () => ({ id: "conv_codex" }) } as Response);
    renderLanding({}, "/?agent=codex-cli");
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-model-select").textContent).toContain("OpenAI"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Fix the bug" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    expect(createBody().model_resource_id).toBe(4);
  });

  it("drops a hidden model resource after switching to a WorkerTemplate that forbids it", async () => {
    mockAgents([
      {
        id: "codex-cli", name: "codex-cli", display_name: "Codex", description: null, harness: "codex",
        skills: [], workerTypeSlug: "codex-cli", supportedModes: ["acp"], requiresModelResource: true,
      },
      {
        id: "e2e-echo", name: "e2e-echo", display_name: "E2E Echo", description: null, harness: "echo",
        skills: [], workerTypeSlug: "e2e-echo", supportedModes: ["acp"], requiresModelResource: false,
      },
    ]);
    mockModelConfigs([{ id: 4, name: "OpenAI", provider_key: "openai", model: "gpt-5.5", is_default: true }]);
    authenticatedFetchMock.mockResolvedValue({ ok: true, json: async () => ({ id: "conv_echo" }) } as Response);
    renderLanding({}, "/?agent=codex-cli");
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-model-select")).toBeTruthy());
    fireEvent.pointerDown(screen.getByTestId("new-chat-landing-agent-select"), { button: 0 });
    fireEvent.click(screen.getByTestId("new-chat-landing-agent-e2e-echo"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Echo this" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    expect(createBody()).not.toHaveProperty("model_resource_id");
  });

  it("keeps model-required creation disabled when model resources fail to load", () => {
    mockAgents([{
      id: "codex-cli", name: "codex-cli", display_name: "Codex", description: null, harness: "codex",
      skills: [], workerTypeSlug: "codex-cli", supportedModes: ["acp"], requiresModelResource: true,
    }]);
    mockModelConfigError("model resource service unavailable");
    renderLanding({}, "/?agent=codex-cli");
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Fix the bug" } });
    expect(screen.getByTestId("new-chat-landing-submit")).toBeDisabled();
    expect(screen.getByText("model resource service unavailable")).toBeTruthy();
  });

  it("blocks repository-backed sandbox creation until WorkerTemplate workspace options exist", () => {
    renderLanding({ managed_sandboxes_enabled: true });
    fireEvent.pointerDown(screen.getByTestId("new-chat-landing-host-chip"), { button: 0 });
    fireEvent.click(screen.getByTestId("new-chat-landing-sandbox-option"));
    fireEvent.click(screen.getByTestId("new-chat-landing-repo-chip"));
    fireEvent.change(screen.getByTestId("new-chat-landing-repo-input"), {
      target: { value: "https://github.com/org/repo" },
    });
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Audit it" } });
    expect(screen.getByTestId("new-chat-landing-submit")).toBeDisabled();
    expect(authenticatedFetchMock).not.toHaveBeenCalled();
  });

  it("renders the server error without navigating after a rejected create", async () => {
    authenticatedFetchMock.mockResolvedValue({
      ok: false, status: 409, json: async () => ({ detail: "host is offline" }), text: async () => "host is offline",
    } as Response);
    renderLanding();
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-workspace-chip").textContent).toContain("repo"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Inspect it" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-error").textContent).toContain("host is offline"));
  });

  it("surfaces a project assignment failure instead of silently dropping the project", async () => {
    authenticatedFetchMock
      .mockResolvedValueOnce({ ok: true, json: async () => ({ id: "conv_new" }) } as Response)
      .mockResolvedValueOnce({ ok: false, status: 503, text: async () => "project service unavailable" } as Response);
    renderLanding({}, "/?project=project_1");
    await waitFor(() => expect(screen.getByTestId("new-chat-landing-workspace-chip").textContent).toContain("repo"));
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Organize this work" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(2));
    expect(authenticatedFetchMock.mock.calls[1]?.[0]).toBe("/v1/sessions/conv_new");
    expect(screen.getByTestId("new-chat-landing-error").textContent).toContain("project service unavailable");
  });
});
