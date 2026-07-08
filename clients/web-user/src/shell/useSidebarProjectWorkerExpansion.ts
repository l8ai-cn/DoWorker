import { useCallback, useEffect, useMemo, useState } from "react";
import {
  collectVisibleGroupedSessionIds,
  EXPANDED_WORKER_GROUP_SECTIONS_STORAGE_KEY,
  EXPANDED_WORKSPACE_GROUP_SECTIONS_STORAGE_KEY,
  type SidebarWorkerGroup,
  workProjectStorageKey,
} from "./sidebarNav";

function readExpandedKeys(storageKey: string): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(storageKey);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === "string") : [];
  } catch {
    return [];
  }
}

function writeExpandedKeys(storageKey: string, keys: string[]) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(storageKey, JSON.stringify(keys));
}

export function useSidebarProjectWorkerExpansion(
  groups: readonly SidebarWorkerGroup[],
  searchQuery: string,
  activeConversationId?: string,
) {
  const [expandedWorkers, setExpandedWorkers] = useState(() =>
    readExpandedKeys(EXPANDED_WORKER_GROUP_SECTIONS_STORAGE_KEY),
  );
  const [expandedProjects, setExpandedProjects] = useState(() =>
    readExpandedKeys(EXPANDED_WORKSPACE_GROUP_SECTIONS_STORAGE_KEY),
  );

  const toggleWorker = useCallback((key: string) => {
    setExpandedWorkers((prev) => {
      const next = prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key];
      writeExpandedKeys(EXPANDED_WORKER_GROUP_SECTIONS_STORAGE_KEY, next);
      return next;
    });
  }, []);

  const toggleProject = useCallback((key: string) => {
    setExpandedProjects((prev) => {
      const next = prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key];
      writeExpandedKeys(EXPANDED_WORKSPACE_GROUP_SECTIONS_STORAGE_KEY, next);
      return next;
    });
  }, []);

  const expandWorker = useCallback((key: string) => {
    setExpandedWorkers((prev) => {
      if (prev.includes(key)) return prev;
      const next = [...prev, key];
      writeExpandedKeys(EXPANDED_WORKER_GROUP_SECTIONS_STORAGE_KEY, next);
      return next;
    });
  }, []);

  const expandProject = useCallback((key: string) => {
    setExpandedProjects((prev) => {
      if (prev.includes(key)) return prev;
      const next = [...prev, key];
      writeExpandedKeys(EXPANDED_WORKSPACE_GROUP_SECTIONS_STORAGE_KEY, next);
      return next;
    });
  }, []);

  const effectiveExpandedWorkers = useMemo(() => {
    if (searchQuery) return new Set(groups.map((g) => g.key));
    return new Set(expandedWorkers);
  }, [searchQuery, groups, expandedWorkers]);

  const effectiveExpandedProjects = useMemo(() => {
    if (searchQuery) {
      return new Set(
        groups.flatMap((worker) =>
          worker.projectGroups.map((project) => workProjectStorageKey(worker.key, project.key)),
        ),
      );
    }
    return new Set(expandedProjects);
  }, [searchQuery, groups, expandedProjects]);

  const activeGroupKeys = useMemo(() => {
    if (!activeConversationId) return null;
    for (const worker of groups) {
      for (const project of worker.projectGroups) {
        if (project.conversations.some((c) => c.id === activeConversationId)) {
          return {
            workerKey: worker.key,
            projectKey: workProjectStorageKey(worker.key, project.key),
          };
        }
      }
    }
    return null;
  }, [activeConversationId, groups]);

  useEffect(() => {
    if (!activeGroupKeys) return;
    expandWorker(activeGroupKeys.workerKey);
    expandProject(activeGroupKeys.projectKey);
  }, [activeGroupKeys, expandWorker, expandProject]);

  const collectVisibleIds = useCallback(
    (sectionCollapsed: boolean) =>
      collectVisibleGroupedSessionIds(groups, {
        sectionCollapsed,
        expandedWorkerKeys: effectiveExpandedWorkers,
        expandedProjectKeys: effectiveExpandedProjects,
      }),
    [groups, effectiveExpandedWorkers, effectiveExpandedProjects],
  );

  return {
    effectiveExpandedWorkers,
    effectiveExpandedProjects,
    toggleWorker,
    toggleProject,
    collectVisibleIds,
  };
}
