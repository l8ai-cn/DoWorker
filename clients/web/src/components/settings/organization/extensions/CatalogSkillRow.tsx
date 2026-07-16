"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { CatalogSkill } from "@/lib/api";
import { GitBranch, PenLine, RefreshCw, Trash2 } from "lucide-react";
import type { TranslationFn } from "../GeneralSettings";
import { SkillTagEditor } from "./SkillTagEditor";

interface CatalogSkillRowProps {
  t: TranslationFn;
  skill: CatalogSkill;
  syncing: boolean;
  saving: boolean;
  saveFailed: boolean;
  onSync: (slug: string) => void;
  onDelete: (slug: string) => void;
  onEditTags: () => void;
  onUpdateTags: (slug: string, tags: string[]) => Promise<void>;
}

export function CatalogSkillRow({
  t,
  skill,
  syncing,
  saving,
  saveFailed,
  onSync,
  onDelete,
  onEditTags,
  onUpdateTags,
}: CatalogSkillRowProps) {
  const imported = skill.install_source === "import";
  const name = skill.display_name || skill.slug;

  return (
    <div className="flex items-start justify-between gap-3 border-b border-border/50 px-4 py-3 last:border-b-0">
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="truncate text-sm font-medium">{name}</span>
          <Badge variant={imported ? "secondary" : "outline"} className="shrink-0 text-xs">
            {imported ? <GitBranch className="mr-1 h-3 w-3" /> : <PenLine className="mr-1 h-3 w-3" />}
            {t(`extensions.skillCatalog.${imported ? "sourceImport" : "sourceGitops"}`)}
          </Badge>
          {skill.version > 0 && <Badge variant="outline">v{skill.version}</Badge>}
          {!skill.is_active && (
            <Badge variant="destructive">{t("extensions.skillCatalog.inactive")}</Badge>
          )}
        </div>
        <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">{skill.slug}</p>
        {skill.description && (
          <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{skill.description}</p>
        )}
        <div className="mt-2 flex flex-wrap gap-1.5">
          {skill.tags.length > 0 ? skill.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="font-normal">{tag}</Badge>
          )) : (
            <span className="text-xs text-muted-foreground">
              {t("extensions.skillCatalog.untagged")}
            </span>
          )}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1">
        <SkillTagEditor
          t={t}
          skillName={name}
          tags={skill.tags}
          saving={saving}
          saveFailed={saveFailed}
          onOpen={onEditTags}
          onSave={(tags) => onUpdateTags(skill.slug, tags)}
        />
        {imported && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            disabled={syncing}
            onClick={() => onSync(skill.slug)}
            aria-label={`${t("extensions.skillCatalog.syncUpstream")}: ${name}`}
            title={t("extensions.skillCatalog.syncUpstream")}
          >
            <RefreshCw className={`h-4 w-4 ${syncing ? "animate-spin" : ""}`} />
          </Button>
        )}
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => onDelete(skill.slug)}
          aria-label={`${t("extensions.skillCatalog.delete")}: ${name}`}
          title={t("extensions.skillCatalog.delete")}
          className="text-destructive hover:text-destructive"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
