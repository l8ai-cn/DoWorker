"use client";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { CatalogSkill } from "@/lib/api";
import { RefreshCw, Trash2, Plus, GitBranch, PenLine } from "lucide-react";
import type { TranslationFn } from "../GeneralSettings";

interface CatalogSkillListProps {
  t: TranslationFn;
  loading: boolean;
  skills: CatalogSkill[];
  syncingSlug: string | null;
  onSync: (slug: string) => void;
  onDelete: (slug: string) => void;
  onImport: () => void;
}

export function CatalogSkillList({
  t,
  loading,
  skills,
  syncingSlug,
  onSync,
  onDelete,
  onImport,
}: CatalogSkillListProps) {
  return (
    <div className="surface-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{t("extensions.skillCatalog.title")}</h2>
          <p className="text-sm text-muted-foreground">
            {t("extensions.skillCatalog.description")}
          </p>
        </div>
        <Button onClick={onImport}>
          <Plus className="w-4 h-4 mr-1" />
          {t("extensions.skillCatalog.import")}
        </Button>
      </div>

      {loading ? (
        <div className="text-center py-4 text-muted-foreground">{t("extensions.loading")}</div>
      ) : skills.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          {t("extensions.skillCatalog.noSkills")}
        </div>
      ) : (
        <div className="space-y-3">
          {skills.map((skill) => {
            const imported = skill.install_source === "import";
            return (
              <div
                key={skill.id}
                className="surface-card p-4 flex items-center justify-between gap-4"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium truncate">{skill.display_name || skill.slug}</span>
                    <Badge variant={imported ? "secondary" : "outline"} className="text-xs shrink-0">
                      {imported ? (
                        <>
                          <GitBranch className="w-3 h-3 mr-1" />
                          {t("extensions.skillCatalog.sourceImport")}
                        </>
                      ) : (
                        <>
                          <PenLine className="w-3 h-3 mr-1" />
                          {t("extensions.skillCatalog.sourceGitops")}
                        </>
                      )}
                    </Badge>
                    {skill.version > 0 && (
                      <Badge variant="outline" className="text-xs shrink-0">v{skill.version}</Badge>
                    )}
                    {!skill.is_active && (
                      <Badge variant="destructive" className="text-xs shrink-0">
                        {t("extensions.skillCatalog.inactive")}
                      </Badge>
                    )}
                  </div>
                  <p className="font-mono text-xs text-muted-foreground truncate">{skill.slug}</p>
                  {imported && skill.upstream_url && (
                    <p className="text-xs text-muted-foreground truncate mt-0.5" title={skill.upstream_url}>
                      {skill.upstream_url}
                      {skill.upstream_subdir ? ` · ${skill.upstream_subdir}` : ""}
                    </p>
                  )}
                  {skill.description && (
                    <p className="text-xs text-muted-foreground line-clamp-2 mt-1">{skill.description}</p>
                  )}
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  {imported && (
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={syncingSlug === skill.slug}
                      onClick={() => onSync(skill.slug)}
                      title={t("extensions.skillCatalog.syncUpstream")}
                    >
                      <RefreshCw className={`w-4 h-4 ${syncingSlug === skill.slug ? "animate-spin" : ""}`} />
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(skill.slug)}
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
