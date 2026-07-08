"use client";

import { useTranslations } from "next-intl";
import { parseExpertKnowledgeMounts, type Expert } from "@/lib/api/expertApi";

function ConfigRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[140px_1fr] gap-3 py-2 border-b border-border/50 last:border-0">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="text-sm break-words">{value}</dd>
    </div>
  );
}

export function ExpertConfigList({ expert }: { expert: Expert }) {
  const t = useTranslations("experts");
  const mounts = parseExpertKnowledgeMounts(expert.knowledge_mounts);

  return (
    <div className="flex-1 overflow-y-auto px-8 py-6">
      <h2 className="text-sm font-medium mb-3">{t("configTitle")}</h2>
      <dl className="surface-card rounded-lg border border-border px-4">
        <ConfigRow label={t("agent")} value={expert.agent_slug} />
        <ConfigRow label={t("interactionMode")} value={expert.interaction_mode} />
        <ConfigRow label={t("perpetual")} value={expert.perpetual ? "Yes" : "No"} />
        {expert.runner_id && <ConfigRow label={t("runner")} value={`#${expert.runner_id}`} />}
        {expert.repository_id && (
          <ConfigRow label={t("repository")} value={`#${expert.repository_id}`} />
        )}
        {expert.branch_name && <ConfigRow label={t("branch")} value={expert.branch_name} />}
        <ConfigRow
          label={t("prompt")}
          value={
            expert.prompt?.trim() ? (
              <pre className="whitespace-pre-wrap font-mono text-xs bg-muted/50 rounded p-2">
                {expert.prompt}
              </pre>
            ) : (
              t("noPrompt")
            )
          }
        />
        <ConfigRow
          label={t("skills")}
          value={expert.skill_slugs?.length ? expert.skill_slugs.join(", ") : t("noSkills")}
        />
        <ConfigRow
          label={t("knowledge")}
          value={
            mounts.length
              ? mounts.map((m) => (m.mode === "rw" ? `${m.slug} [rw]` : m.slug)).join(", ")
              : t("noKnowledge")
          }
        />
        <ConfigRow
          label={t("envBundles")}
          value={
            expert.used_env_bundles?.length
              ? expert.used_env_bundles.join(", ")
              : t("noEnvBundles")
          }
        />
        {expert.source_pod_key && (
          <ConfigRow label="Source" value={t("publishedFrom", { podKey: expert.source_pod_key })} />
        )}
      </dl>
    </div>
  );
}
