import type { ComponentType } from "react";

import type { AgentArtifactItem } from "../agentArtifactContracts";
import type { AgentSessionRuntime } from "../contracts";
import type { ContentRendererRegistration } from "../registry/ContentRendererRegistry";

export interface AgentContentRendererProps {
  filename: string;
  item: AgentArtifactItem;
  presentation?: "developer" | "user";
  runtime: AgentSessionRuntime;
  sessionId: string;
}

export type AgentContentRendererComponent =
  ComponentType<AgentContentRendererProps>;

export type AgentContentRendererRegistration = ContentRendererRegistration<
  AgentContentRendererComponent,
  AgentContentRendererComponent,
  AgentContentRendererComponent
>;
