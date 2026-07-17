"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2 } from "lucide-react";

import { AlertMessage } from "@/components/ui/alert-message";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { submitExpertMarketRelease } from "@/lib/api/expertMarketApi";

interface ExpertMarketSubmissionDialogProps {
  expertSlug: string;
  marketSlug: string;
  marketSlugLocked?: boolean;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmitted: () => void;
}

function list(value: string): string[] {
  return [...new Set(value.split(/[,\n]/).map((item) => item.trim()).filter(Boolean))];
}

export function ExpertMarketSubmissionDialog({
  expertSlug,
  marketSlug,
  marketSlugLocked = false,
  open,
  onOpenChange,
  onSubmitted,
}: ExpertMarketSubmissionDialogProps) {
  const t = useTranslations("experts.marketSubmission");
  const [slug, setSlug] = useState(marketSlug);
  const [summary, setSummary] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("");
  const [tags, setTags] = useState("");
  const [outcomes, setOutcomes] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) setSlug(marketSlug);
  }, [marketSlug, open]);

  async function submit() {
    setSubmitting(true);
    setError("");
    try {
      await submitExpertMarketRelease(expertSlug, {
        slug: slug.trim(),
        summary: summary.trim(),
        description: description.trim(),
        category: category.trim(),
        icon: "rocket",
        tags: list(tags),
        outcomes: list(outcomes),
      });
      onOpenChange(false);
      onSubmitted();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setSubmitting(false);
    }
  }

  const disabled = submitting || !slug.trim() || !summary.trim() || !category.trim();
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-lg"
        title={t("dialogTitle")}
        description={t("dialogDescription")}
      >
        <div className="space-y-4 px-6 py-4">
          {error ? <AlertMessage type="error" message={error} /> : null}
          <Field label={t("slugLabel")} value={slug} onChange={setSlug}
            readOnly={marketSlugLocked} />
          <Field label={t("summaryLabel")} value={summary} onChange={setSummary} />
          <Field label={t("categoryLabel")} value={category} onChange={setCategory} />
          <div className="space-y-2">
            <Label htmlFor="market-description">{t("descriptionLabel")}</Label>
            <Textarea id="market-description" value={description}
              onChange={(event) => setDescription(event.target.value)} />
          </div>
          <Field label={t("tagsLabel")} value={tags} onChange={setTags}
            hint={t("tagsHint")} />
          <Field label={t("outcomesLabel")} value={outcomes} onChange={setOutcomes}
            hint={t("outcomesHint")} multiline />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>
            {t("cancel")}
          </Button>
          <Button onClick={submit} disabled={disabled}>
            {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
            {submitting ? t("submitting") : t("submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function Field({ label, value, onChange, hint, multiline, readOnly }: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  hint?: string;
  multiline?: boolean;
  readOnly?: boolean;
}) {
  const id = `market-${label.toLowerCase().replace(/\s+/g, "-")}`;
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      {multiline ? (
        <Textarea id={id} value={value}
          onChange={(event) => onChange(event.target.value)} />
      ) : (
        <Input id={id} value={value} readOnly={readOnly}
          onChange={(event) => onChange(event.target.value)} />
      )}
      {hint ? <p className="text-xs text-muted-foreground">{hint}</p> : null}
    </div>
  );
}
