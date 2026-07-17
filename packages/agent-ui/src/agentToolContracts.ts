import type { ToolRendererKey } from "./registry/rendererKeys";

export type AgentToolStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed";

export type AgentToolResult =
  | {
      id: string;
      kind: "text";
      text: string;
    }
  | {
      id: string;
      kind: "data";
      value: unknown;
    }
  | {
      artifactId: string;
      id: string;
      kind: "artifact";
      mediaType: string | null;
      representationId: string | null;
      revision: bigint;
      role: string;
      schemaVersion: string;
    };

export interface AgentToolActivityItem {
  id: string;
  kind: "tool";
  identity: ToolRendererKey;
  title: string;
  detail?: string;
  input?: string;
  inputValue?: unknown;
  output?: string;
  results: readonly AgentToolResult[];
  status: AgentToolStatus;
}
