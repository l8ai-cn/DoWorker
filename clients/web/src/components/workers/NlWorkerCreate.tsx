"use client";

import { Sparkles, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

interface NlWorkerCreateProps {
  prompt: string;
  filling: boolean;
  onPromptChange: (prompt: string) => void;
  onFill: (prompt: string) => void;
}

export function NlWorkerCreate({
  prompt,
  filling,
  onPromptChange,
  onFill,
}: NlWorkerCreateProps) {
  const t = useTranslations();
  const trimmed = prompt.trim();

  return (
    <section
      className="mb-6 rounded-lg border border-border bg-surface-raised p-4"
      data-testid="worker-fill-panel"
    >
      <div className="mb-3 flex items-center gap-2">
        <Sparkles className="h-4 w-4 text-primary" />
        <h2 className="text-sm font-medium">{t("workers.create.nl.title")}</h2>
      </div>
      <Textarea
        value={prompt}
        onChange={(event) => onPromptChange(event.target.value)}
        placeholder={t("workers.create.nl.placeholder")}
        rows={3}
        maxLength={10000}
        disabled={filling}
        onKeyDown={(e) => {
          if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            e.preventDefault();
            if (trimmed && !filling) onFill(trimmed);
          }
        }}
      />
      <div className="mt-3 flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {t("workers.create.nl.hint")}
        </p>
        <Button type="button" onClick={() => onFill(trimmed)} disabled={!trimmed || filling}>
          {filling && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t("workers.create.nl.submit")}
        </Button>
      </div>
    </section>
  );
}
