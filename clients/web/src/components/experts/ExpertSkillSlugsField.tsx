"use client";

import { useState, useCallback, type KeyboardEvent } from "react";
import { Sparkles, X, Plus } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

function normalizeSlug(raw: string): string {
  return raw
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 100);
}

interface Props {
  value: string[];
  onChange: (slugs: string[]) => void;
  emptyLabel: string;
  addLabel: string;
  placeholder: string;
  removeLabel: string;
}

export function ExpertSkillSlugsField({
  value,
  onChange,
  emptyLabel,
  addLabel,
  placeholder,
  removeLabel,
}: Props) {
  const [draft, setDraft] = useState("");

  const add = useCallback(() => {
    const slug = normalizeSlug(draft);
    if (!slug || value.includes(slug)) {
      setDraft("");
      return;
    }
    onChange([...value, slug]);
    setDraft("");
  }, [draft, value, onChange]);

  const remove = useCallback(
    (slug: string) => onChange(value.filter((s) => s !== slug)),
    [value, onChange],
  );

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      add();
    }
  };

  return (
    <div>
      {value.length > 0 ? (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {value.map((slug) => (
            <span
              key={slug}
              className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/30 px-2 py-0.5 text-xs"
            >
              <Sparkles className="h-3 w-3 shrink-0 text-primary" />
              <span className="max-w-[12rem] truncate" title={slug}>
                {slug}
              </span>
              <button
                type="button"
                className="shrink-0 text-muted-foreground hover:text-destructive"
                onClick={() => remove(slug)}
                aria-label={removeLabel}
                title={removeLabel}
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      ) : (
        <p className="mb-2 text-xs text-muted-foreground">{emptyLabel}</p>
      )}

      <div className="flex items-center gap-2">
        <Input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder={placeholder}
          className="h-8 text-sm"
        />
        <Button
          type="button"
          size="sm"
          variant="outline"
          className="h-8 shrink-0 gap-1"
          onClick={add}
          disabled={!normalizeSlug(draft)}
        >
          <Plus className="h-3.5 w-3.5" />
          {addLabel}
        </Button>
      </div>
    </div>
  );
}
