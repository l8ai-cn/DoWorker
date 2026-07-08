import { ChevronRightIcon, GitBranchIcon, SquarePenIcon } from "lucide-react";
import { type MouseEvent, type ReactNode, Fragment } from "react";
import { Link } from "react-router-dom";
import { SessionStateBadge } from "@/components/SessionStateBadge";
import { Button } from "@/components/ui/button";
import type { Conversation } from "@/hooks/useConversations";
import type { SessionState } from "@/hooks/useSessionState";
import { isConversationUnseen } from "@/hooks/useUnseenConversations";
import { cn } from "@/lib/utils";
import {
  type SidebarWorkerGroup,
  type WorkerConnectionStatus,
  type WorkProjectGitInfo,
  newSessionLandingPath,
  workProjectShowsHeader,
  workProjectStorageKey,
} from "./sidebarNav";

export function groupMarkerState(conversations: Conversation[]): SessionState | null {
  let awaiting = 0;
  let unseen = false;
  let running = false;
  for (const conversation of conversations) {
    const pending = conversation.pending_elicitations_count ?? 0;
    if (pending > 0) awaiting += pending;
    else if (isConversationUnseen(conversation.id, conversation.updated_at, conversation.status)) {
      unseen = true;
    } else if (conversation.status === "running") running = true;
  }
  if (awaiting > 0) return { kind: "awaiting", count: awaiting };
  if (unseen) return { kind: "unseen" };
  if (running) return { kind: "running" };
  return null;
}

function WorkerConnectionDot({ status }: { status: WorkerConnectionStatus }) {
  const tone =
    status === "online"
      ? "bg-status-green"
      : status === "offline"
        ? "bg-muted-foreground/50"
        : status === "mixed"
          ? "bg-status-yellow"
          : "bg-muted-foreground/30";
  const label =
    status === "online"
      ? "Worker connected"
      : status === "offline"
        ? "Worker offline"
        : status === "mixed"
          ? "Mixed worker connection"
          : "Worker connection unknown";
  return (
    <span
      aria-label={label}
      title={label}
      data-testid="worker-connection-dot"
      data-status={status}
      className={cn("size-1.5 shrink-0 rounded-full", tone)}
    />
  );
}

function GitRepoHint({
  git,
  workspacePath,
}: {
  git: WorkProjectGitInfo;
  workspacePath: string | null;
}) {
  const pathTitle = workspacePath ?? undefined;
  if (git.linked && git.branchLabel) {
    return (
      <span
        className="flex min-w-0 items-center gap-1 text-[10px] text-status-green"
        title={pathTitle}
        data-testid="git-repo-linked"
      >
        <GitBranchIcon className="size-3 shrink-0" aria-hidden />
        <span className="truncate">{git.branchLabel}</span>
      </span>
    );
  }
  if (git.hasWorkspace) {
    return (
      <span
        className="text-[10px] text-muted-foreground"
        title={pathTitle}
        data-testid="git-repo-none"
      >
        No Git repo
      </span>
    );
  }
  return null;
}

function NestedSectionHeader({
  title,
  subtitle,
  marker,
  collapsed,
  onToggleCollapsed,
  leading,
  indent = 0,
  hasAction = false,
}: {
  title: string;
  subtitle?: ReactNode;
  marker?: SessionState | null;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  leading?: ReactNode;
  indent?: number;
  hasAction?: boolean;
}) {
  return (
    <h2>
      <button
        type="button"
        aria-expanded={!collapsed}
        onClick={onToggleCollapsed}
        style={indent > 0 ? { paddingLeft: `${indent * 0.75 + 0.5}rem` } : undefined}
        className={cn(
          "group flex w-full items-center gap-1 rounded-md px-2 py-1 text-left text-sm text-muted-foreground transition-colors hover:text-foreground",
          hasAction && "md:group-hover/worker-header:pr-8",
        )}
      >
        {leading}
        <span className="flex min-w-0 flex-1 flex-col gap-0.5">
          <span className="truncate">{title}</span>
          {!collapsed && subtitle}
        </span>
        <ChevronRightIcon
          className={cn(
            "size-3.5 shrink-0 transition-[transform,opacity]",
            !collapsed && "rotate-90",
            "md:opacity-0 md:group-hover:opacity-100 md:group-focus-visible:opacity-100",
          )}
        />
        {collapsed && marker && (
          <span className="ml-auto flex shrink-0 items-center">
            <SessionStateBadge state={marker} />
          </span>
        )}
      </button>
    </h2>
  );
}

export interface SidebarProjectWorkerGroupsProps {
  title: string;
  conversations: Conversation[];
  sectionCollapsed: boolean;
  onToggleSectionCollapsed: () => void;
  groups: SidebarWorkerGroup[];
  expandedWorkers: ReadonlySet<string>;
  expandedProjects: ReadonlySet<string>;
  onToggleWorker: (key: string) => void;
  onToggleProject: (key: string) => void;
  renderRow: (conversation: Conversation) => ReactNode;
  onNavigate?: (e: MouseEvent<HTMLAnchorElement>) => void;
  footer?: ReactNode;
}

