"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import type { SkillMarketItem } from "@/lib/api";
import type { TranslationFn } from "../GeneralSettings";

interface SkillMarketDetailDrawerProps {
  skill: SkillMarketItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onInstall: (skill: SkillMarketItem) => void;
  t: TranslationFn;
}

export function SkillMarketDetailDrawer({
  skill,
  open,
  onOpenChange,
  onInstall,
  t,
}: SkillMarketDetailDrawerProps) {
  if (!skill) return null;

  const name = skill.display_name || skill.slug;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full max-w-md overflow-y-auto">
        <SheetHeader>
          <SheetTitle>{name}</SheetTitle>
        </SheetHeader>
        <div className="mt-6 space-y-4">
          <div className="flex flex-wrap gap-2">
            {skill.category && (
              <Badge variant="outline">{skill.category}</Badge>
            )}
            {skill.license && (
              <Badge variant="secondary" className="max-w-full truncate" title={skill.license}>
                {skill.license}
              </Badge>
            )}
            {skill.version > 0 && (
              <Badge variant="outline">v{skill.version}</Badge>
            )}
          </div>

          {skill.description ? (
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">
              {skill.description}
            </p>
          ) : (
            <p className="text-sm text-muted-foreground italic">
              {t("extensions.skillMarket.noDescription")}
            </p>
          )}

          <dl className="space-y-3 text-sm">
            <DetailRow label={t("extensions.slug")} value={skill.slug} mono />
            {skill.content_sha && (
              <DetailRow
                label={t("extensions.skillMarket.contentSha")}
                value={skill.content_sha}
                mono
              />
            )}
          </dl>

          <Button className="w-full" onClick={() => onInstall(skill)}>
            {t("extensions.install")}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function DetailRow({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div>
      <dt className="text-muted-foreground mb-1">{label}</dt>
      <dd className={mono ? "font-mono text-xs break-all" : "break-all"}>{value}</dd>
    </div>
  );
}
