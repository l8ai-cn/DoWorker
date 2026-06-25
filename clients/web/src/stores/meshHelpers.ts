import { Play, Hourglass, Pause, type LucideIcon } from "lucide-react";

export const getPodStatusInfo = (status: string) => {
  const statusMap: Record<string, { label: string; color: string; bgColor: string }> = {
    initializing: {
      label: "Initializing",
      color: "text-info",
      bgColor: "bg-info-bg",
    },
    running: {
      label: "Running",
      color: "text-success",
      bgColor: "bg-success-bg",
    },
    paused: {
      label: "Paused",
      color: "text-warning",
      bgColor: "bg-warning-bg",
    },
    terminated: {
      label: "Terminated",
      color: "text-muted-foreground",
      bgColor: "bg-muted",
    },
    failed: {
      label: "Failed",
      color: "text-danger",
      bgColor: "bg-danger-bg",
    },
  };
  return statusMap[status] || statusMap.terminated;
};

export const getAgentStatusInfo = (agentStatus: string): {
  label: string; color: string; dotColor: string; bgColor: string; icon: LucideIcon;
} => {
  const statusMap: Record<string, {
    label: string; color: string; dotColor: string; bgColor: string; icon: LucideIcon;
  }> = {
    executing: {
      label: "Executing", color: "text-success",
      dotColor: "bg-success", bgColor: "bg-success-bg", icon: Play,
    },
    waiting: {
      label: "Waiting for Input", color: "text-warning",
      dotColor: "bg-warning", bgColor: "bg-warning-bg", icon: Hourglass,
    },
    idle: {
      label: "Idle", color: "text-muted-foreground",
      dotColor: "bg-muted-foreground", bgColor: "bg-muted", icon: Pause,
    },
  };
  return statusMap[agentStatus] || statusMap.idle;
};

export const getBindingStatusInfo = (status: string) => {
  const statusMap: Record<string, { label: string; color: string }> = {
    active: { label: "Active", color: "stroke-success" },
    pending: { label: "Pending", color: "stroke-warning" },
    revoked: { label: "Revoked", color: "stroke-danger" },
    expired: { label: "Expired", color: "stroke-muted-foreground" },
  };
  return statusMap[status] || statusMap.active;
};
