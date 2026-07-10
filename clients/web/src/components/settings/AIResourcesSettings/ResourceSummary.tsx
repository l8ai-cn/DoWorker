import { useTranslations } from "next-intl";
import type { EffectiveResource, ProviderConnection } from "@/lib/api";

interface ResourceSummaryProps {
  connections: ProviderConnection[];
  effectiveResources: EffectiveResource[];
}

export function ResourceSummary({ connections, effectiveResources }: ResourceSummaryProps) {
  const t = useTranslations();
  const resources = connections.flatMap((connection) => connection.resources);
  const hasUsage = resources.some((resource) => resource.usageSummary);
  const selectable = effectiveResources.filter((resource) => resource.selectable).length;

  return (
    <dl className="grid gap-3 sm:grid-cols-3">
      <SummaryValue label={t("settings.aiResources.summary.connections")} value={String(connections.length)} />
      <SummaryValue label={t("settings.aiResources.summary.selectable")} value={String(selectable)} />
      <SummaryValue
        label={t("settings.aiResources.summary.usage")}
        value={hasUsage ? t("settings.aiResources.usageAvailable") : t("settings.aiResources.usageNotConnected")}
      />
    </dl>
  );
}

function SummaryValue({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/60 bg-muted/30 px-3 py-3">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 text-sm font-semibold text-foreground">{value}</dd>
    </div>
  );
}
