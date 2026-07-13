"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Spinner } from "@/components/ui/spinner";
import type { CatalogSkill } from "@/lib/api";
import { AlertCircle, Plus, Tags } from "lucide-react";
import type { TranslationFn } from "../GeneralSettings";
import { CatalogSkillRow } from "./CatalogSkillRow";
import { SkillCatalogFilters, type CatalogViewMode } from "./SkillCatalogFilters";

interface CatalogSkillListProps {
  t: TranslationFn;
  loading: boolean;
  loadError?: boolean;
  skills: CatalogSkill[];
  syncingSlug: string | null;
  savingSlug: string | null;
  saveErrorSlug: string | null;
  onSync: (slug: string) => void;
  onDelete: (slug: string) => void;
  onImport: () => void;
  onRetry?: () => void;
  onEditTags: () => void;
  onUpdateTags: (slug: string, tags: string[]) => Promise<void>;
}

export function CatalogSkillList(props: CatalogSkillListProps) {
  const { t, loading, loadError, skills, onImport, onRetry } = props;
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [viewMode, setViewMode] = useState<CatalogViewMode>("flat");
  const tags = useMemo(
    () => [...new Set(skills.flatMap((skill) => skill.tags))].sort(),
    [skills],
  );
  const filtered = useMemo(
    () => selectedTags.length === 0
      ? skills
      : skills.filter((skill) => skill.tags.some((tag) => selectedTags.includes(tag))),
    [selectedTags, skills],
  );

  const toggleTag = (tag: string) => {
    setSelectedTags((current) => current.includes(tag)
      ? current.filter((item) => item !== tag)
      : [...current, tag]);
  };

  return (
    <section className="surface-card overflow-hidden">
      <div className="flex items-start justify-between gap-4 px-4 py-4">
        <div>
          <h2 className="text-base font-semibold">{t("extensions.skillCatalog.title")}</h2>
          <p className="mt-0.5 text-sm text-muted-foreground">
            {t("extensions.skillCatalog.description")}
          </p>
        </div>
        <Button onClick={onImport} className="shrink-0">
          <Plus className="mr-1 h-4 w-4" />
          {t("extensions.skillCatalog.import")}
        </Button>
      </div>

      {!loading && !loadError && skills.length > 0 && (
        <SkillCatalogFilters
          t={t}
          tags={tags}
          selectedTags={selectedTags}
          viewMode={viewMode}
          onTagToggle={toggleTag}
          onClear={() => setSelectedTags([])}
          onViewModeChange={setViewMode}
        />
      )}

      {loading ? (
        <div className="flex items-center justify-center gap-2 border-t border-border/60 py-10 text-sm text-muted-foreground">
          <Spinner size="sm" />
          {t("extensions.loading")}
        </div>
      ) : loadError ? (
        <EmptyState
          size="compact"
          icon={<AlertCircle className="h-5 w-5" />}
          title={t("extensions.failedToLoadSkills")}
          actions={onRetry && (
            <Button size="sm" variant="outline" onClick={onRetry}>
              {t("extensions.skillCatalog.retry")}
            </Button>
          )}
        />
      ) : skills.length === 0 ? (
        <EmptyState
          size="compact"
          icon={<Tags className="h-5 w-5" />}
          title={t("extensions.skillCatalog.noSkills")}
        />
      ) : filtered.length === 0 ? (
        <EmptyState
          size="compact"
          icon={<Tags className="h-5 w-5" />}
          title={t("extensions.skillCatalog.noFilterResults")}
          actions={(
            <Button size="sm" variant="outline" onClick={() => setSelectedTags([])}>
              {t("extensions.skillCatalog.clearFilters")}
            </Button>
          )}
        />
      ) : viewMode === "flat" ? (
        <div>{filtered.map((skill) => renderRow(skill, props, skill.id))}</div>
      ) : (
        <CatalogTagGroups skills={filtered} props={props} />
      )}
    </section>
  );
}

function CatalogTagGroups({
  skills,
  props,
}: {
  skills: CatalogSkill[];
  props: CatalogSkillListProps;
}) {
  const groups = new Map<string, CatalogSkill[]>();
  for (const skill of skills) {
    const groupTags = skill.tags.length > 0
      ? skill.tags
      : [props.t("extensions.skillCatalog.untagged")];
    for (const tag of groupTags) groups.set(tag, [...(groups.get(tag) ?? []), skill]);
  }

  return (
    <div className="divide-y divide-border/70">
      {[...groups.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([tag, groupSkills]) => (
        <section key={tag} role="region" aria-label={tag}>
          <div className="bg-surface-muted px-4 py-2 text-xs font-semibold text-muted-foreground">
            {tag} <span className="font-normal">({groupSkills.length})</span>
          </div>
          {groupSkills.map((skill) => renderRow(skill, props, `${tag}-${skill.id}`))}
        </section>
      ))}
    </div>
  );
}

function renderRow(skill: CatalogSkill, props: CatalogSkillListProps, key: React.Key) {
  return (
    <CatalogSkillRow
      key={key}
      t={props.t}
      skill={skill}
      syncing={props.syncingSlug === skill.slug}
      saving={props.savingSlug === skill.slug}
      saveFailed={props.saveErrorSlug === skill.slug}
      onSync={props.onSync}
      onDelete={props.onDelete}
      onEditTags={props.onEditTags}
      onUpdateTags={props.onUpdateTags}
    />
  );
}
