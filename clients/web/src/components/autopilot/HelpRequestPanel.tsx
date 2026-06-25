"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAutopilotStore, useAutopilotThinking } from "@/stores/autopilot";
import {
  AlertTriangle,
  CheckCircle,
  XCircle,
  MessageSquare,
  Terminal,
  HelpCircle,
} from "lucide-react";

interface HelpRequestPanelProps {
  autopilotControllerKey: string;
  className?: string;
  onApprove?: (continueExecution: boolean, additionalIterations?: number) => void;
  onCustomResponse?: () => void;
}

export function HelpRequestPanel({
  autopilotControllerKey,
  className,
  onApprove,
  onCustomResponse,
}: HelpRequestPanelProps) {
  const approveAutopilotController = useAutopilotStore((s) => s.approveAutopilotController);
  const thinking = useAutopilotThinking(autopilotControllerKey);

  if (!thinking?.help_request) {
    return null;
  }

  const helpRequest = thinking.help_request;

  const handleSuggestionClick = (suggestion: { action: string; label: string }) => {
    switch (suggestion.action) {
      case "approve":
        if (onApprove) {
          onApprove(true);
        } else {
          approveAutopilotController(autopilotControllerKey, { continue_execution: true });
        }
        break;
      case "skip":
        if (onApprove) {
          onApprove(true, 5); // Add extra iterations for recovery
        } else {
          approveAutopilotController(autopilotControllerKey, {
            continue_execution: true,
            additional_iterations: 5,
          });
        }
        break;
      case "stop":
        if (onApprove) {
          onApprove(false);
        } else {
          approveAutopilotController(autopilotControllerKey, { continue_execution: false });
        }
        break;
      case "custom":
        onCustomResponse?.();
        break;
      default:
        if (onApprove) {
          onApprove(true);
        }
    }
  };

  return (
    <div className={cn("rounded-lg border border-warning/50 bg-warning-bg p-4 shadow-sm", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 mb-3">
        <div className="flex items-center justify-center w-8 h-8 rounded-full bg-warning/20">
          <AlertTriangle className="h-4 w-4 text-warning" />
        </div>
        <div>
          <h3 className="font-semibold text-sm text-warning">Help Requested</h3>
          <p className="text-xs text-muted-foreground">
            Control Agent needs human intervention
          </p>
        </div>
        <Badge variant="outline" className="ml-auto bg-warning text-white">
          Iteration #{thinking.iteration}
        </Badge>
      </div>

      {/* Reason */}
      <div className="mb-3">
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground mb-1">
          <HelpCircle className="h-3 w-3" />
          <span>Reason</span>
        </div>
        <p className="text-sm">{helpRequest.reason}</p>
      </div>

      {/* Context */}
      {helpRequest.context && (
        <div className="mb-3">
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground mb-1">
            <MessageSquare className="h-3 w-3" />
            <span>Context</span>
          </div>
          <p className="text-sm text-muted-foreground">{helpRequest.context}</p>
        </div>
      )}

      {/* Terminal Excerpt */}
      {helpRequest.terminal_excerpt && (
        <div className="mb-4">
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground mb-1">
            <Terminal className="h-3 w-3" />
            <span>Terminal Output</span>
          </div>
          <div className="rounded bg-muted p-3 overflow-x-auto">
            <pre className="text-xs text-foreground font-mono whitespace-pre-wrap break-all">
              {helpRequest.terminal_excerpt}
            </pre>
          </div>
        </div>
      )}

      {/* Suggestions */}
      {helpRequest.suggestions && helpRequest.suggestions.length > 0 && (
        <div className="border-t border-warning/20 pt-3">
          <div className="text-xs text-muted-foreground mb-2">Suggested Actions</div>
          <div className="flex flex-wrap gap-2">
            {helpRequest.suggestions.map((suggestion, index) => (
              <Button
                key={index}
                size="sm"
                variant={suggestion.action === "approve" ? "default" : "outline"}
                onClick={() => handleSuggestionClick(suggestion)}
                className={cn(
                  suggestion.action === "approve" && "bg-success hover:bg-success/90",
                  suggestion.action === "stop" && "text-danger border-danger/50 hover:bg-danger/10"
                )}
              >
                {suggestion.action === "approve" && <CheckCircle className="h-3 w-3 mr-1" />}
                {suggestion.action === "stop" && <XCircle className="h-3 w-3 mr-1" />}
                {suggestion.label}
              </Button>
            ))}
          </div>
        </div>
      )}

      {/* Default Actions (if no suggestions provided) */}
      {(!helpRequest.suggestions || helpRequest.suggestions.length === 0) && (
        <div className="border-t border-warning/20 pt-3">
          <div className="text-xs text-muted-foreground mb-2">Actions</div>
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="default"
              className="bg-success hover:bg-success/90"
              onClick={() => handleSuggestionClick({ action: "approve", label: "Continue" })}
            >
              <CheckCircle className="h-3 w-3 mr-1" />
              Approve & Continue
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleSuggestionClick({ action: "skip", label: "Skip" })}
            >
              Skip This Step
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="text-danger border-danger/50 hover:bg-danger/10"
              onClick={() => handleSuggestionClick({ action: "stop", label: "Stop" })}
            >
              <XCircle className="h-3 w-3 mr-1" />
              Stop Autopilot
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

export default HelpRequestPanel;
