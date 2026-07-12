export type EventType =
  | "user_message"
  | "agent_message"
  | "agent_thought"
  | "tool_call"
  | "plan"
  | "permission_request"
  | "ask_user"
  | "phase"
  | "error";

export interface AskUserField {
  name: string;
  label: string;
  type: "text" | "textarea" | "select" | "radio" | "checkbox" | "number";
  options?: string[];
  placeholder?: string;
  required?: boolean;
  defaultValue?: string;
}

export interface AskUserForm {
  title: string;
  description?: string;
  fields: AskUserField[];
  submitLabel?: string;
}

export type ToolKind = "read" | "write" | "edit" | "shell" | "search" | "fetch" | "other";

export interface DiffHunk {
  header?: string;
  lines: { kind: "add" | "del" | "ctx" | "hunk"; text: string }[];
}

export interface PlanItem {
  text: string;
  status: "pending" | "in_progress" | "completed";
}

export interface AgentEvent {
  id: string;
  type: EventType;
  ts: string;
  title: string;
  detail?: string;
  tool?: string;
  toolKind?: ToolKind;
  status?: "pending" | "in_progress" | "completed" | "failed";
  duration?: string;
  command?: string;
  cwd?: string;
  output?: string;
  exitCode?: number;
  filePath?: string;
  additions?: number;
  deletions?: number;
  diff?: DiffHunk[];
  plan?: PlanItem[];
  query?: string;
  results?: { title: string; url?: string; snippet?: string }[];
  markdown?: string;
  images?: { src: string; caption?: string; alt?: string }[];
  attachments?: { name: string; kind: "image" | "file"; src?: string; note?: string }[];
  form?: AskUserForm;
  answer?: Record<string, string | boolean>;
  elicitationId?: string;
  phaseIndex?: number;
  phaseTotal?: number;
  phaseEmoji?: string;
  phaseSummary?: string;
}

export type SessionStatus = "running" | "waiting_approval" | "completed" | "failed" | "idle";

export interface SessionMetrics {
  tokensIn: number;
  tokensOut: number;
  toolCalls: number;
  filesChanged: number;
  elapsed: string;
  cost?: string;
}

export interface AgentSession {
  id: string;
  interactionMode?: "acp" | "pty" | null;
  projectId: string;
  title: string;
  agent: string;
  branch: string;
  status: SessionStatus;
  updatedAt: string;
  eventCount: number;
  preview: string;
  metrics?: SessionMetrics;
  events: AgentEvent[];
}

export interface Project {
  id: string;
  name: string;
  repo: string;
  host: string;
  color: string;
  online: boolean;
  sessionIds: string[];
}

export const statusMeta: Record<
  SessionStatus,
  { label: string; dotClass: string; textClass: string; ring: string }
> = {
  running: { label: "运行中", dotClass: "bg-primary pulse-dot", textClass: "text-primary", ring: "ring-primary/30" },
  waiting_approval: { label: "待审批", dotClass: "bg-warning pulse-dot", textClass: "text-warning", ring: "ring-warning/30" },
  completed: { label: "已完成", dotClass: "bg-success", textClass: "text-success", ring: "ring-success/20" },
  failed: { label: "失败", dotClass: "bg-destructive", textClass: "text-destructive", ring: "ring-destructive/25" },
  idle: { label: "空闲", dotClass: "bg-muted-foreground/60", textClass: "text-muted-foreground", ring: "ring-border" },
};
