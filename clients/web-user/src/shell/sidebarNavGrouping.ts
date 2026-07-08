import type { Conversation } from "@/hooks/useConversations";
import { PROJECT_LABEL_KEY } from "@/hooks/useConversations";
import {
  DEFAULT_PROJECT_GROUP_KEY,
  DEFAULT_PROJECT_GROUP_LABEL,
  type ActiveChatOverride,
  formatWorkerDisplayName,
  getConversationAgentType,
  sortByUpdatedAtDesc,
} from "./sidebarNav";

function workerGroupKey(conversation: Conversation): string {
  return getConversationAgentType(conversation);
}

function pathBasename(path: string): string {
  const normalized = path.replace(/\\/g, "/").replace(/\/+$/, "");
  const slash = normalized.lastIndexOf("/");
  return slash >= 0 ? normalized.slice(slash + 1) : normalized;
}

export function workProjectGroupKey(conversation: Conversation): string {
  const manual = conversation.labels?.[PROJECT_LABEL_KEY]?.trim();
  if (manual) return `project:${manual}`;
  const workspace = conversation.workspace?.trim() ?? "";
  if (workspace) return `workspace:${workspace}`;
  return DEFAULT_PROJECT_GROUP_KEY;
}

export function workProjectGroupLabel(conversation: Conversation): string {
  const manual = conversation.labels?.[PROJECT_LABEL_KEY]?.trim();
  if (manual) return manual;
  const workspace = conversation.workspace?.trim();
  if (workspace) return pathBasename(workspace);
  return DEFAULT_PROJECT_GROUP_LABEL;
}

export function workProjectWorkspacePath(conversations: Conversation[]): string | null {
  for (const conversation of conversations) {
    const workspace = conversation.workspace?.trim();
    if (workspace) return workspace;
  }
  return null;
}

export interface WorkProjectGitInfo {
  linked: boolean;
  branchLabel: string | null;
  hasWorkspace: boolean;
}

export function workProjectGitInfo(conversations: Conversation[]): WorkProjectGitInfo {
  const branches = [
    ...new Set(
      conversations
        .map((c) => c.git_branch?.trim())
        .filter((branch): branch is string => !!branch),
    ),
  ];
  const hasWorkspace = conversations.some((c) => !!c.workspace?.trim());
  if (branches.length === 1) {
    return { linked: true, branchLabel: branches[0]!, hasWorkspace };
  }
  if (branches.length > 1) {
    return { linked: true, branchLabel: `${branches.length} branches`, hasWorkspace };
  }
  return { linked: false, branchLabel: null, hasWorkspace };
}

export type WorkerConnectionStatus = "online" | "offline" | "mixed" | "unknown";

export function workerConnectionStatus(
  conversation: Conversation,
): Exclude<WorkerConnectionStatus, "mixed"> {
  if (conversation.runner_online === true) return "online";
  if (conversation.runner_online === false) return "offline";
  return "unknown";
}

export function aggregateWorkerConnectionStatus(
  conversations: Conversation[],
): WorkerConnectionStatus {
  let sawOnline = false;
  let sawOffline = false;
  for (const conversation of conversations) {
    const status = workerConnectionStatus(conversation);
    if (status === "online") sawOnline = true;
    if (status === "offline") sawOffline = true;
  }
  if (sawOnline && sawOffline) return "mixed";
  if (sawOnline) return "online";
  if (sawOffline) return "offline";
  return "unknown";
}

export interface SidebarWorkProjectGroup {
  key: string;
  label: string;
  workspacePath: string | null;
  git: WorkProjectGitInfo;
  conversations: Conversation[];
}

export interface SidebarWorkerGroup {
  key: string;
  label: string;
  /** Dominant session agent_id in this group — used by the sidebar create pencil. */
  agentId: string | null;
  connectionStatus: WorkerConnectionStatus;
  projectGroups: SidebarWorkProjectGroup[];
  conversations: Conversation[];
}

