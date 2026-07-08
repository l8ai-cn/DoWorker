"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Sparkles, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { quickTaskApi } from "@/lib/api/quickTaskApi";
import { getApiErrorCode, getLocalizedErrorMessage } from "@/lib/api";
import { getShortPodKey } from "@/lib/pod-display-name";

const WIZARD_FALLBACK_CODES = new Set([
  "NO_RUNNER_FOR_AGENT",
  "AGENT_NOT_FOUND",
]);

interface NlWorkerCreateProps {
  orgSlug: string;
  onNeedsWizard: (prompt: string) => void;
}

export function NlWorkerCreate({ orgSlug, onNeedsWizard }: NlWorkerCreateProps) {
  const t = useTranslations();
  const router = useRouter();
  const [prompt, setPrompt] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const trimmed = prompt.trim();

  async function submit() {
    if (!trimmed || submitting) return;
    setSubmitting(true);
    try {
      const result = await quickTaskApi.create({ prompt: trimmed });
      toast.success(t("workers.create.nl.queued"), {
        description: getShortPodKey(result.pod_key),
      });
      router.push(`/${orgSlug}/workspace?pod=${encodeURIComponent(result.pod_key)}`);
    } catch (err) {
      const code = getApiErrorCode(err);
      if (code && WIZARD_FALLBACK_CODES.has(code)) {
        toast.info(t("workers.create.nl.fallback"));
        onNeedsWizard(trimmed);
      } else {
        toast.error(
          getLocalizedErrorMessage(err, t, t("workers.create.nl.error")),
        );
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="mb-8 rounded-xl bg-surface-raised p-5 ring-1 ring-border/35">
      <div className="mb-3 flex items-center gap-2">
        <Sparkles className="h-4 w-4 text-primary" />
        <h2 className="text-sm font-medium">{t("workers.create.nl.title")}</h2>
      </div>
      <Textarea
        value={prompt}
        onChange={(e) => setPrompt(e.target.value)}
        placeholder={t("workers.create.nl.placeholder")}
        rows={3}
        maxLength={10000}
        disabled={submitting}
        onKeyDown={(e) => {
          if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            e.preventDefault();
            void submit();
          }
        }}
      />
      <div className="mt-3 flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {t("workers.create.nl.hint")}
        </p>
        <Button type="button" onClick={() => void submit()} disabled={!trimmed || submitting}>
          {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t("workers.create.nl.submit")}
        </Button>
      </div>
    </section>
  );
}
