"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useExpertStore } from "@/stores/expert";
import { useCurrentOrg } from "@/stores/auth";
import { getPodDisplayName } from "@/lib/pod-display-name";
import type { Pod } from "@/stores/pod";

function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 100);
}

interface PublishExpertDialogProps {
  pod: Pod | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PublishExpertDialog({ pod, open, onOpenChange }: PublishExpertDialogProps) {
  const t = useTranslations("experts.publish");
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const publishFromPod = useExpertStore((s) => s.publishFromPod);

  const defaultName = pod ? getPodDisplayName(pod) : "";
  const [name, setName] = useState(defaultName);
  const [slug, setSlug] = useState(slugify(defaultName));
  const [submitting, setSubmitting] = useState(false);
  const [slugTouched, setSlugTouched] = useState(false);

  useEffect(() => {
    if (!open || !pod) return;
    const nextName = getPodDisplayName(pod);
    setName(nextName);
    setSlug(slugify(nextName));
    setSlugTouched(false);
  }, [open, pod]);

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugTouched) setSlug(slugify(value));
  };

  const handleSubmit = async () => {
    if (!pod || !name.trim() || !slug.trim()) return;
    setSubmitting(true);
    try {
      const expert = await publishFromPod(pod.pod_key, {
        name: name.trim(),
        slug: slug.trim(),
      });
      toast.success(t("success"));
      onOpenChange(false);
      if (currentOrg?.slug) {
        router.push(`/${currentOrg.slug}/experts/${expert.slug}`);
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="expert-name">{t("nameLabel")}</Label>
            <Input
              id="expert-name"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder={t("namePlaceholder")}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="expert-slug">{t("slugLabel")}</Label>
            <Input
              id="expert-slug"
              value={slug}
              onChange={(e) => {
                setSlugTouched(true);
                setSlug(e.target.value);
              }}
              placeholder={t("slugPlaceholder")}
            />
            <p className="text-xs text-muted-foreground">{t("slugHint")}</p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || !name.trim() || !slug.trim()}>
            {t("submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
