"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { SkillMarketItem } from "@/lib/api";
import type { TranslationFn } from "../GeneralSettings";
import { skillAccent } from "./skill-market-accent";

interface SkillMarketCardProps {
  skill: SkillMarketItem;
  t: TranslationFn;
  onView: () => void;
  onInstall: () => void;
}

export function SkillMarketCard({ skill, t, onView, onInstall }: SkillMarketCardProps) {
  const name = skill.display_name || skill.slug;
  const accent = skillAccent(skill.category || skill.slug);
  const initial = name.trim().charAt(0).toUpperCase() || "?";

  return (
    <div
      className="group flex flex-col gap-3 rounded-xl border border-border p-4 cursor-pointer transition-all hover:border-primary/40 hover:shadow-sm"
      onClick={onView}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onView();
        }
      }}
      role="button"
      tabIndex={0}
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-sm font-semibold",
            accent.bg,
            accent.text,
          )}
        >
          {initial}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate font-medium">{name}</span>
            {skill.version > 0 && (
              <span className="shrink-0 text-xs text-muted-foreground">v{skill.version}</span>
            )}
          </div>
          <p className="truncate font-mono text-xs text-muted-foreground">{skill.slug}</p>
        </div>
      </div>

      {skill.description && (
        <p className="line-clamp-2 min-h-[2.5rem] text-sm text-muted-foreground">
          {skill.description}
        </p>
      )}

      {(skill.category || skill.license) && (
        <div className="flex flex-wrap items-center gap-1.5">
          {skill.category && (
            <Badge variant="outline" className="text-xs">
              {skill.category}
            </Badge>
          )}
          {skill.license && (
            <Badge
              variant="secondary"
              className="max-w-[10rem] truncate text-xs"
              title={skill.license}
            >
              {skill.license}
            </Badge>
          )}
        </div>
      )}

      <div className="mt-auto flex items-center justify-between gap-2 pt-1">
        <div className="flex min-w-0 items-center gap-3 text-xs text-muted-foreground">
          {skill.content_sha && (
            <span className="truncate font-mono" title={skill.content_sha}>
              {skill.content_sha.slice(0, 8)}
            </span>
          )}
        </div>
        <Button
          size="sm"
          className="shrink-0"
          onClick={(e) => {
            e.stopPropagation();
            onInstall();
          }}
        >
          {t("extensions.install")}
        </Button>
      </div>
    </div>
  );
}
