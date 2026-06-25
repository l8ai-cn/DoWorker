import * as React from "react";
import {
  ArrowRight,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Clock,
  Eye,
  Send,
  MessageSquare,
  Play,
  Loader2,
} from "lucide-react";
import type { AutopilotController } from "@/stores/autopilot";

export interface AutopilotFloatingPanelProps {
  autopilotController: AutopilotController;
  className?: string;
  onClose?: () => void;
}

export type NormalizedDecisionType = "continue" | "completed" | "need_help" | "give_up";

export const decisionConfig: Record<
  NormalizedDecisionType,
  { label: string; bgColor: string; textColor: string; icon: React.ReactNode }
> = {
  continue: {
    label: "Continue",
    bgColor: "bg-info",
    textColor: "text-info",
    icon: <ArrowRight className="h-3 w-3" />,
  },
  completed: {
    label: "Completed",
    bgColor: "bg-success",
    textColor: "text-success",
    icon: <CheckCircle className="h-3 w-3" />,
  },
  need_help: {
    label: "Need Help",
    bgColor: "bg-warning",
    textColor: "text-warning",
    icon: <AlertTriangle className="h-3 w-3" />,
  },
  give_up: {
    label: "Give Up",
    bgColor: "bg-danger",
    textColor: "text-danger",
    icon: <XCircle className="h-3 w-3" />,
  },
};

export const actionConfig: Record<string, { label: string; icon: React.ReactNode }> = {
  observe: { label: "Observing", icon: <Eye className="h-3 w-3" /> },
  send_input: { label: "Sending Input", icon: <Send className="h-3 w-3" /> },
  wait: { label: "Waiting", icon: <Clock className="h-3 w-3" /> },
  none: { label: "No Action", icon: <MessageSquare className="h-3 w-3" /> },
};

export const iterationPhaseConfig: Record<
  string,
  { label: string; color: string; icon: React.ReactNode }
> = {
  prompt: {
    label: "Initial",
    color: "bg-info",
    icon: <Send className="h-3 w-3" />,
  },
  started: {
    label: "Started",
    color: "bg-info",
    icon: <Play className="h-3 w-3" />,
  },
  control_running: {
    label: "Running",
    color: "bg-warning",
    icon: <Loader2 className="h-3 w-3 animate-spin" />,
  },
  action_sent: {
    label: "Sent",
    color: "bg-success",
    icon: <Send className="h-3 w-3" />,
  },
  completed: {
    label: "Done",
    color: "bg-success",
    icon: <CheckCircle className="h-3 w-3" />,
  },
  error: {
    label: "Error",
    color: "bg-danger",
    icon: <XCircle className="h-3 w-3" />,
  },
};

export function normalizeDecisionType(backendType: string): NormalizedDecisionType {
  const mapping: Record<string, NormalizedDecisionType> = {
    "CONTINUE": "continue",
    "TASK_COMPLETED": "completed",
    "NEED_HUMAN_HELP": "need_help",
    "GIVE_UP": "give_up",
    "continue": "continue",
    "completed": "completed",
    "need_help": "need_help",
    "give_up": "give_up",
  };
  return mapping[backendType] || "continue";
}
