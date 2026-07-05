"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { Bug, FlaskConical, SearchCode, Sparkles, Terminal, X, type LucideIcon } from "lucide-react";
import {
  WORKSPACE_RECIPES,
  type WorkspaceRecipeSelection,
} from "@/components/workspace/workspace-recipes";

interface RecipeCardProps {
  recipeId: string;
  title: string;
  description: string;
  agents: string;
  duration: string;
  onClick?: () => void;
}

const RECIPE_ICONS: Record<string, LucideIcon> = {
  explain: SearchCode,
  tests: FlaskConical,
  bug: Bug,
};

function RecipeCard({ recipeId, title, description, agents, duration, onClick }: RecipeCardProps) {
  const Icon = RECIPE_ICONS[recipeId] ?? Terminal;

  return (
    <button
      type="button"
      onClick={onClick}
      className="group w-full rounded-lg bg-surface-raised p-3.5 text-left shadow-[var(--shadow-soft)] ring-1 ring-border/45 transition-all hover:-translate-y-0.5 hover:bg-card hover:ring-primary/25"
    >
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-md bg-surface-muted text-muted-foreground ring-1 ring-border/35 group-hover:text-primary">
            <Icon className="h-3.5 w-3.5" />
          </span>
          <span className="text-[13px] font-semibold text-foreground">{title}</span>
        </div>
        <p className="text-[11px] leading-4 text-muted-foreground">{description}</p>
        <div className="flex items-center gap-1.5 pt-1">
          <span className="rounded bg-surface-muted px-1.5 py-0.5 font-mono text-[10px] font-medium text-muted-foreground">
            {agents}
          </span>
          <span className="text-[10px] text-muted-foreground/70">· {duration}</span>
        </div>
      </div>
    </button>
  );
}

interface WorkspaceEmptyStateProps {
  onCreatePod: (recipe?: WorkspaceRecipeSelection) => void;
}

export function WorkspaceEmptyState({ onCreatePod }: WorkspaceEmptyStateProps) {
  const t = useTranslations();
  const [showBanner, setShowBanner] = useState(true);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "n") {
        const target = e.target as HTMLElement | null;
        const tag = target?.tagName;
        if (tag === "INPUT" || tag === "TEXTAREA" || target?.isContentEditable) return;
        e.preventDefault();
        onCreatePod();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onCreatePod]);

  return (
    <div className="flex h-full flex-col bg-background">
      {showBanner && (
        <div className="flex items-center gap-2.5 bg-[color-mix(in_srgb,var(--primary)_7%,var(--background))] px-6 py-2.5 text-[13px] shadow-[inset_0_-1px_0_color-mix(in_srgb,var(--border)_36%,transparent)]">
          <Sparkles className="h-3.5 w-3.5 text-primary" />
          <span className="font-medium text-foreground">{t("workspace.banner.newUser")}</span>
          <a href="#" className="font-medium text-primary hover:underline">
            {t("workspace.banner.watchIntro")}
          </a>
          <div className="flex-1" />
          <button
            type="button"
            onClick={() => setShowBanner(false)}
            className="text-muted-foreground hover:text-foreground"
            aria-label="Dismiss"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      )}

      <div className="flex flex-1 flex-col items-center justify-center gap-8 px-6 py-10">
        <div className="flex w-[520px] max-w-full flex-col items-center gap-4 text-center">
          <div className="flex h-[72px] w-[72px] items-center justify-center rounded-2xl bg-surface-raised shadow-[var(--shadow-soft)] ring-1 ring-border/60">
            <Terminal className="h-8 w-8 text-primary" />
          </div>
          <h1 className="text-2xl font-semibold text-foreground">
            {t("workspace.emptyHeroTitle")}
          </h1>
          <p className="max-w-[460px] text-sm leading-[22px] text-muted-foreground">
            {t("workspace.emptyHeroDescription")}
          </p>
          <div className="flex items-center gap-2.5 pt-3">
            <button
              type="button"
              onClick={onCreatePod}
              className="flex h-10 items-center gap-2 rounded-lg bg-primary px-5 text-sm font-semibold text-primary-foreground shadow-[0_8px_20px_color-mix(in_srgb,var(--primary)_18%,transparent)] hover:bg-primary-hover"
            >
              <span className="text-base leading-none">+</span>
              {t("workspace.createNewPod")}
            </button>
          </div>
        </div>

        <div className="flex w-[720px] max-w-full flex-col gap-2.5">
          <div className="text-center text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground/80">
            {t("workspace.recipesHeading")}
          </div>
          <div className="grid gap-3 sm:grid-cols-3">
            {WORKSPACE_RECIPES.map((recipe) => (
              <RecipeCard
                key={recipe.id}
                recipeId={recipe.id}
                title={t(`workspace.recipes.${recipe.id}.title`)}
                description={t(`workspace.recipes.${recipe.id}.description`)}
                agents={recipe.agentLabel}
                duration={t(`workspace.recipes.${recipe.id}.duration`)}
                onClick={() =>
                  onCreatePod({
                    agentSlug: recipe.agentSlug,
                    prompt: t(`workspace.recipes.${recipe.id}.prompt`),
                  })
                }
              />
            ))}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center gap-5 font-mono text-xs text-muted-foreground">
          <span>⌘K  {t("workspace.hints.search")}</span>
          <span>⌘N  {t("workspace.hints.createPod")}</span>
        </div>
      </div>
    </div>
  );
}
