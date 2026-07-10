import { create } from "zustand";
import { persist } from "zustand/middleware";
import { legacyPersistStorage } from "@/lib/zustand-legacy-persist";

/**
 * Pod creation preferences - remembers last user choices.
 *
 * EnvBundle preferences are split by kind to mirror the dialog UI:
 *   - `lastCredentialName`: single-select pick (empty = "use Agent default")
 *   - `lastRuntimeBundleNames`: ordered list of runtime bundle names
 *
 * Names (not IDs) are stored because bundle names are stable across
 * rename/recreate while IDs are not.
 */
interface PodCreationPreferences {
  lastAgentSlug: string | null;
  lastRepositoryId: number | null;
  lastCredentialName: string;
  lastRuntimeBundleNames: string[];
  lastBranchName: string | null;
  // Slugs are stable across rename/reinstall and repo-scoped; restored as the
  // initial selection filtered to the skills actually installed on the repo.
  lastSkillSlugs: string[];
  lastDestroyPolicy: "manual" | "idle" | "completed";
  lastDestroyAfterMinutes: number;
  lastKnowledgeMounts: { slug: string; mode: "ro" | "rw" }[];

  setLastChoices: (
    choices: Partial<
      Pick<
        PodCreationPreferences,
        | "lastAgentSlug"
        | "lastRepositoryId"
        | "lastCredentialName"
        | "lastRuntimeBundleNames"
        | "lastBranchName"
        | "lastSkillSlugs"
        | "lastDestroyPolicy"
        | "lastDestroyAfterMinutes"
        | "lastKnowledgeMounts"
      >
    >
  ) => void;
  clearLastChoices: () => void;

  // Hydration state for SSR
  _hasHydrated: boolean;
  setHasHydrated: (state: boolean) => void;
}

export const usePodCreationStore = create<PodCreationPreferences>()(
  persist(
    (set) => ({
      lastAgentSlug: null,
      lastRepositoryId: null,
      lastCredentialName: "",
      lastRuntimeBundleNames: [],
      lastBranchName: null,
      lastSkillSlugs: [],
      lastDestroyPolicy: "manual",
      lastDestroyAfterMinutes: 120,
      lastKnowledgeMounts: [],

      setLastChoices: (choices) => set((state) => ({ ...state, ...choices })),
      clearLastChoices: () =>
        set({
          lastAgentSlug: null,
          lastRepositoryId: null,
          lastCredentialName: "",
          lastRuntimeBundleNames: [],
          lastBranchName: null,
          lastSkillSlugs: [],
          lastDestroyPolicy: "manual",
          lastDestroyAfterMinutes: 120,
          lastKnowledgeMounts: [],
        }),

      // Hydration
      _hasHydrated: false,
      setHasHydrated: (state) => set({ _hasHydrated: state }),
    }),
    {
      name: "do-worker-pod-creation",
      storage: legacyPersistStorage("agentsmesh-pod-creation"),
      version: 6,
      // v1 stored `lastBundleName: string | null`; v2 unified into
      // `lastBundleNames: string[]`; v3 splits back into credential
      // (single) + runtime (multi) to match the dialog UI. Legacy values
      // are dropped — we can't classify a name without re-querying the
      // bundle list, and the user will see their primary bundles
      // re-applied on next agent select anyway. v4 adds `lastSkillSlugs`.
      migrate: (persistedState: unknown, version: number) => {
        const s = (persistedState as Record<string, unknown>) ?? {};
        if (version < 3) {
          delete s.lastBundleName;
          delete s.lastBundleNames;
          s.lastCredentialName = "";
          s.lastRuntimeBundleNames = [];
        }
        if (version < 4) {
          s.lastSkillSlugs = [];
        }
        if (version < 5) {
          s.lastDestroyPolicy = "manual";
          s.lastDestroyAfterMinutes = 120;
        }
        if (version < 6) {
          // v5 stored blockstore workspace IDs; the git-backed KB feature
          // keys mounts by {slug, mode} so old values can't be migrated.
          delete s.lastKnowledgeBaseIds;
          s.lastKnowledgeMounts = [];
        }
        return s as unknown as PodCreationPreferences;
      },
      partialize: (state) => ({
        lastAgentSlug: state.lastAgentSlug,
        lastRepositoryId: state.lastRepositoryId,
        lastCredentialName: state.lastCredentialName,
        lastRuntimeBundleNames: state.lastRuntimeBundleNames,
        lastBranchName: state.lastBranchName,
        lastSkillSlugs: state.lastSkillSlugs,
        lastDestroyPolicy: state.lastDestroyPolicy,
        lastDestroyAfterMinutes: state.lastDestroyAfterMinutes,
        lastKnowledgeMounts: state.lastKnowledgeMounts,
      }),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      },
    }
  )
);
