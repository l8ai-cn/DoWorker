import { create } from "zustand";

export type DoAgentGoal = {
  id: string;
  title?: string;
  status?: string;
};

type DoAgentConsoleState = {
  goalsByPod: Record<string, DoAgentGoal[]>;
  lastResponseKey: Record<string, number>;
  setGoals: (podKey: string, goals: DoAgentGoal[]) => void;
  touchResponse: (podKey: string) => void;
};

export const useDoAgentConsoleStore = create<DoAgentConsoleState>((set) => ({
  goalsByPod: {},
  lastResponseKey: {},
  setGoals: (podKey, goals) =>
    set((s) => ({ goalsByPod: { ...s.goalsByPod, [podKey]: goals } })),
  touchResponse: (podKey) =>
    set((s) => ({
      lastResponseKey: { ...s.lastResponseKey, [podKey]: Date.now() },
    })),
}));

// Stable empty reference — returning a fresh `[]` from the selector makes
// zustand's useSyncExternalStore snapshot differ every render, triggering an
// infinite re-render loop ("Maximum update depth exceeded").
const EMPTY_GOALS: DoAgentGoal[] = [];

export function useDoAgentGoals(podKey: string): DoAgentGoal[] {
  return useDoAgentConsoleStore((s) => s.goalsByPod[podKey] ?? EMPTY_GOALS);
}

export function useDoAgentResponseTick(podKey: string): number {
  return useDoAgentConsoleStore((s) => s.lastResponseKey[podKey] ?? 0);
}
