"use client";

import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ExpertSkillSlugsField } from "./ExpertSkillSlugsField";
import { ExpertConfigFields } from "./ExpertConfigFields";
import { useExpertEditForm } from "./useExpertEditForm";
import type { Expert } from "@/lib/api/expertApi";

const SELECT_CLASS =
  "h-9 w-full rounded-md bg-surface-raised px-3 text-sm ring-1 ring-border/35 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  expert?: Expert | null;
  onSaved?: (expert: Expert) => void;
}

export function ExpertEditDrawer({ open, onOpenChange, expert = null, onSaved }: Props) {
  const t = useTranslations("experts.edit");
  const f = useExpertEditForm(open, expert);

  const handleSubmit = async () => {
    const saved = await f.submit();
    if (!saved) return;
    toast.success(f.isEdit ? t("updateSuccess") : t("createSuccess"));
    onOpenChange(false);
    onSaved?.(saved);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-full max-w-md flex-col overflow-hidden">
        <SheetHeader>
          <SheetTitle>{f.isEdit ? t("editTitle") : t("createTitle")}</SheetTitle>
          <SheetDescription>{f.isEdit ? t("editDescription") : t("createDescription")}</SheetDescription>
        </SheetHeader>

        <div className="-mx-1 flex-1 space-y-4 overflow-y-auto px-1 py-4">
          <div className="space-y-1.5">
            <Label htmlFor="expert-name">{t("nameLabel")}</Label>
            <Input
              id="expert-name"
              value={f.form.name}
              onChange={(e) => f.patch({ name: e.target.value })}
              placeholder={t("namePlaceholder")}
            />
          </div>

          {!f.isEdit && (
            <>
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
            </>
          )}

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
            <Label htmlFor="expert-prompt">{t("promptLabel")}</Label>
            <Textarea
              id="expert-prompt"
              value={f.form.prompt}
              onChange={(e) => f.patch({ prompt: e.target.value })}
              placeholder={t("promptPlaceholder")}
              className="min-h-[96px] font-mono text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="expert-mode">{t("interactionModeLabel")}</Label>
            <select
              id="expert-mode"
              className={SELECT_CLASS}
              value={f.form.interactionMode}
              onChange={(e) => f.patch({ interactionMode: e.target.value })}
            >
              <option value="pty">{t("modePty")}</option>
              <option value="acp">{t("modeAcp")}</option>
            </select>
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              className="h-3.5 w-3.5"
              checked={f.form.perpetual}
              onChange={(e) => f.patch({ perpetual: e.target.checked })}
            />
            {t("perpetualLabel")}
          </label>

          <div className="space-y-1.5">
            <Label>{t("skillsLabel")}</Label>
            <ExpertSkillSlugsField
              value={f.form.skillSlugs}
              onChange={(skillSlugs) => f.patch({ skillSlugs })}
              emptyLabel={t("skillsEmpty")}
              addLabel={t("skillsAdd")}
              placeholder={t("skillsPlaceholder")}
              removeLabel={t("skillsRemove")}
            />
          </div>

          <ExpertConfigFields open={open} form={f.form} patch={f.patch} />

          {f.error && <p className="text-xs text-destructive">{f.error}</p>}
        </div>

        <SheetFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!f.canSubmit || f.submitting}>
            {f.submitting ? t("saving") : f.isEdit ? t("save") : t("create")}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
