import type { AgentPickerOption } from "./agent-display";

export type MobileWorkerSelection =
  | { kind: "ready"; current: AgentPickerOption; message: null }
  | { kind: "unauthenticated" | "loading" | "error" | "empty"; current: null; message: string };

export function resolveMobileWorkerSelection(
  agents: AgentPickerOption[],
  selectedID: string,
  authenticated: boolean,
  loading: boolean,
  error: string | null,
): MobileWorkerSelection {
  if (!authenticated) {
    return { kind: "unauthenticated", current: null, message: "登录后加载可用 Worker。" };
  }
  if (loading) {
    return { kind: "loading", current: null, message: "正在加载可用 Worker…" };
  }
  if (error) {
    return { kind: "error", current: null, message: `无法加载可用 Worker：${error}` };
  }
  if (agents.length === 0) {
    return {
      kind: "empty",
      current: null,
      message: "当前组织没有可用 Worker，请先在工作区配置并启用 Worker。",
    };
  }
  const compatibleAgents = agents.filter((agent) => agent.supportedModes.length > 0);
  if (compatibleAgents.length === 0) {
    return {
      kind: "error",
      current: null,
      message: "可用 Worker 未声明支持的交互模式。",
    };
  }
  return {
    kind: "ready",
    current: compatibleAgents.find((agent) => agent.id === selectedID) ?? compatibleAgents[0],
    message: null,
  };
}
