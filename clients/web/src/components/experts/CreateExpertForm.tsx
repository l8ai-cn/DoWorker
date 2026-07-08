"use client";

import type { ReactNode } from "react";
import { useRouter, useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Bot, FileText, Sparkles, BookOpen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ExpertSkillSlugsField } from "./ExpertSkillSlugsField";
import { ExpertAvatarField } from "./ExpertAvatarField";
import { ExpertTypeSelect } from "./ExpertTypeSelect";
import { KnowledgeBaseMountSelect } from "@/components/pod/CreatePodForm/KnowledgeBaseMountSelect";
import { useExpertEditForm } from "./useExpertEditForm";

const SELECT_CLASS =
  "h-9 w-full rounded-md bg-surface-raised px-3 text-sm ring-1 ring-border/35 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35";

interface SectionProps {
  index: number;
  icon: ReactNode;
  title: string;
  description?: string;
  children: ReactNode;
}

function Section({ index, icon, title, description, children }: SectionProps) {
  return (
    <section className="surface-card space-y-4 p-5">
      <div className="flex items-start gap-3">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
          {icon}
        </div>
        <div className="min-w-0">
          <h2 className="text-sm font-semibold">
            <span className="mr-1.5 text-muted-foreground">{index}.</span>
            {title}
          </h2>
          {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
        </div>
      </div>
      <div className="space-y-4">{children}</div>
    </section>
  );
}

export function CreateExpertForm() {
  const t = useTranslations("experts.create");
  const router = useRouter();
  const params = useParams();
  const orgSlug = String(params.org ?? "");
  // open=true, expert=null -> create mode (loads agents, resets to empty form).
  const f = useExpertEditForm(true, null);

  const handleSubmit = async () => {
    const saved = await f.submit();
    if (!saved) return;
    toast.success(t("createSuccess"));
    router.push(`/${orgSlug}/experts/${saved.slug}`);
  };

  return (
    <div className="mx-auto flex h-full w-full max-w-3xl flex-col overflow-y-auto px-6 py-6">
      <div className="mb-5">
        <h1 className="text-lg font-semibold">{t("title")}</h1>
        <p className="mt-0.5 text-sm text-muted-foreground">{t("subtitle")}</p>
      </div>

      <div className="space-y-4">
        <Section
          index={0}
          icon={<Bot className="h-4 w-4" />}
          title={t("basicsTitle")}
          description={t("basicsDescription")}
        >
          <div className="space-y-1.5">
            <Label htmlFor="expert-name">{t("nameLabel")}</Label>
            <Input
              id="expert-name"
              value={f.form.name}
              onChange={(e) => f.patch({ name: e.target.value })}
              placeholder={t("namePlaceholder")}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="expert-slug">{t("slugLabel")}</Label>
            <Input
              id="expert-slug"
              value={f.form.slug}
              onChange={(e) => f.setSlug(e.target.value)}
              placeholder={t("slugPlaceholder")}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="expert-description">{t("descriptionLabel")}</Label>
            <Textarea
              id="expert-description"
              value={f.form.description}
              onChange={(e) => f.patch({ description: e.target.value })}
              placeholder={t("descriptionPlaceholder")}
              className="min-h-[64px]"
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t("avatarLabel")}</Label>
            <ExpertAvatarField
              value={f.form.avatar}
              onChange={(avatar) => f.patch({ avatar })}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="expert-type">{t("typeLabel")}</Label>
            <ExpertTypeSelect
              id="expert-type"
              value={f.form.expertType}
              onChange={(expertType) => f.patch({ expertType })}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="expert-agent">{t("agentLabel")}</Label>
            <select
              id="expert-agent"
              className={SELECT_CLASS}
              value={f.form.agentSlug}
              onChange={(e) => f.patch({ agentSlug: e.target.value })}
            >
              <option value="" disabled>
                {t("agentPlaceholder")}
              </option>
              {f.agents.map((a) => (
                <option key={a.slug} value={a.slug}>
                  {a.name || a.slug}
                </option>
              ))}
            </select>
          </div>
        </Section>

        <Section
          index={1}
          icon={<FileText className="h-4 w-4" />}
          title={t("agentfileTitle")}
          description={t("agentfileDescription")}
        >
          <div className="space-y-1.5">
            <Label htmlFor="expert-agentfile">{t("agentfileLabel")}</Label>
            <Textarea
              id="expert-agentfile"
              value={f.form.agentfileLayer}
              onChange={(e) => f.patch({ agentfileLayer: e.target.value })}
              placeholder={t("agentfilePlaceholder")}
              className="min-h-[200px] font-mono text-xs"
            />
          </div>
        </Section>

        <Section
          index={2}
          icon={<Sparkles className="h-4 w-4" />}
          title={t("skillsTitle")}
          description={t("skillsDescription")}
        >
          <ExpertSkillSlugsField
            value={f.form.skillSlugs}
            onChange={(skillSlugs) => f.patch({ skillSlugs })}
            emptyLabel={t("skillsEmpty")}
            addLabel={t("skillsAdd")}
            placeholder={t("skillsPlaceholder")}
            removeLabel={t("skillsRemove")}
          />
        </Section>

        <Section
          index={3}
          icon={<BookOpen className="h-4 w-4" />}
          title={t("knowledgeTitle")}
          description={t("knowledgeDescription")}
        >
          <KnowledgeBaseMountSelect
            selectedMounts={f.form.knowledgeMounts}
            onChange={(knowledgeMounts) => f.patch({ knowledgeMounts })}
            embedded
          />
        </Section>

        {f.error && <p className="text-xs text-destructive">{f.error}</p>}

        <div className="flex items-center justify-end gap-2 pb-4">
          <Button variant="outline" onClick={() => router.push(`/${orgSlug}/experts`)}>
            {t("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!f.canSubmit || f.submitting}>
            {f.submitting ? t("creating") : t("create")}
          </Button>
        </div>
      </div>
    </div>
  );
}
