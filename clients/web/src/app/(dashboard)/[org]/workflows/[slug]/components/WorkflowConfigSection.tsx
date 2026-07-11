import Link from "next/link";
import { Bot, Layers, Shield, Timer, Hash, Zap, ExternalLink } from "lucide-react";
import type { WorkflowData } from "@/stores/workflow";
import { ConfigRow } from "./ConfigRow";

interface WorkflowConfigSectionProps {
  workflow: WorkflowData;
  orgSlug: string;
  t: (key: string) => string;
}

export function WorkflowConfigSection({ workflow, orgSlug, t }: WorkflowConfigSectionProps) {
  return (
    <div className="grid grid-cols-1 xl:grid-cols-[1fr_1fr] gap-3 mb-8">
      <ConfigPanel workflow={workflow} t={t} />
      <ApiTriggerPanel workflow={workflow} orgSlug={orgSlug} t={t} />
    </div>
  );
}

function ConfigPanel({ workflow, t }: { workflow: WorkflowData; t: (key: string) => string }) {
  const concurrencyLabel =
    workflow.concurrency_policy === "skip"
      ? t("workflows.policySkip")
      : workflow.concurrency_policy === "queue"
        ? t("workflows.policyQueue")
        : t("workflows.policyReplace");

  return (
    <section data-testid="workflow-config-panel">
      <h2 className="text-sm font-semibold mb-3">{t("workflows.configuration")}</h2>
      <div className="surface-card overflow-hidden h-[calc(100%-2rem)]">
        <div className="p-4 space-y-3">
          <ConfigRow icon={<Bot className="w-3.5 h-3.5" />} label={t("workflows.mode")}
            value={workflow.execution_mode === "autopilot" ? t("workflows.modeAutopilot") : t("workflows.modeDirect")} />
          <ConfigRow icon={<Layers className="w-3.5 h-3.5" />} label={t("workflows.sandbox")}
            value={workflow.sandbox_strategy === "persistent" ? t("workflows.sandboxPersistent") : t("workflows.sandboxFresh")} />
          <ConfigRow icon={<Shield className="w-3.5 h-3.5" />} label={t("workflows.concurrency")} value={concurrencyLabel} />
          <ConfigRow icon={<Timer className="w-3.5 h-3.5" />} label={t("workflows.timeout")}
            value={`${workflow.timeout_minutes ?? 0} ${t("workflows.minutes")}`} />
          <ConfigRow icon={<Shield className="w-3.5 h-3.5" />} label={t("workflows.sessionLabel")}
            value={workflow.session_persistence ? t("workflows.sessionKeep") : t("workflows.sessionFresh")} />
          <ConfigRow icon={<Hash className="w-3.5 h-3.5" />} label={t("workflows.maxConcurrent")}
            value={String(workflow.max_concurrent_runs ?? 0)} valueTestId="workflow-max-concurrent" />
          {(workflow.max_retained_runs ?? 0) > 0 && (
            <ConfigRow icon={<Hash className="w-3.5 h-3.5" />} label={t("workflows.maxRetainedRuns")}
              value={String(workflow.max_retained_runs ?? 0)} />
          )}
          <ConfigRow icon={<Timer className="w-3.5 h-3.5" />} label={t("workflows.triggerLabel")}
            value={workflow.cron_expression ? (
              <span className="px-1.5 py-0.5 rounded bg-warning-bg text-warning text-[10px] font-medium font-mono">
                {workflow.cron_expression}
              </span>
            ) : (
              <span className="text-muted-foreground">{t("workflows.onDemand")}</span>
            )} />
        </div>
        {workflow.callback_url && (
          <div className="px-4 pb-3">
            <ConfigRow icon={<Zap className="w-3.5 h-3.5" />} label={t("workflows.webhookUrl")}
              value={<span className="text-xs font-mono truncate max-w-[140px] sm:max-w-[200px] md:max-w-[300px] inline-block align-bottom">{workflow.callback_url}</span>} />
          </div>
        )}
        <div className="panel-lift bg-surface-muted/40 p-4 pt-5">
          <div className="text-xs font-medium text-muted-foreground mb-2">{t("workflows.prompt")}</div>
          <pre className="p-3 bg-muted/50 rounded-lg text-sm whitespace-pre-wrap font-mono leading-relaxed max-h-32 overflow-y-auto text-foreground/80">
            {workflow.prompt_template}
          </pre>
        </div>
      </div>
    </section>
  );
}

function ApiTriggerPanel({ workflow, orgSlug, t }: { workflow: WorkflowData; orgSlug: string; t: (key: string) => string }) {
  return (
    <section>
      <h2 className="text-sm font-semibold mb-3">{t("workflows.apiTrigger")}</h2>
      <div className="surface-card p-4 h-[calc(100%-2rem)]">
        <p className="text-xs text-muted-foreground mb-3">{t("workflows.apiTriggerDesc")}</p>
        <div className="relative">
          <div className="absolute top-2 left-3 text-[10px] text-muted-foreground font-medium uppercase tracking-wider">
            {t("workflows.curlExample")}
          </div>
          <pre suppressHydrationWarning className="pt-7 pb-3 px-3 bg-muted/50 rounded-lg text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all text-foreground/70 leading-relaxed">
{`curl -X POST \\
  ${typeof window !== "undefined" ? window.location.origin : ""}/api/v1/ext/orgs/${orgSlug}/workflows/${workflow.slug}/trigger \\
  -H "X-API-Key: amk_your_api_key_here" \\
  -H "Content-Type: application/json"`}
          </pre>
        </div>
        <p className="text-[10px] text-muted-foreground mt-2">{t("workflows.apiKeyHint")}</p>
        <Link href={`/${orgSlug}/settings`}
          className="inline-flex items-center gap-1 text-[10px] text-primary hover:underline mt-1">
          {t("workflows.manageApiKeys")}
          <ExternalLink className="w-2.5 h-2.5" />
        </Link>
      </div>
    </section>
  );
}
