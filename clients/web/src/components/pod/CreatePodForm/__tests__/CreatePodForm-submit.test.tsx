import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useWorkerCreateDraft } from "../../hooks/useWorkerCreateDraft";
import { createOptions, mockRepository, modelResource } from "./test-utils";

const mockPreflight = vi.fn();
const mockCreate = vi.fn();
const mockFill = vi.fn();

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    podApi: {
      listWorkerCreateOptions: vi.fn(),
      fillWorkerDraft: (...args: unknown[]) => mockFill(...args),
      preflightWorker: (...args: unknown[]) => mockPreflight(...args),
      create: (...args: unknown[]) => mockCreate(...args),
    },
  };
});
vi.mock("../../hooks/useWorkerCreateOptions", () => ({
  useWorkerCreateOptions: () => ({ status: "ready", data: createOptions() }),
}));
vi.mock("../../hooks/useWorkerCreateDependencies", () => ({
  useWorkerCreateDependencies: () => ({
    modelResources: { status: "ready", data: [modelResource()] },
    toolModelResources: { status: "ready", data: [] },
    runtimeBundles: { status: "ready", data: [] },
    credentialBundles: { status: "ready", data: [] },
    configBundles: { status: "ready", data: [] },
    skills: { status: "ready", data: [] },
  }),
}));
vi.mock("@/lib/terminal-size", () => ({
  estimateWorkspaceTerminalSize: () => ({ cols: 120, rows: 40 }),
}));

describe("useWorkerCreateDraft submission", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPreflight.mockResolvedValue({
      issues: [],
      resolved_spec_json: "{}",
      options_revision: "runtime-catalog-1",
    });
    mockCreate.mockResolvedValue({
      pod: { pod_key: "worker-1" },
    });
  });

  it("preflights and creates with only the structured WorkerSpec contract", async () => {
    const onSuccess = vi.fn();
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
        initialTask: "Fix the failing test.",
        ticketSlug: "TASK-7",
        onSuccess,
      }),
    );

    await waitFor(() => expect(result.current.validity.workspace).toBe(true));
    await act(async () => {
      await result.current.goToStep(4);
    });
    await act(async () => {
      await result.current.createWorker();
    });

    expect(mockPreflight).toHaveBeenCalledWith(result.current.state.draft);
    expect(mockCreate).toHaveBeenCalledWith({
      ticket_slug: "TASK-7",
      cols: 120,
      rows: 40,
      worker_spec: result.current.state.draft,
    });
    expect(onSuccess).toHaveBeenCalledWith(expect.objectContaining({
      pod_key: "worker-1",
    }));
  });

  it("does not create when preflight reports a blocking issue", async () => {
    mockPreflight.mockResolvedValue({
      issues: [{
        code: "invalid",
        field: "worker_spec.branch",
        message: "Branch is required",
        severity: "blocking",
      }],
      resolved_spec_json: "{}",
      options_revision: "runtime-catalog-1",
    });
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
      }),
    );

    await waitFor(() => expect(result.current.validity.workspace).toBe(true));
    await act(async () => {
      await result.current.goToStep(4);
    });
    await act(async () => {
      await result.current.createWorker();
    });
    expect(mockCreate).not.toHaveBeenCalled();
  });

  it("sends only one request for concurrent create attempts", async () => {
    const pending = deferred<{ pod: { pod_key: string } }>();
    mockCreate.mockReturnValueOnce(pending.promise);
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
      }),
    );

    await waitFor(() => expect(result.current.validity.workspace).toBe(true));
    await act(async () => {
      await result.current.goToStep(4);
    });

    let first!: ReturnType<typeof result.current.createWorker>;
    let second!: ReturnType<typeof result.current.createWorker>;
    await act(async () => {
      first = result.current.createWorker();
      second = result.current.createWorker();
      await Promise.resolve();
    });

    expect(mockCreate).toHaveBeenCalledTimes(1);
    pending.resolve({ pod: { pod_key: "worker-1" } });
    await act(async () => {
      await Promise.all([first, second]);
    });
  });

  it("keeps the create lock while an older request is still pending", async () => {
    const pending = deferred<{ pod: { pod_key: string } }>();
    mockCreate.mockReturnValueOnce(pending.promise);
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
      }),
    );

    await waitFor(() => expect(result.current.validity.workspace).toBe(true));
    await act(async () => {
      await result.current.goToStep(4);
    });

    let first!: ReturnType<typeof result.current.createWorker>;
    await act(async () => {
      first = result.current.createWorker();
      await Promise.resolve();
    });
    expect(result.current.state.create.status).toBe("loading");
    await act(async () => {
      result.current.patchDraft({ alias: "updated-draft" });
      await Promise.resolve();
    });
    expect(result.current.state.create.status).toBe("idle");
    await act(async () => {
      await result.current.goToStep(4);
    });
    expect(result.current.state.preflight.status).toBe("ready");
    await act(async () => {
      await result.current.createWorker();
    });

    expect(mockCreate).toHaveBeenCalledTimes(1);
    pending.resolve({ pod: { pod_key: "worker-1" } });
    await act(async () => {
      await first;
    });
  });

  it("does not create from a preflight result for an older draft", async () => {
    const pending = deferred<{
      issues: never[];
      resolved_spec_json: string;
      options_revision: string;
    }>();
    mockPreflight.mockReturnValueOnce(pending.promise);
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
      }),
    );
    await waitFor(() => expect(result.current.validity.workspace).toBe(true));

    let preflight!: ReturnType<typeof result.current.goToStep>;
    await act(async () => {
      preflight = result.current.goToStep(4);
      await Promise.resolve();
    });
    await waitFor(() => expect(mockPreflight).toHaveBeenCalled());
    act(() => result.current.patchDraft({ alias: "manual-edit" }));
    pending.resolve({
      issues: [],
      resolved_spec_json: "{}",
      options_revision: "runtime-catalog-1",
    });
    await act(async () => {
      await preflight;
      await result.current.createWorker();
    });

    expect(mockCreate).not.toHaveBeenCalled();
    expect(result.current.state.draft.alias).toBe("manual-edit");
    expect(result.current.state.preflight.status).toBe("idle");
  });

  it("does not replace manual edits with an older Fill with AI result", async () => {
    const pending = deferred<{
      draft: ReturnType<typeof useWorkerCreateDraft>["state"]["draft"];
      issues: never[];
    }>();
    mockFill.mockReturnValueOnce(pending.promise);
    const { result } = renderHook(() =>
      useWorkerCreateDraft({
        enabled: true,
        repositories: [mockRepository],
        initialWorkerTypeSlug: "codex-cli",
        initialRepositoryId: 51,
      }),
    );
    await waitFor(() => expect(result.current.validity.workspace).toBe(true));
    const sourceDraft = result.current.state.draft;

    let filling!: ReturnType<typeof result.current.fillWithAI>;
    await act(async () => {
      filling = result.current.fillWithAI("Configure the worker");
      await Promise.resolve();
    });
    act(() => result.current.patchDraft({ alias: "manual-edit" }));
    pending.resolve({
      draft: { ...sourceDraft, alias: "ai-edit" },
      issues: [],
    });
    await act(async () => {
      await filling;
    });

    expect(result.current.state.draft.alias).toBe("manual-edit");
    expect(result.current.state.fill.status).toBe("idle");
  });
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
