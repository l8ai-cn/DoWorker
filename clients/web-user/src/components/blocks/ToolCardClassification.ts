import type { RenderItem } from "@/lib/renderItems";

const ADVISE_MODELS_NAMES = new Set(["sys_advise_models", "mcp__omnigent__sys_advise_models"]);
const SESSION_SEND_NAMES = new Set(["sys_session_send", "mcp__omnigent__sys_session_send"]);

export function isSmartRoutingTool(item: RenderItem): boolean {
  return item.kind === "tool" && ADVISE_MODELS_NAMES.has(item.execution.name);
}

export function isPersistentToolCard(item: RenderItem): boolean {
  return (
    item.kind === "tool" &&
    (ADVISE_MODELS_NAMES.has(item.execution.name) || SESSION_SEND_NAMES.has(item.execution.name))
  );
}
