import {
  CheckCircle2, XCircle, Clock, Ban, SkipForward, Loader2, AlertTriangle,
} from "lucide-react";

interface StatusConfig {
  icon: React.ElementType;
  color: string;
  bg: string;
  labelKey: string;
}

export const STATUS_CONFIG: Record<string, StatusConfig> = {
  completed: { icon: CheckCircle2, color: "text-success", bg: "bg-success-bg", labelKey: "loops.statusCompleted" },
  failed: { icon: XCircle, color: "text-danger", bg: "bg-danger-bg", labelKey: "loops.statusFailed" },
  timeout: { icon: AlertTriangle, color: "text-warning", bg: "bg-warning-bg", labelKey: "loops.statusTimeout" },
  cancelled: { icon: Ban, color: "text-muted-foreground", bg: "bg-muted", labelKey: "loops.statusCancelled" },
  skipped: { icon: SkipForward, color: "text-muted-foreground", bg: "bg-muted", labelKey: "loops.statusSkipped" },
  running: { icon: Loader2, color: "text-info", bg: "bg-info-bg", labelKey: "loops.statusRunning" },
  pending: { icon: Clock, color: "text-warning", bg: "bg-warning-bg", labelKey: "loops.statusPending" },
};
