// Rust SSOT (Phase 3): the canonical expert list/current cache lives in the
// Rust core (`AppState.experts`), reached via the `WasmExpertState` view.
// This store is a thin mirror — it holds only ephemeral request status
// (loading/error) plus a `_tick` that invalidates the Rust-backed selectors
// below (same pattern as `stores/repository.ts`). Networking stays on the
// shared `expertApi` (lightFetch); expert has no proto/Connect coverage, so
// the fold crosses the wasm boundary as JSON rather than prost bytes.
import { create } from "zustand";
import { useMemo } from "react";
import { getExpertState } from "@/lib/wasm-core";
import {
  expertApi,
  type CreateExpertInput,
  type Expert,
  type PublishExpertInput,
  type RunExpertInput,
  type UpdateExpertInput,
} from "@/lib/api/expertApi";
import type { PodData } from "@/lib/api/facade/pod";

interface ExpertState {
  _tick: number;
  loading: boolean;
  expertLoading: boolean;
  error: string | null;

  fetchExperts: () => Promise<void>;
  fetchExpert: (slug: string) => Promise<void>;
  clearCurrentExpert: () => void;
  createExpert: (input: CreateExpertInput) => Promise<Expert>;
  updateExpert: (slug: string, input: UpdateExpertInput) => Promise<Expert>;
  deleteExpert: (slug: string) => Promise<void>;
  runExpert: (slug: string, input?: RunExpertInput) => Promise<{ pod: PodData; warning?: string }>;
  publishFromPod: (podKey: string, input: PublishExpertInput) => Promise<Expert>;
  clearError: () => void;
}

function message(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

const es = () => getExpertState();
const bump = () => useExpertStore.setState((s) => ({ _tick: s._tick + 1 }));

export const useExpertStore = create<ExpertState>((set, get) => ({
  _tick: 0,
  loading: false,
  expertLoading: false,
  error: null,

  fetchExperts: async () => {
    set({ loading: true, error: null });
    try {
      const result = await expertApi.list();
      es().apply_fetched_experts(JSON.stringify(result));
      bump();
      set({ loading: false });
    } catch (e) {
      set({ error: message(e), loading: false });
    }
  },

  fetchExpert: async (slug) => {
    set({ expertLoading: true, error: null });
    try {
      const expert = await expertApi.get(slug);
      es().apply_fetched_expert(JSON.stringify({ expert }));
      bump();
      set({ expertLoading: false });
    } catch (e) {
      es().clear_current_expert();
      bump();
      set({ error: message(e), expertLoading: false });
    }
  },

  clearCurrentExpert: () => {
    es().clear_current_expert();
    bump();
  },

  createExpert: async (input) => {
    const expert = await expertApi.create(input);
    await get().fetchExperts();
    return expert;
  },

  updateExpert: async (slug, input) => {
    const expert = await expertApi.update(slug, input);
    es().apply_fetched_expert(JSON.stringify({ expert }));
    bump();
    await get().fetchExperts();
    return expert;
  },

  deleteExpert: async (slug) => {
    await expertApi.delete(slug);
    es().remove_expert(slug);
    bump();
  },

  runExpert: async (slug, input) => {
    const result = await expertApi.run(slug, input);
    await get().fetchExpert(slug);
    await get().fetchExperts();
    return result;
  },

  publishFromPod: async (podKey, input) => {
    const expert = await expertApi.publishFromPod(podKey, input);
    await get().fetchExperts();
    return expert;
  },

  clearError: () => set({ error: null }),
}));

export function useExperts(): Expert[] {
  const tick = useExpertStore((s) => s._tick);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  return useMemo(() => {
    try {
      const parsed = JSON.parse(es().experts_json());
      return Array.isArray(parsed) ? (parsed as Expert[]) : [];
    } catch {
      return [];
    }
  }, [tick]);
}

export function useCurrentExpert(): Expert | null {
  const tick = useExpertStore((s) => s._tick);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  return useMemo(() => {
    const raw = es().current_expert_json();
    if (typeof raw !== "string") return null;
    try {
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === "object" && !Array.isArray(parsed)
        ? (parsed as Expert)
        : null;
    } catch {
      return null;
    }
  }, [tick]);
}
