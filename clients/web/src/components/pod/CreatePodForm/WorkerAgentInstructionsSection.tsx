"use client";

import { useMemo } from "react";
import { FileText } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { AgentfileCodeEditor } from "./AgentfileCodeEditor";
import type { AgentfileCompletionContext } from "@/lib/codemirror-agentfile";
import type { ConfigField, RepositoryData, EnvBundleSummary } from "@/lib/api";

interface WorkerAgentInstructionsSectionProps {
  generatedLayer: string;
  rawMode: boolean;
  rawText: string;
  onRawModeChange: (enabled: boolean) => void;
  onRawTextChange: (text: string) => void;
  configFields: ConfigField[];
  repositories: RepositoryData[];
  envBundles: EnvBundleSummary[];
  t: (key: string) => string;
}

export function WorkerAgentInstructionsSection({
  generatedLayer,
  rawMode,
  rawText,
  onRawModeChange,
  onRawTextChange,
  configFields,
  repositories,
  envBundles,
  t,
}: WorkerAgentInstructionsSectionProps) {
  const completionContext = useMemo<AgentfileCompletionContext>(
    () => ({
      configFields,
      repositories: repositories.map((r) => ({
        slug: r.slug,
        name: r.name,
        default_branch: r.default_branch,
      })),
      envBundles: envBundles.map((b) => ({ name: b.name })),
    }),
    [configFields, repositories, envBundles],
  );

  const preview = rawMode ? rawText : generatedLayer;

  return (
    <div className="space-y-3" data-testid="worker-agent-instructions">
      <div className="flex items-start justify-between gap-3">
        <div className="flex gap-2">
          <FileText className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
          <div>
            <p className="text-sm font-medium">{t("ide.createPod.agentInstructionsTitle")}</p>
            <p className="text-xs leading-5 text-muted-foreground">
              {t("ide.createPod.agentInstructionsDescription")}
            </p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Label htmlFor="agentfile-custom-mode" className="text-xs text-muted-foreground">
            {t("ide.createPod.agentInstructionsCustom")}
          </Label>
          <Switch
            id="agentfile-custom-mode"
            checked={rawMode}
            onCheckedChange={onRawModeChange}
          />
        </div>
      </div>

      {rawMode ? (
        <div className="min-h-[160px] overflow-hidden rounded-md border border-border">
          <AgentfileCodeEditor
            value={rawText}
            onChange={onRawTextChange}
            completionContext={completionContext}
          />
        </div>
      ) : (
        <pre className="min-h-[120px] overflow-x-auto whitespace-pre-wrap rounded-md border border-border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
          {preview.trim() ? preview : t("ide.createPod.agentInstructionsEmptyPreview")}
        </pre>
      )}

      <p className="text-xs text-muted-foreground">{t("ide.createPod.agentInstructionsFooter")}</p>
    </div>
  );
}
