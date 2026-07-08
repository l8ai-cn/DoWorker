import { useSyncExternalStore } from "react";
import { projects, type Project } from "./mock-agents";

const listeners = new Set<() => void>();
let snapshot: Project[] = projects.slice();
const getSnapshot = () => snapshot;
const notify = () => {
  snapshot = projects.slice();
  listeners.forEach((l) => l());
};

export function addProject(input: {
  name: string;
  repo: string;
  host: string;
  color?: string;
}): Project {
  const id = `p-${Math.random().toString(36).slice(2, 7)}`;
  const p: Project = {
    id,
    name: input.name.trim() || "未命名项目",
    repo: input.repo.trim() || "acme/new-project",
    host: input.host.trim() || "local-devbox",
    color: input.color || "primary",
    online: true,
    sessionIds: [],
  };
  projects.push(p);
  notify();
  return p;
}

export function useProjects(): Project[] {
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb);
      return () => listeners.delete(cb);
    },
    getSnapshot,
    getSnapshot,
  );
}