/** Pick the most common non-empty agent_id among grouped sessions. */
export function resolveWorkerGroupAgentId(conversations: Conversation[]): string | null {
  const counts = new Map<string, number>();
  for (const conversation of conversations) {
    const id = conversation.agent_id?.trim();
    if (!id) continue;
    counts.set(id, (counts.get(id) ?? 0) + 1);
  }
  if (counts.size === 0) return null;
  return [...counts.entries()].sort((a, b) => b[1] - a[1])[0]![0];
}

export function newSessionLandingPath(params: { agent?: string; project?: string }): string {
  const q = new URLSearchParams();
  if (params.agent?.trim()) q.set("agent", params.agent.trim());
  if (params.project?.trim()) q.set("project", params.project.trim());
  const query = q.toString();
  return query ? `/?${query}` : "/";
}

/** Worker → work directory (project) → sessions. */
export function groupConversationsByWorkerAndProject(
  conversations: Conversation[],
  activeOverride: ActiveChatOverride | null,
): SidebarWorkerGroup[] {
  const sorted = sortByUpdatedAtDesc(conversations, activeOverride);
  const workerMap = new Map<string, Conversation[]>();

  for (const conversation of sorted) {
    const key = workerGroupKey(conversation);
    const bucket = workerMap.get(key) ?? [];
    bucket.push(conversation);
    workerMap.set(key, bucket);
  }

  const groups: SidebarWorkerGroup[] = [];
  for (const [workerKey, workerConversations] of workerMap) {
    const projectMap = new Map<string, { label: string; conversations: Conversation[] }>();
    for (const conversation of workerConversations) {
      const projectKey = workProjectGroupKey(conversation);
      const label = workProjectGroupLabel(conversation);
      const entry = projectMap.get(projectKey) ?? { label, conversations: [] };
      entry.conversations.push(conversation);
      if (!projectMap.has(projectKey)) projectMap.set(projectKey, entry);
    }

    const projectGroups: SidebarWorkProjectGroup[] = [];
    for (const [projectKey, { label, conversations: projectConversations }] of projectMap) {
      projectGroups.push({
        key: projectKey,
        label,
        workspacePath: workProjectWorkspacePath(projectConversations),
        git: workProjectGitInfo(projectConversations),
        conversations: projectConversations,
      });
    }

    projectGroups.sort((a, b) => {
      const aMax = Math.max(...a.conversations.map((c) => c.updated_at));
      const bMax = Math.max(...b.conversations.map((c) => c.updated_at));
      return bMax - aMax;
    });

    groups.push({
      key: workerKey,
      label: formatWorkerDisplayName(workerKey),
      agentId: resolveWorkerGroupAgentId(workerConversations),
      connectionStatus: aggregateWorkerConnectionStatus(workerConversations),
      projectGroups,
      conversations: workerConversations,
    });
  }

  groups.sort((a, b) => {
    const aMax = Math.max(...a.conversations.map((c) => c.updated_at));
    const bMax = Math.max(...b.conversations.map((c) => c.updated_at));
    return bMax - aMax;
  });

  return groups;
}

export function workProjectStorageKey(workerKey: string, projectKey: string): string {
  return `${workerKey}::${projectKey}`;
}

export function workProjectShowsHeader(
  project: SidebarWorkProjectGroup,
  projectCount: number,
): boolean {
  if (projectCount > 1) return true;
  if (project.workspacePath) return true;
  if (project.key.startsWith("project:")) return true;
  return false;
}

export function collectVisibleGroupedSessionIds(
  groups: readonly SidebarWorkerGroup[],
  options: {
    sectionCollapsed: boolean;
    expandedWorkerKeys: ReadonlySet<string>;
    expandedProjectKeys: ReadonlySet<string>;
  },
): string[] {
  if (options.sectionCollapsed) return [];
  const ids: string[] = [];
  for (const worker of groups) {
    if (!options.expandedWorkerKeys.has(worker.key)) continue;
    for (const project of worker.projectGroups) {
      const projectKey = workProjectStorageKey(worker.key, project.key);
      const showHeader = workProjectShowsHeader(project, worker.projectGroups.length);
      if (!showHeader || options.expandedProjectKeys.has(projectKey)) {
        ids.push(...project.conversations.map((c) => c.id));
      }
    }
  }
  return ids;
}
