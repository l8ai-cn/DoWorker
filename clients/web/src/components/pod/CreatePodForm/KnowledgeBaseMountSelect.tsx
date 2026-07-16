"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { BookOpen, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
  listKnowledgeBases,
  type KnowledgeBase,
  type KnowledgeMountSelection,
} from "@/lib/api/facade/knowledgeBaseApi";
interface KnowledgeBaseMountSelectProps {
  selectedMounts: KnowledgeMountSelection[];
  onChange: (mounts: KnowledgeMountSelection[]) => void;
  embedded?: boolean;
}

export function KnowledgeBaseMountSelect({
  selectedMounts,
  onChange,
  embedded = false,
}: KnowledgeBaseMountSelectProps) {
  const t = useTranslations();
  const params = useParams();
  const orgSlug = String(params.org ?? "");
  const [reloadToken, setReloadToken] = useState(0);
  const requestKey = `${orgSlug}:${reloadToken}`;
  const [loaded, setLoaded] = useState<{
    key: string;
    kbs: KnowledgeBase[];
    error: string | null;
  }>({ key: "", kbs: [], error: null });
  const loading = loaded.key !== requestKey;
  const kbs = loading ? [] : loaded.kbs;
  const error = loading ? null : loaded.error;

  useEffect(() => {
    let cancelled = false;
    listKnowledgeBases(orgSlug)
      .then((items) => {
        if (!cancelled) setLoaded({ key: requestKey, kbs: items, error: null });
      })
      .catch(() => {
        if (!cancelled) {
          setLoaded({
            key: requestKey,
            kbs: [],
            error: t("ide.createPod.knowledgeBasesLoadFailed"),
          });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, requestKey]);

  const mountOf = (kb: KnowledgeBase) => selectedMounts.find((mount) =>
    sameKnowledgeMount(mount, { id: kb.id, slug: kb.slug, mode: "ro" }),
  );

  const toggle = (kb: KnowledgeBase) => {
    const selected = mountOf(kb);
    onChange(
      selected
        ? selectedMounts.filter((mount) => !sameKnowledgeMount(mount, selected))
        : [...selectedMounts, { id: kb.id, slug: kb.slug, mode: "ro" }],
    );
  };

  const setMode = (selected: KnowledgeMountSelection, mode: "ro" | "rw") => {
    onChange(selectedMounts.map((mount) =>
      sameKnowledgeMount(mount, selected) ? { ...mount, mode } : mount,
    ));
  };

  return (
    <section>
      {!embedded && (
        <div className="mb-2 flex items-center justify-between gap-2">
          <label className="text-sm font-medium">{t("ide.createPod.knowledgeBases")}</label>
          <Link
            href={`/${orgSlug}/knowledge-base`}
            className="text-xs font-medium text-primary hover:underline"
          >
            {t("ide.createPod.manageKnowledgeBases")}
          </Link>
        </div>
      )}
      {embedded && (
        <div className="mb-2 flex justify-end">
          <Link
            href={`/${orgSlug}/knowledge-base`}
            className="text-xs font-medium text-primary hover:underline"
          >
            {t("ide.createPod.manageKnowledgeBases")}
          </Link>
        </div>
      )}

      {selectedMounts.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {selectedMounts.map((m) => (
            <span
              key={m.id ?? m.slug}
              className="inline-flex items-center gap-1 rounded-md border border-border bg-muted/30 px-2 py-0.5 text-xs"
            >
              <BookOpen className="h-3 w-3 text-primary" />
              <span className="max-w-[10rem] truncate" title={mountSlug(m, kbs)}>
                {mountSlug(m, kbs)}
              </span>
              <button
                type="button"
                className={`rounded px-1 font-mono text-[10px] font-semibold uppercase ${
                  m.mode === "rw"
                    ? "bg-primary/15 text-primary"
                    : "bg-muted text-muted-foreground"
                }`}
                onClick={() => setMode(m, m.mode === "rw" ? "ro" : "rw")}
                title={t("ide.createPod.knowledgeModeToggle")}
              >
                {m.mode === "rw"
                  ? t("ide.createPod.knowledgeModeReadWrite")
                  : t("ide.createPod.knowledgeModeReadOnly")}
              </button>
              <button
                type="button"
                className="text-muted-foreground hover:text-destructive"
                onClick={() => onChange(
                  selectedMounts.filter((mount) => !sameKnowledgeMount(mount, m)),
                )}
                aria-label={t("ide.createPod.removeKnowledgeBase")}
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}

      {loading ? (
        <div className="flex items-center py-2 text-sm text-muted-foreground">
          <Spinner size="sm" className="mr-2" />
          {t("ide.createPod.loadingKnowledgeBases")}
        </div>
      ) : error ? (
        <div role="alert" className="flex items-center justify-between gap-3 py-2">
          <p className="text-sm text-destructive">{error}</p>
          <Button type="button" variant="outline" size="sm" onClick={() => setReloadToken((value) => value + 1)}>
            {t("tickets.detail.retry")}
          </Button>
        </div>
      ) : kbs.length === 0 ? (
        <p className="py-2 text-xs text-muted-foreground">
          {t("ide.createPod.noKnowledgeBasesHint")}
        </p>
      ) : (
        <div className="surface-card max-h-40 overflow-y-auto">
          {kbs.map((kb) => (
            <label
              key={kb.slug}
              className="flex cursor-pointer items-center gap-2 border-b border-border px-2 py-1.5 last:border-b-0 hover:bg-surface-muted"
            >
              <input
                type="checkbox"
                className="h-3.5 w-3.5"
                checked={Boolean(mountOf(kb))}
                onChange={() => toggle(kb)}
              />
              <BookOpen className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="min-w-0 flex-1 truncate text-sm" title={kb.name}>
                {kb.name}
              </span>
              <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
                {kb.slug}
              </span>
            </label>
          ))}
        </div>
      )}
      <p className="mt-1 text-xs text-muted-foreground">
        {t("ide.createPod.knowledgeBasesHint")}
      </p>
    </section>
  );
}
function mountSlug(mount: KnowledgeMountSelection, knowledgeBases: KnowledgeBase[]): string {
  return mount.slug
    || knowledgeBases.find((item) => item.id === mount.id)?.slug
    || `#${mount.id}`;
}

function sameKnowledgeMount(left: KnowledgeMountSelection, right: KnowledgeMountSelection): boolean {
  if (left.id && right.id) return left.id === right.id;
  return left.slug !== "" && left.slug === right.slug;
}
