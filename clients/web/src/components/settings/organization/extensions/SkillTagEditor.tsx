"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Pencil, X } from "lucide-react";
import type { TranslationFn } from "../GeneralSettings";

interface SkillTagEditorProps {
  t: TranslationFn;
  skillName: string;
  tags: string[];
  saving: boolean;
  saveFailed: boolean;
  onOpen: () => void;
  onSave: (tags: string[]) => Promise<void>;
}

function normalizeTags(tags: string[]): string[] {
  return [...new Set(tags.map((tag) => tag.trim().toLowerCase()).filter(Boolean))].sort();
}

export function SkillTagEditor({
  t,
  skillName,
  tags,
  saving,
  saveFailed,
  onOpen,
  onSave,
}: SkillTagEditorProps) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState(tags);
  const [input, setInput] = useState("");
  const tagLimitReached = draft.length >= 20;

  const setEditorOpen = (nextOpen: boolean) => {
    if (saving) return;
    setOpen(nextOpen);
    if (nextOpen) {
      onOpen();
      setDraft(tags);
      setInput("");
    }
  };

  const addInputTag = () => {
    if (tagLimitReached) return;
    const next = normalizeTags([...draft, input]);
    if (next.length === draft.length && input.trim() === "") return;
    setDraft(next);
    setInput("");
  };

  const save = async () => {
    try {
      await onSave(normalizeTags([...draft, input]));
      setOpen(false);
    } catch {
      // The parent owns the localized error state and toast.
    }
  };

  return (
    <Popover open={open} onOpenChange={setEditorOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`${t("extensions.skillCatalog.editTags")}: ${skillName}`}
          title={t("extensions.skillCatalog.editTags")}
        >
          <Pencil className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 space-y-3 p-3">
        <div>
          <p className="text-sm font-medium">{t("extensions.skillCatalog.editTags")}</p>
          <p className="truncate text-xs text-muted-foreground">{skillName}</p>
        </div>
        <div className="flex min-h-7 flex-wrap gap-1.5">
          {draft.length === 0 && (
            <span className="text-xs text-muted-foreground">
              {t("extensions.skillCatalog.untagged")}
            </span>
          )}
          {draft.map((tag) => (
            <Badge key={tag} variant="secondary" className="gap-1 pr-1">
              {tag}
              <button
                type="button"
                disabled={saving}
                aria-label={`${t("extensions.skillCatalog.removeTag")}: ${tag}`}
                onClick={() => setDraft((current) => current.filter((item) => item !== tag))}
                className="rounded-sm p-0.5 hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
        <Input
          value={input}
          disabled={saving || tagLimitReached}
          maxLength={40}
          aria-label={t("extensions.skillCatalog.tagInput")}
          placeholder={t("extensions.skillCatalog.tagInputPlaceholder")}
          onChange={(event) => setInput(event.target.value)}
          onKeyDown={(event) => {
            if (event.key !== "Enter") return;
            event.preventDefault();
            addInputTag();
          }}
        />
        {saveFailed && (
          <p role="alert" className="text-xs text-destructive">
            {t("extensions.skillCatalog.failedToSaveTags")}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            size="sm"
            variant="ghost"
            disabled={saving}
            onClick={() => setEditorOpen(false)}
          >
            {t("extensions.skillCatalog.cancelTags")}
          </Button>
          <Button type="button" size="sm" disabled={saving} onClick={save}>
            {saving
              ? t("extensions.skillCatalog.savingTags")
              : t("extensions.skillCatalog.saveTags")}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