export function SidebarProjectWorkerGroups({
  title,
  conversations,
  sectionCollapsed,
  onToggleSectionCollapsed,
  groups,
  expandedWorkers,
  expandedProjects,
  onToggleWorker,
  onToggleProject,
  renderRow,
  onNavigate,
  footer,
}: SidebarProjectWorkerGroupsProps) {
  return (
    <section className="group/section relative">
      <NestedSectionHeader
        title={title}
        collapsed={sectionCollapsed}
        onToggleCollapsed={onToggleSectionCollapsed}
        marker={sectionCollapsed ? groupMarkerState(conversations) : null}
      />
      {!sectionCollapsed && (
        <>
          <div className="flex flex-col gap-0.5">
            {groups.map((worker) => (
              <WorkerGroup
                key={worker.key}
                group={worker}
                expanded={expandedWorkers.has(worker.key)}
                expandedProjects={expandedProjects}
                onToggleWorker={() => onToggleWorker(worker.key)}
                onToggleProject={onToggleProject}
                renderRow={renderRow}
                onNavigate={onNavigate}
              />
            ))}
          </div>
          {footer}
        </>
      )}
    </section>
  );
}

function WorkerGroup({
  group,
  expanded,
  expandedProjects,
  onToggleWorker,
  onToggleProject,
  renderRow,
  onNavigate,
}: {
  group: SidebarWorkerGroup;
  expanded: boolean;
  expandedProjects: ReadonlySet<string>;
  onToggleWorker: () => void;
  onToggleProject: (key: string) => void;
  renderRow: (conversation: Conversation) => ReactNode;
  onNavigate?: (e: MouseEvent<HTMLAnchorElement>) => void;
}) {
  return (
    <section data-testid={`sidebar-worker-group-${group.key}`}>
      <div className="group/worker-header relative">
        <NestedSectionHeader
          title={group.label}
          collapsed={!expanded}
          onToggleCollapsed={onToggleWorker}
          marker={!expanded ? groupMarkerState(group.conversations) : null}
          leading={<WorkerConnectionDot status={group.connectionStatus} />}
          indent={1}
          hasAction={group.agentId != null}
        />
        {group.agentId && (
          <div className="absolute top-0.5 right-1 flex items-center transition-opacity md:opacity-0 md:group-focus-within/worker-header:opacity-100 md:group-hover/worker-header:opacity-100">
            <WorkerGroupNewSessionButton
              agentId={group.agentId}
              workerLabel={group.label}
              onNavigate={onNavigate}
            />
          </div>
        )}
      </div>
      {expanded && (
        <div className="flex flex-col gap-0.5">
          {group.projectGroups.map((project) => {
            const projectKey = workProjectStorageKey(group.key, project.key);
            const showHeader = workProjectShowsHeader(project, group.projectGroups.length);
            const projectExpanded = !showHeader || expandedProjects.has(projectKey);
            if (!showHeader) {
              return (
                <ul key={projectKey} className="flex flex-col gap-0.5 pl-4">
                  {project.conversations.map((conversation) => (
                    <Fragment key={conversation.id}>{renderRow(conversation)}</Fragment>
                  ))}
                </ul>
              );
            }
            return (
              <section key={projectKey} data-testid={`sidebar-project-group-${projectKey}`}>
                <NestedSectionHeader
                  title={project.label}
                  subtitle={
                    project.git.linked || project.git.hasWorkspace ? (
                      <GitRepoHint git={project.git} workspacePath={project.workspacePath} />
                    ) : null
                  }
                  collapsed={!projectExpanded}
                  onToggleCollapsed={() => onToggleProject(projectKey)}
                  marker={!projectExpanded ? groupMarkerState(project.conversations) : null}
                  indent={2}
                />
                {projectExpanded && (
                  <ul className="flex flex-col gap-0.5 pl-6">
                    {project.conversations.map((conversation) => (
                      <Fragment key={conversation.id}>{renderRow(conversation)}</Fragment>
                    ))}
                  </ul>
                )}
              </section>
            );
          })}
        </div>
      )}
    </section>
  );
}

function WorkerGroupNewSessionButton({
  agentId,
  workerLabel,
  onNavigate,
}: {
  agentId: string;
  workerLabel: string;
  onNavigate?: (e: MouseEvent<HTMLAnchorElement>) => void;
}) {
  return (
    <Button
      asChild
      type="button"
      variant="ghost"
      size="icon-sm"
      aria-label={`New session with ${workerLabel}`}
      data-testid="worker-new-session"
    >
      <Link
        to={newSessionLandingPath({ agent: agentId })}
        onClick={(e) => {
          e.stopPropagation();
          onNavigate?.(e);
        }}
      >
        <SquarePenIcon className="size-3.5" />
      </Link>
    </Button>
  );
}
