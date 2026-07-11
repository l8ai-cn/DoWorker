"use client";

import { AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

interface WorkflowDetailErrorStateProps {
  error: string;
  retryLabel: string;
  onRetry: () => void;
}

export function WorkflowDetailErrorState({ error, retryLabel, onRetry }: WorkflowDetailErrorStateProps) {
  return (
    <div className="flex h-full flex-col items-center justify-center py-20 text-center">
      <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-md bg-destructive/10">
        <AlertCircle className="h-6 w-6 text-destructive" />
      </div>
      <p className="mb-3 text-sm text-muted-foreground">{error}</p>
      <Button variant="outline" size="sm" className="gap-1.5" onClick={onRetry}>
        <RefreshCw className="h-3.5 w-3.5" />
        {retryLabel}
      </Button>
    </div>
  );
}
