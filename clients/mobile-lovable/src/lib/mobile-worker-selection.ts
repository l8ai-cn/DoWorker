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
  return {
    kind: "ready",
    current: agents.find((agent) => agent.id === selectedID) ?? agents[0],
    message: null,
  };
}
