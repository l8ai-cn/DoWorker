"use client";

import { useCallback } from "react";
import { Sparkles, X } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import type { InstalledSkill } from "@/lib/viewModels/extension";

interface Props {
  skills: InstalledSkill[];
  selectedSlugs: string[];
  onChange: (slugs: string[]) => void;
  loading?: boolean;
  t: (key: string) => string;
}

export function SkillMultiSelect({
  skills,
  selectedSlugs = [],
  onChange,
  loading,
  t,
}: Props) {
  const toggle = useCallback(
    (slug: string) => {
      if (selectedSlugs.includes(slug)) {
        onChange(selectedSlugs.filter((s) => s !== slug));
      } else {
        onChange([...selectedSlugs, slug]);
      }
    },
    [selectedSlugs, onChange],
  );

  const remove = useCallback(
    (slug: string) => onChange(selectedSlugs.filter((s) => s !== slug)),
    [selectedSlugs, onChange],
  );

  if (loading) {
    return (
      <div>
        <label className="block text-sm font-medium mb-2">
          {t("ide.createPod.selectSkills")}
        </label>
        <div className="flex items-center text-sm text-muted-foreground py-2">
          <Spinner size="sm" className="mr-2" />
          {t("common.loading")}
        </div>
      </div>
    );
  }

  return (
    <div>
      <label className="block text-sm font-medium mb-2">
        {t("ide.createPod.selectSkills")}
      </label>

      {selectedSlugs.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {selectedSlugs.map((slug) => (
            <span
              key={slug}
              className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/30 px-2 py-0.5 text-xs"
            >
              <Sparkles className="w-3 h-3 text-primary shrink-0" />
              <span className="truncate max-w-[12rem]" title={slug}>
                {slug}
              </span>
              <button
                type="button"
                className="text-muted-foreground hover:text-destructive shrink-0"
                onClick={() => remove(slug)}
                title={t("common.delete")}
                aria-label={t("common.delete")}
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          ))}
        </div>
      )}

      {skills.length === 0 ? (
        <p className="text-xs text-muted-foreground py-2">
          {t("ide.createPod.noSkillsAvailableHint")}
        </p>
      ) : (
        <div className="surface-card max-h-48 overflow-y-auto">
          {skills.map((skill) => {
            const checked = selectedSlugs.includes(skill.slug);
            return (
              <label
                key={skill.id}
                className="flex items-center gap-2 px-2 py-1.5 border-b border-border last:border-b-0 motion-interactive hover:bg-surface-muted cursor-pointer"
              >
                <input
                  type="checkbox"
                  className="h-3.5 w-3.5"
                  checked={checked}
                  onChange={() => toggle(skill.slug)}
                />
                <Sparkles className="w-4 h-4 text-muted-foreground shrink-0" />
                <span className="text-sm flex-1 truncate" title={skill.slug}>
                  {skill.slug}
                </span>
                <span className="text-[10px] uppercase tracking-wide text-muted-foreground shrink-0">
                  {skill.scope === "org"
                    ? t("ide.createPod.skillScope.org")
                    : t("ide.createPod.skillScope.user")}
                </span>
              </label>
            );
          })}
        </div>
      )}

      <p className="text-xs text-muted-foreground mt-1">
        {selectedSlugs.length === 0
          ? t("ide.createPod.noSkillsSelectedHint")
          : t("ide.createPod.skillsSelectedHint")}
      </p>
    </div>
  );
}

export default SkillMultiSelect;
