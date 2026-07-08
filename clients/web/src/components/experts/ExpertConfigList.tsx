"use client";

import { useTranslations } from "next-intl";
import {
  BookOpen,
  Bot,
  GitBranch,
  MessageSquareText,
  Package,
  Server,
  Sparkles,
  Terminal,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { parseExpertKnowledgeMounts, type Expert } from "@/lib/api/expertApi";

function SectionCard({
  icon: Icon,
  title,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="surface-card rounded-xl border border-border overflow-hidden">
      <header className="flex items-center gap-2 border-b border-border/60 bg-muted/30 px-4 py-2.5">
        <Icon className="h-4 w-4 text-muted-foreground" />
        <h3 className="text-sm font-medium">{title}</h3>
      </header>
      <div className="p-4">{children}</div>
    </section>
  );
}

function FieldRow({
  icon: Icon,
  label,
  value,
}: {
  icon?: React.ComponentType<{ className?: string }>;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3 py-2.5 border-b border-border/40 last:border-0">
      <dt className="flex w-32 shrink-0 items-center gap-1.5 text-sm text-muted-foreground">
        {Icon && <Icon className="h-3.5 w-3.5" />}
        {label}
      </dt>
      <dd className="min-w-0 flex-1 text-sm break-words">{value}</dd>
    </div>
  );
}

function ChipList({
  items,
  empty,
  variant = "secondary",
}: {
  items: string[];
  empty: string;
  variant?: "secondary" | "info" | "success";
}) {
  if (!items.length) {
    return <span className="text-sm text-muted-foreground">{empty}</span>;
  }
  return (
    <div className="flex flex-wrap gap-1.5">
      {items.map((item) => (
        <Badge key={item} variant={variant} className="font-normal">
          {item}
        </Badge>
      ))}
    </div>
  );
}

export function ExpertConfigList({ expert }: { expert: Expert }) {
  const t = useTranslations("experts");
  const mounts = parseExpertKnowledgeMounts(expert.knowledge_mounts);
  const knowledgeItems = mounts.map((m) => (m.mode === "rw" ? `${m.slug} · rw` : m.slug));

  const modeLabel =
    expert.interaction_mode?.toLowerCase() === "acp"
      ? t("edit.modeAcp")
      : t("edit.modePty");

  return (
    <div className="flex-1 overflow-y-auto px-8 py-6">
      <div className="mx-auto flex max-w-3xl flex-col gap-4">
        <SectionCard icon={Server} title={t("configSectionRuntime")}>
          <dl>
            <FieldRow
              icon={Bot}
              label={t("agent")}
              value={<span className="font-mono text-xs">{expert.agent_slug}</span>}
            />
            <FieldRow
              icon={Terminal}
              label={t("interactionMode")}
              value={<Badge variant="outline" className="font-normal">{modeLabel}</Badge>}
            />
            <FieldRow
              icon={Zap}
              label={t("perpetual")}
              value={
                <Badge variant={expert.perpetual ? "success" : "secondary"} className="font-normal">
                  {expert.perpetual ? t("perpetualOn") : t("perpetualOff")}
                </Badge>
              }
            />
            {expert.runner_id != null && (
              <FieldRow
                icon={Server}
                label={t("runner")}
                value={<span className="font-mono text-xs">#{expert.runner_id}</span>}
              />
            )}
            {expert.repository_id != null && (
              <FieldRow
                icon={GitBranch}
                label={t("repository")}
                value={<span className="font-mono text-xs">#{expert.repository_id}</span>}
              />
            )}
            {expert.branch_name && (
              <FieldRow
                icon={GitBranch}
                label={t("branch")}
                value={<span className="font-mono text-xs">{expert.branch_name}</span>}
              />
            )}
          </dl>
        </SectionCard>

        <SectionCard icon={Sparkles} title={t("configSectionCapabilities")}>
          <dl>
            <FieldRow
              icon={Sparkles}
              label={t("skills")}
              value={<ChipList items={expert.skill_slugs ?? []} empty={t("noSkills")} variant="info" />}
            />
            <FieldRow
              icon={BookOpen}
              label={t("knowledge")}
              value={<ChipList items={knowledgeItems} empty={t("noKnowledge")} variant="success" />}
            />
            <FieldRow
              icon={Package}
              label={t("envBundles")}
              value={<ChipList items={expert.used_env_bundles ?? []} empty={t("noEnvBundles")} />}
            />
          </dl>
        </SectionCard>

        <SectionCard icon={MessageSquareText} title={t("prompt")}>
          {expert.prompt?.trim() ? (
            <pre className="max-h-72 overflow-auto whitespace-pre-wrap rounded-lg bg-muted/50 p-3 font-mono text-xs leading-relaxed">
              {expert.prompt}
            </pre>
          ) : (
            <p className="text-sm text-muted-foreground">{t("noPrompt")}</p>
          )}
        </SectionCard>

        {expert.source_pod_key && (
          <p className="px-1 text-xs text-muted-foreground">
            {t("publishedFrom", { podKey: expert.source_pod_key })}
          </p>
        )}
      </div>
    </div>
  );
}
