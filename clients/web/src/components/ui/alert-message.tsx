"use client";

import { X } from "lucide-react";
import { cn } from "@/lib/utils";

type AlertType = "error" | "success" | "warning" | "info";

interface AlertMessageProps {
  type: AlertType;
  message: string;
  onDismiss?: () => void;
  className?: string;
}

const alertStyles: Record<AlertType, { container: string; text: string }> = {
  error: {
    container: "bg-danger-bg border-danger/30",
    text: "text-danger",
  },
  success: {
    container: "bg-success-bg border-success/30",
    text: "text-success",
  },
  warning: {
    container: "bg-warning-bg border-warning/30",
    text: "text-warning",
  },
  info: {
    container: "bg-info-bg border-info/30",
    text: "text-info",
  },
};

export function AlertMessage({ type, message, onDismiss, className }: AlertMessageProps) {
  const styles = alertStyles[type];

  return (
    <div
      role="alert"
      aria-live={type === "error" ? "assertive" : "polite"}
      className={cn(
        "flex items-start justify-between gap-2 p-3 rounded-md border text-sm",
        styles.container,
        className
      )}
    >
      <p className={styles.text}>{message}</p>
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          className={cn(
            "flex-shrink-0 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors",
            styles.text
          )}
          aria-label="Dismiss"
        >
          <X className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}

export default AlertMessage;
