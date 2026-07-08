import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";

// Networking stays on expertApi (JSON lightFetch); mock it so the store's
// actions fold controlled payloads into the Rust-backed cache (mocked in
// clients/web/src/test/setup.ts as an in-memory WasmExpertState fake).
const listMock = vi.fn();
const getMock = vi.fn();
const deleteMock = vi.fn();
const runMock = vi.fn();
const publishMock = vi.fn();
const createMock = vi.fn();
const updateMock = vi.fn();

vi.mock("@/lib/api/expertApi", () => ({
  expertApi: {
    list: (...a: unknown[]) => listMock(...a),
    get: (...a: unknown[]) => getMock(...a),
    delete: (...a: unknown[]) => deleteMock(...a),
    run: (...a: unknown[]) => runMock(...a),
    publishFromPod: (...a: unknown[]) => publishMock(...a),
    create: (...a: unknown[]) => createMock(...a),
    update: (...a: unknown[]) => updateMock(...a),
  },
}));

import { useExpertStore, useExperts, useCurrentExpert } from "../expert";

function makeExpert(slug: string) {
  return { id: 1, slug, name: slug.toUpperCase(), agent_slug: "claude-code" };
}

describe("Expert store (Rust SSOT mirror)", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    listMock.mockResolvedValue({ experts: [], total: 0 });
    getMock.mockResolvedValue(makeExpert("noop"));
    await act(async () => {
      await useExpertStore.getState().fetchExperts();
      useExpertStore.getState().clearCurrentExpert();
    });
  });

  it("fetchExperts folds the list into the Rust cache read by useExperts", async () => {
    listMock.mockResolvedValue({ experts: [makeExpert("alpha"), makeExpert("beta")], total: 2 });
    const { result } = renderHook(() => useExperts());
    await act(async () => {
      await useExpertStore.getState().fetchExperts();
    });
    expect(result.current.map((e) => e.slug)).toEqual(["alpha", "beta"]);
    expect(useExpertStore.getState().loading).toBe(false);
  });

  it("fetchExpert folds the current expert read by useCurrentExpert", async () => {
    getMock.mockResolvedValue(makeExpert("reviewer"));
    const { result } = renderHook(() => useCurrentExpert());
    await act(async () => {
      await useExpertStore.getState().fetchExpert("reviewer");
    });
    expect(result.current?.slug).toBe("reviewer");
  });

  it("deleteExpert removes from cache and clears matching current", async () => {
    listMock.mockResolvedValue({ experts: [makeExpert("alpha")], total: 1 });
    getMock.mockResolvedValue(makeExpert("alpha"));
    const experts = renderHook(() => useExperts());
    const current = renderHook(() => useCurrentExpert());
    await act(async () => {
      await useExpertStore.getState().fetchExperts();
      await useExpertStore.getState().fetchExpert("alpha");
    });
    expect(experts.result.current).toHaveLength(1);
    await act(async () => {
      await useExpertStore.getState().deleteExpert("alpha");
    });
    expect(experts.result.current).toHaveLength(0);
    expect(current.result.current).toBeNull();
  });

  it("createExpert posts and refreshes the list into the cache", async () => {
    createMock.mockResolvedValue(makeExpert("reviewer"));
    listMock.mockResolvedValue({ experts: [makeExpert("reviewer")], total: 1 });
    const { result } = renderHook(() => useExperts());
    let created: unknown;
    await act(async () => {
      created = await useExpertStore.getState().createExpert({
        name: "Reviewer",
        slug: "reviewer",
        agent_slug: "claude-code",
      });
    });
    expect(createMock).toHaveBeenCalledOnce();
    expect((created as { slug: string }).slug).toBe("reviewer");
    expect(result.current.map((e) => e.slug)).toEqual(["reviewer"]);
  });

  it("updateExpert patches and folds the result into current + list", async () => {
    updateMock.mockResolvedValue({ ...makeExpert("alpha"), name: "Renamed" });
    listMock.mockResolvedValue({ experts: [{ ...makeExpert("alpha"), name: "Renamed" }], total: 1 });
    const current = renderHook(() => useCurrentExpert());
    await act(async () => {
      await useExpertStore.getState().updateExpert("alpha", {
        name: "Renamed",
        skill_slugs: ["s1"],
      });
    });
    expect(updateMock).toHaveBeenCalledWith("alpha", { name: "Renamed", skill_slugs: ["s1"] });
    expect(current.result.current?.name).toBe("Renamed");
  });

  it("fetchExperts records error and stops loading on network failure", async () => {
    listMock.mockRejectedValue(new Error("boom"));
    await act(async () => {
      await useExpertStore.getState().fetchExperts();
    });
    expect(useExpertStore.getState().error).toBe("boom");
    expect(useExpertStore.getState().loading).toBe(false);
  });
});
