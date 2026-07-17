import type {
  AgentConnectionStatus,
  AgentSessionStatus,
} from "./contracts";

type ActivityStatus = "pending" | "running" | "completed" | "failed";

export function englishSessionStatus(
  status: AgentSessionStatus,
  connection: AgentConnectionStatus,
) {
  if (connection === "reconnecting") return "Reconnecting";
  if (connection === "disconnected") return "Offline";
  if (status === "running" || status === "waiting") return "Working";
  if (status === "launching") return "Starting";
  if (status === "failed") return "Failed";
  if (status === "completed") return "Completed";
  return "Ready";
}

export function chineseSessionStatus(
  status: AgentSessionStatus,
  connection: AgentConnectionStatus,
) {
  if (connection === "reconnecting") return "正在重连";
  if (connection === "disconnected") return "已离线";
  if (status === "running" || status === "waiting") return "执行中";
  if (status === "launching") return "正在启动";
  if (status === "failed") return "执行失败";
  if (status === "completed") return "已完成";
  return "就绪";
}

export function englishActivityStatus(status: ActivityStatus) {
  return {
    pending: "Pending",
    running: "Running",
    completed: "Completed",
    failed: "Failed",
  }[status];
}

export function chineseActivityStatus(status: ActivityStatus) {
  return {
    pending: "等待中",
    running: "执行中",
    completed: "已完成",
    failed: "失败",
  }[status];
}
