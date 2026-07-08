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
import { useCurrentOrg } from "@/stores/auth";
import { buildLoopAiGuidePrompt } from "./loop-ai-guide-prompt";

const WIZARD_FALLBACK_CODES = new Set(["NO_RUNNER_FOR_AGENT", "AGENT_NOT_FOUND"]);
const EXAMPLE_KEYS = ["aiGuideExample1", "aiGuideExample2", "aiGuideExample3"] as const;

interface LoopNlCreateProps {
  /** Prefills the manual form below when AI auto-assign fails or user switches down. */
  onNeedsWizard: (idea: string) => void;
}

export function LoopNlCreate({ onNeedsWizard }: LoopNlCreateProps) {
  const t = useTranslations();
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const [idea, setIdea] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const trimmed = idea.trim();

  async function submit() {
    if (!trimmed || submitting || !currentOrg) return;
    setSubmitting(true);
    try {
      const result = await quickTaskApi.create({
        prompt: buildLoopAiGuidePrompt(trimmed),
        alias: t("loops.aiGuidePodAlias"),
      });
      toast.success(t("loops.aiGuideStarted"), {
        description: t("loops.aiGuideStartedDesc"),
      });
      router.push(`/${currentOrg.slug}/workspace?pod=${encodeURIComponent(result.pod_key)}`);
    } catch (err) {
      const code = getApiErrorCode(err);
      if (code && WIZARD_FALLBACK_CODES.has(code)) {
        toast.info(t("loops.aiGuideFallback"));
        onNeedsWizard(trimmed);
      } else {
        toast.error(getLocalizedErrorMessage(err, t, t("loops.aiGuideError")));
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="rounded-xl bg-surface-raised p-5 ring-1 ring-border/35">
      <div className="mb-3 flex items-center gap-2">
        <Sparkles className="h-4 w-4 text-primary" />
        <h2 className="text-sm font-medium">{t("loops.aiGuideTitle")}</h2>
      </div>
      <p className="mb-3 text-xs text-muted-foreground">{t("loops.aiGuideDescription")}</p>
      <Textarea
        value={idea}
        onChange={(e) => setIdea(e.target.value)}
        placeholder={t("loops.aiGuidePlaceholder")}
        rows={3}
        maxLength={4000}
        disabled={submitting}
        onKeyDown={(e) => {
          if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            e.preventDefault();
            void submit();
          }
        }}
      />
      <div className="mt-2 flex flex-wrap gap-1.5">
        {EXAMPLE_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            disabled={submitting}
            onClick={() => setIdea(t(`loops.${key}`))}
            className="rounded-full border border-border/60 px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
          >
            {t(`loops.${key}`)}
          </button>
        ))}
      </div>
      <div className="mt-3 flex items-center justify-between gap-3">
        <p className="text-xs text-muted-foreground">{t("loops.aiGuideHint")}</p>
        <Button type="button" onClick={() => void submit()} disabled={!trimmed || submitting} className="shrink-0 gap-1.5">
          {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
          {t("loops.aiGuideStart")}
        </Button>
      </div>
    </section>
  );
}
