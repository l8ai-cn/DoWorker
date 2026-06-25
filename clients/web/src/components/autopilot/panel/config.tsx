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
  Pause,
  Square,
  Hand,
  Loader2,
} from "lucide-react";
import type {
  NormalizedDecisionType,
  DecisionTypeConfig,
  ActionTypeConfig,
  IterationPhaseConfig,
} from "./types";
import type { AutopilotController } from "@/stores/autopilot";

export const decisionConfig: Record<NormalizedDecisionType, DecisionTypeConfig> = {
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

export const actionConfig: Record<string, ActionTypeConfig> = {
  observe: { label: "Observing", icon: <Eye className="h-3 w-3" /> },
  send_input: { label: "Sending Input", icon: <Send className="h-3 w-3" /> },
  wait: { label: "Waiting", icon: <Clock className="h-3 w-3" /> },
  none: { label: "No Action", icon: <MessageSquare className="h-3 w-3" /> },
};

export const iterationPhaseConfig: Record<string, IterationPhaseConfig> = {
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

export const phaseConfig: Record<
  AutopilotController["phase"],
  { label: string; color: string; icon: React.ReactNode }
> = {
  initializing: {
    label: "Initializing",
    color: "text-info",
    icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
  },
  running: {
    label: "Running",
    color: "text-success",
    icon: <Play className="h-3.5 w-3.5" />,
  },
  paused: {
    label: "Paused",
    color: "text-warning",
    icon: <Pause className="h-3.5 w-3.5" />,
  },
  user_takeover: {
    label: "User Control",
    color: "text-primary",
    icon: <Hand className="h-3.5 w-3.5" />,
  },
  waiting_approval: {
    label: "Waiting Approval",
    color: "text-warning",
    icon: <AlertTriangle className="h-3.5 w-3.5" />,
  },
  completed: {
    label: "Completed",
    color: "text-success",
    icon: <CheckCircle className="h-3.5 w-3.5" />,
  },
  failed: {
    label: "Failed",
    color: "text-danger",
    icon: <XCircle className="h-3.5 w-3.5" />,
  },
  stopped: {
    label: "Stopped",
    color: "text-muted-foreground",
    icon: <Square className="h-3.5 w-3.5" />,
  },
  max_iterations: {
    label: "Max Iterations",
    color: "text-warning",
    icon: <Clock className="h-3.5 w-3.5" />,
  },
};
