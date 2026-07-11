"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  X,
  Loader2,
  AlertCircle,
  CheckCircle2,
  RefreshCw,
  Clock,
} from "lucide-react";

interface InitProgress {
  progress: number;
  phase: string;
  message: string;
}

interface PaneLoadingStateProps {
  podStatus: string;
  initProgress?: InitProgress;
  onClose?: () => void;
}

const SLOW_INIT_THRESHOLD_SEC = 120;

function useElapsedSeconds() {
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setElapsed((prev) => prev + 1);
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  return elapsed;
}

function formatElapsed(totalSeconds: number): string {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes === 0) return `${seconds}s`;
  return `${minutes}m ${seconds.toString().padStart(2, "0")}s`;
}

export function PaneLoadingState({
  podStatus,
  initProgress,
  onClose,
}: PaneLoadingStateProps) {
  const isCompleted = podStatus === "completed";
  const elapsed = useElapsedSeconds();
  const isSlowInit = !isCompleted && elapsed >= SLOW_INIT_THRESHOLD_SEC;

  return (
    <div className="flex-1 flex items-center justify-center bg-terminal-bg">
      <div className="text-center p-4 max-w-sm">
        {isCompleted ? (
          <CheckCircle2 className="w-12 h-12 text-success mx-auto mb-3" />
        ) : (
          <Loader2 className="w-12 h-12 text-primary animate-spin mx-auto mb-3" />
        )}
        <p className="text-terminal-text font-medium mb-1">
          {isCompleted
            ? "Pod completed"
            : initProgress?.message || "Waiting for Pod to be ready..."}
        </p>
        {initProgress ? (
          <div className="mt-3 space-y-2">
            <Progress value={initProgress.progress} className="h-2" />
            <p className="text-xs text-terminal-text-muted">
              {initProgress.phase} - {initProgress.progress}%
            </p>
          </div>
        ) : (
          <p className="text-sm text-terminal-text-muted">
            {isCompleted ? (
              <>Status: <span className="text-success">{podStatus}</span></>
            ) : (
              <>Status: <span className="text-warning">{podStatus}</span></>
            )}
          </p>
        )}
        {!isCompleted && (
          <p className="text-xs text-terminal-text-muted mt-2 flex items-center justify-center gap-1">
            <Clock className="w-3 h-3" />
            {formatElapsed(elapsed)}
          </p>
        )}
        {isSlowInit && (
          <div className="mt-3 p-2 rounded bg-warning-bg border border-warning/30">
            <p className="text-xs text-warning">
              Taking longer than expected. The runner may be cloning a large repository or experiencing connectivity issues.
            </p>
          </div>
        )}
        {(podStatus === "unknown" || isCompleted || isSlowInit) && onClose && (
          <Button
            variant="outline"
            size="sm"
            className="mt-4 text-danger border-danger/50 hover:bg-danger/10"
            onClick={onClose}
          >
            <X className="w-4 h-4 mr-2" />
            Close
          </Button>
        )}
      </div>
    </div>
  );
}

interface PaneReconnectingStateProps {
  onClose?: () => void;
}

export function PaneReconnectingState({
  onClose,
}: PaneReconnectingStateProps) {
  const t = useTranslations("workspace.reconnecting");
  return (
    <div className="flex-1 flex items-center justify-center bg-terminal-bg">
      <div className="text-center p-4 max-w-sm">
        <RefreshCw className="w-12 h-12 text-warning mx-auto mb-3 animate-spin" />
        <p className="text-terminal-text font-medium mb-1">
          {t("title")}
        </p>
        <p className="text-sm text-terminal-text-muted mb-4">
          {t("description")}
        </p>
        {onClose && (
          <Button
            variant="outline"
            size="sm"
            className="text-muted-foreground border-border hover:bg-muted"
            onClick={onClose}
          >
            <X className="w-4 h-4 mr-2" />
            Close
          </Button>
        )}
      </div>
    </div>
  );
}

interface PaneErrorStateProps {
  error: string;
  onClose?: () => void;
}

export function PaneErrorState({
  error,
  onClose,
}: PaneErrorStateProps) {
  return (
    <div className="flex-1 flex items-center justify-center bg-terminal-bg">
      <div className="text-center p-4">
        <AlertCircle className="w-12 h-12 text-danger mx-auto mb-3" />
        <p className="text-terminal-text font-medium mb-1">{error}</p>
        <p className="text-sm text-terminal-text-muted mb-4">
          The pod cannot be connected. Please check the pod status or create a new one.
        </p>
        {onClose && (
          <Button
            variant="outline"
            size="sm"
            className="text-danger border-danger/50 hover:bg-danger/10"
            onClick={onClose}
          >
            <X className="w-4 h-4 mr-2" />
            Close
          </Button>
        )}
      </div>
    </div>
  );
}
