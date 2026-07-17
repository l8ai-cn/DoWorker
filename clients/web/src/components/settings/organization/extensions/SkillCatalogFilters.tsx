"use client";

import { Button } from "@/components/ui/button";
import { List, Tags, X } from "lucide-react";
import type { TranslationFn } from "../GeneralSettings";

export type CatalogViewMode = "flat" | "tags";

interface SkillCatalogFiltersProps {
  t: TranslationFn;
  tags: string[];
  selectedTags: string[];
  viewMode: CatalogViewMode;
  onTagToggle: (tag: string) => void;
  onClear: () => void;
  onViewModeChange: (mode: CatalogViewMode) => void;
}

export function SkillCatalogFilters({
  t,
  tags,
  selectedTags,
  viewMode,
  onTagToggle,
  onClear,
  onViewModeChange,
}: SkillCatalogFiltersProps) {
  return (
    <div className="flex flex-col gap-3 border-b border-border/60 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">
          {t("extensions.skillCatalog.filterTags")}
        </span>
        {tags.map((tag) => (
          <Button
            key={tag}
            type="button"
            size="sm"
            variant={selectedTags.includes(tag) ? "secondary" : "outline"}
            aria-pressed={selectedTags.includes(tag)}
            onClick={() => onTagToggle(tag)}
            className="h-7 px-2"
          >
            {tag}
          </Button>
        ))}
        {selectedTags.length > 0 && (
          <Button
            type="button"
            size="sm"
            variant="ghost"
            onClick={onClear}
            className="h-7 px-2 text-muted-foreground"
          >
            <X className="mr-1 h-3.5 w-3.5" />
            {t("extensions.skillCatalog.clearFilters")}
          </Button>
        )}
      </div>

      <div
        className="inline-flex self-start rounded-md bg-surface-muted p-1"
        role="group"
        aria-label={t("extensions.skillCatalog.viewMode")}
      >
        <button
          type="button"
          aria-label={t("extensions.skillCatalog.groupFlat")}
          aria-pressed={viewMode === "flat"}
          onClick={() => onViewModeChange("flat")}
          className={`flex h-7 items-center gap-1.5 rounded px-2 text-xs font-medium ${
            viewMode === "flat"
              ? "bg-surface-raised text-foreground shadow-xs"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <List className="h-3.5 w-3.5" />
          {t("extensions.skillCatalog.groupFlat")}
        </button>
        <button
          type="button"
          aria-label={t("extensions.skillCatalog.groupByTag")}
          aria-pressed={viewMode === "tags"}
          onClick={() => onViewModeChange("tags")}
          className={`flex h-7 items-center gap-1.5 rounded px-2 text-xs font-medium ${
            viewMode === "tags"
              ? "bg-surface-raised text-foreground shadow-xs"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <Tags className="h-3.5 w-3.5" />
          {t("extensions.skillCatalog.groupByTag")}
        </button>
      </div>
    </div>
  );
}
