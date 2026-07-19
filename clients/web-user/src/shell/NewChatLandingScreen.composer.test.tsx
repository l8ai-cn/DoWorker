import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/identity", () => ({ authenticatedFetch: vi.fn() }));
vi.mock("@/lib/workerSessionPlan", () => ({
  buildSessionWorkerPlan: vi.fn(async ({ mode }: { mode: string }) => ({
    worker_spec: {
      options_revision: "catalog-test", runtime_image_id: 11, placement_policy: "automatic",
      compute_target_id: 21, deployment_mode: "pooled", resource_profile_id: 31,
    },
    automation_level: mode === "pty" ? "interactive" : "autonomous",
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
  renderLanding,
  setPendingInitialPromptMock,
  setupLandingMocks,
  useHostFilesystemMock,
} from "./newChatLandingTestHarness";

function success(id = "conv_new") {
  authenticatedFetchMock.mockResolvedValue({ ok: true, json: async () => ({ id }) } as Response);
}

function createBody(): Record<string, unknown> {
  return JSON.parse((authenticatedFetchMock.mock.calls[0][1] as RequestInit).body as string);
}

async function ready() {
  await waitFor(() => expect(screen.getByTestId("new-chat-landing-workspace-chip").textContent).toContain("repo"));
}

beforeEach(setupLandingMocks);
afterEach(() => {
  cleanup();
  resetLandingDraft();
  localStorage.clear();
});

describe("NewChatLandingScreen first message handoff", () => {
  it("creates a plan with an empty history and hands the sanitized prompt to ChatPage", async () => {
    success();
    renderLanding();
    await ready();
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "  read\x07 README  " } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    const body = createBody();
    expect(body.initial_items).toEqual([]);
    expect(body.automation_level).toBe("interactive");
    expect(body).not.toHaveProperty("agent_slug");
    expect(body).not.toHaveProperty("model_override");
    expect(setPendingInitialPromptMock).toHaveBeenCalledWith("conv_new", {
      text: "read README", skill: null, files: [],
    });
  });

  it("carries a selected PNG through the pending first prompt", async () => {
    success();
    renderLanding();
    await ready();
    const file = new File(["image"], "diagram.png", { type: "image/png" });
    fireEvent.change(screen.getByTestId("new-chat-landing-file-input"), { target: { files: [file] } });
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Explain this" } });
    fireEvent.click(screen.getByTestId("new-chat-landing-submit"));
    await waitFor(() => expect(setPendingInitialPromptMock).toHaveBeenCalledWith("conv_new", {
      text: "Explain this", skill: null, files: [file],
    }));
  });

  it("removes an attached file before starting the session", async () => {
    renderLanding();
    const file = new File(["draft"], "notes.txt", { type: "text/plain" });
    fireEvent.change(screen.getByTestId("new-chat-landing-file-input"), { target: { files: [file] } });
    expect(screen.getByText("notes.txt")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Remove notes.txt" }));
    expect(screen.queryByText("notes.txt")).toBeNull();
  });

  it("accepts dropped files and clears the drop overlay", () => {
    renderLanding();
    const file = new File(["report"], "report.csv", { type: "text/csv" });
    const composer = screen.getByTestId("new-chat-landing-composer");
    fireEvent.dragEnter(composer, { dataTransfer: { files: [file] } });
    expect(screen.getByText("Drop files here")).toBeTruthy();
    fireEvent.drop(composer, { dataTransfer: { files: [file] } });
    expect(screen.getByText("report.csv")).toBeTruthy();
    expect(screen.queryByText("Drop files here")).toBeNull();
  });

  it("accepts DOCX and explains a rejected file type", () => {
    renderLanding();
    const input = screen.getByTestId("new-chat-landing-file-input");
    const docx = new File(["brief"], "brief.docx", {
      type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    });
    fireEvent.change(input, { target: { files: [docx] } });
    expect(screen.getByText("brief.docx")).toBeTruthy();
    const unsupported = new File(["video"], "clip.mp4", { type: "video/mp4" });
    fireEvent.change(input, { target: { files: [unsupported] } });
    expect(screen.queryByText("clip.mp4")).toBeNull();
    expect(screen.getByTestId("new-chat-landing-attachment-error").textContent).toContain("clip.mp4");
  });

  it("preserves an unfinished message and attachments after the landing screen remounts", async () => {
    const first = renderLanding();
    await ready();
    const file = new File(["draft"], "brief.docx", {
      type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    });
    fireEvent.change(screen.getByTestId("new-chat-landing-file-input"), { target: { files: [file] } });
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Continue this work" } });
    first.unmount();
    renderLanding();
    await ready();
    expect(screen.getByTestId("new-chat-landing-input")).toHaveValue("Continue this work");
    expect(screen.getByText("brief.docx")).toBeTruthy();
  });

  it("hands a bundled skill to the session page as a structured first turn", async () => {
    success();
    mockAgents([{
      id: "reviewer", name: "reviewer", display_name: "Reviewer", description: null, harness: "claude-sdk",
      skills: [{ name: "review-pr", description: "Review a pull request" }],
    }]);
    renderLanding();
    await ready();
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "/review-pr 123" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(setPendingInitialPromptMock).toHaveBeenCalledWith("conv_new", {
      text: "/review-pr 123", skill: { name: "review-pr", args: "123" }, files: [],
    }));
  });

  it("selects a skill from the slash menu without submitting", () => {
    mockAgents([{
      id: "reviewer", name: "reviewer", display_name: "Reviewer", description: null, harness: "claude-sdk",
      skills: [{ name: "review-pr", description: "Review a pull request" }],
    }]);
    renderLanding();
    const input = screen.getByTestId("new-chat-landing-input");
    fireEvent.change(input, { target: { value: "/rev" } });
    expect(screen.getByTestId("slash-menu-item-review-pr")).toBeTruthy();
    fireEvent.keyDown(input, { key: "Tab" });
    expect(input).toHaveValue("/review-pr ");
    expect(authenticatedFetchMock).not.toHaveBeenCalled();
  });

  it("attaches a native-agent workspace file mention", async () => {
    useHostFilesystemMock.mockReturnValue({
      data: {
        entries: [{
          name: "README.md",
          path: "/Users/corey/repo/README.md",
          type: "file",
          bytes: 10,
          modified_at: 0,
        }],
      },
      isLoading: false,
      error: null,
      isPlaceholderData: false,
    } as ReturnType<typeof useHostFilesystemMock>);
    renderLanding();
    await ready();
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Inspect @" } });
    await waitFor(() => expect(screen.getByTestId("file-mention-item-0")).toBeTruthy());
    fireEvent.click(screen.getByTitle("Attach README.md"));
    expect(screen.getByText("@README.md")).toBeTruthy();
  });

  it("leaves native-agent slash input as literal text and requests interactive automation", async () => {
    success("conv_native");
    mockAgents([{
      id: "native", name: "claude-native-ui", display_name: "Claude Code", description: null,
      harness: "claude-native", skills: [{ name: "review-pr", description: "Review" }],
    }]);
    renderLanding();
    await ready();
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "/review-pr 123" } });
    fireEvent.submit(screen.getByTestId("new-chat-landing-composer"));
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
    expect(createBody().automation_level).toBe("interactive");
    expect(setPendingInitialPromptMock).toHaveBeenCalledWith("conv_native", {
      text: "/review-pr 123", skill: null, files: [],
    });
  });

  it("does not submit while an IME composition is active", async () => {
    success();
    renderLanding();
    await ready();
    const input = screen.getByTestId("new-chat-landing-input");
    fireEvent.change(input, { target: { value: "分析" } });
    fireEvent.compositionStart(input);
    fireEvent.keyDown(input, { key: "Enter" });
    expect(authenticatedFetchMock).not.toHaveBeenCalled();
    fireEvent.compositionEnd(input);
    fireEvent.keyDown(input, { key: "Enter" });
    await waitFor(() => expect(authenticatedFetchMock).toHaveBeenCalledTimes(1));
  });

  it("blocks a requested git worktree instead of sending a legacy git body", async () => {
    renderLanding();
    await ready();
    fireEvent.click(screen.getByTestId("new-chat-landing-branch-chip"));
    fireEvent.change(screen.getByTestId("new-chat-landing-branch-input"), { target: { value: "feature/login" } });
    fireEvent.change(screen.getByTestId("new-chat-landing-input"), { target: { value: "Start work" } });
    expect(screen.getByTestId("new-chat-landing-submit")).toBeDisabled();
    expect(authenticatedFetchMock).not.toHaveBeenCalled();
  });
});
