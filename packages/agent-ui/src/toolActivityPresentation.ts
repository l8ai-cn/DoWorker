import { Wrench, type LucideIcon } from "lucide-react";

import type { AgentToolActivityItem } from "./agentToolContracts";
import type { AgentToolRendererRegistration } from "./react/rendererTypes";

export interface ResolvedToolActivityPresentation {
  icon: LucideIcon;
  inputLabel: string;
  label: string;
  outputLabel: string;
  specialized: boolean;
}

export function resolveToolActivityPresentation(
  item: AgentToolActivityItem,
  renderer?: AgentToolRendererRegistration,
): ResolvedToolActivityPresentation {
  const presentation = renderer?.presentation;
  return {
    icon: presentation?.icon ?? Wrench,
    inputLabel: presentation?.inputLabel ?? "Input",
    label: presentation?.label ?? item.identity.semanticKey,
    outputLabel: presentation?.outputLabel ?? "Output",
    specialized: renderer !== undefined,
  };
}

export function toolActivityIdentity(item: AgentToolActivityItem) {
  const { namespace, schemaVersion, semanticKey } = item.identity;
  return `${namespace}/${semanticKey}@${schemaVersion}`;
}
