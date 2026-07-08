// Simple mock task queue per project.
// Real backend would send `session/new` sequentially over ACP; here we just
// simulate "one running, others pending" state in memory.

import { useSyncExternalStore } from "react";

export type QueuedTaskStatus = "pending" | "running" | "done";

export interface QueuedTask {
  id: string;
  projectId: string;
  prompt: string;
  agent: string;
  approvalMode: 0 | 1 | 2; // 自动 / 写审批 / 全审批
  createdAt: number;
  status: QueuedTaskStatus;
}

let tasks: QueuedTask[] = [
  // seed one demo for the API Gateway project so the UI has something to show
  {
    id: "q-seed-1",
    projectId: "p-api",
    prompt: "补一版 /orders 的 rate-limit 中间件",
    agent: "Codex",
    approvalMode: 1,
    createdAt: Date.now() - 20_000,
    status: "pending",
  },
];
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((l) => l());
  scheduleAdvance();
}

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

function getSnapshot() {
  return tasks;
}

/** Enqueue a task; returns the new id. First one auto-runs. */
export function enqueueTask(input: Omit<QueuedTask, "id" | "createdAt" | "status">): string {
  const id = `q-${Math.random().toString(36).slice(2, 8)}`;
  const task: QueuedTask = { ...input, id, createdAt: Date.now(), status: "pending" };
  tasks = [...tasks, task];
  emit();
  return id;
}

export function cancelTask(id: string) {
  tasks = tasks.filter((t) => t.id !== id);
  emit();
}

export function promoteTask(id: string) {
  const i = tasks.findIndex((t) => t.id === id);
  if (i <= 0) return;
  const next = tasks.slice();
  const [item] = next.splice(i, 1);
  // insert after any currently-running task in the same project
  const insertAt = next.findIndex(
    (t) => t.projectId === item.projectId && t.status !== "running",
  );
  next.splice(insertAt === -1 ? next.length : insertAt, 0, item);
  tasks = next;
  emit();
}

/** For each project, if none is running and one is pending, mark it running. */
let advanceTimer: ReturnType<typeof setTimeout> | null = null;
function scheduleAdvance() {
  if (advanceTimer) return;
  advanceTimer = setTimeout(() => {
    advanceTimer = null;
    let changed = false;
    const byProject = new Map<string, QueuedTask[]>();
    for (const t of tasks) {
      const arr = byProject.get(t.projectId) ?? [];
      arr.push(t);
      byProject.set(t.projectId, arr);
    }
    const next = tasks.slice();
    for (const [, arr] of byProject) {
      if (arr.some((t) => t.status === "running")) continue;
      const first = arr.find((t) => t.status === "pending");
      if (first) {
        const idx = next.indexOf(first);
        next[idx] = { ...first, status: "running" };
        changed = true;
      }
    }
    if (changed) {
      tasks = next;
      listeners.forEach((l) => l());
    }
  }, 800);
}

/* ---------- hooks ---------- */

export function useTaskQueue(projectId?: string) {
  const all = useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
  return projectId ? all.filter((t) => t.projectId === projectId) : all;
}

export function useProjectQueueCounts(projectId: string) {
  const list = useTaskQueue(projectId);
  return {
    pending: list.filter((t) => t.status === "pending").length,
    running: list.filter((t) => t.status === "running").length,
    total: list.length,
  };
}
