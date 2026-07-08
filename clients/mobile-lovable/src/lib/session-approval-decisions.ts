import { useSyncExternalStore } from "react";

export type ApprovalDecision = "approved" | "rejected";

let decisionsMap: Record<string, ApprovalDecision> = {};
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((l) => l());
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

function getSnapshot() {
  return decisionsMap;
}

export function useDecisions(): Record<string, ApprovalDecision> {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}

export function useDecision(id: string): ApprovalDecision | null {
  return useDecisions()[id] ?? null;
}

export function setApprovalDecision(id: string, d: ApprovalDecision) {
  if (decisionsMap[id] === d) return;
  decisionsMap = { ...decisionsMap, [id]: d };
  emit();
}
