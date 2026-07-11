"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Search } from "lucide-react";

import { LightAuthButtons } from "@/components/common";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useLightSession } from "@/hooks/useLightSession";
import type { PublicMarketSkill } from "@/lib/public-market-api";
import { MarketplaceMessage, MarketplaceSkillCard } from "./MarketplaceSkillCard";
import { MarketplaceSignalPanel } from "./MarketplaceSignalPanel";

interface MarketplaceSkillBrowserProps {
  skills: PublicMarketSkill[];
  loadError?: string;
}

export function MarketplaceSkillBrowser({ skills, loadError }: MarketplaceSkillBrowserProps) {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("all");
  const { session } = useLightSession();

  const categories = useMemo(() => {
    const values = new Set(skills.map((skill) => skill.category || "workflow"));
    return ["all", ...Array.from(values).sort()];
  }, [skills]);

  const visibleSkills = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return skills.filter((skill) => {
      const matchesCategory = category === "all" || (skill.category || "workflow") === category;
      const haystack = `${skill.display_name} ${skill.slug} ${skill.description}`.toLowerCase();
      return matchesCategory && (!needle || haystack.includes(needle));
    });
  }, [skills, query, category]);

  const installHref = session?.currentOrgSlug ? `/${session.currentOrgSlug}/skills` : "/login";

  return (
    <main className="relative overflow-hidden pt-32">
      <section className="container mx-auto px-4 sm:px-6 lg:px-8 pb-12">
        <div className="grid gap-10 lg:grid-cols-[1.05fr_0.95fr] lg:items-end">
          <div>
            <Badge className="mb-5 border border-[var(--azure-cyan)]/30 bg-[var(--azure-cyan)]/10 text-[var(--azure-cyan)]">
              Public Skill Marketplace
            </Badge>
            <h1 className="font-headline text-5xl font-black tracking-tighter text-foreground sm:text-6xl lg:text-7xl">
              Reusable agent skills, ready for real work.
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-8 text-[var(--azure-text-muted)] sm:text-lg">
              Browse platform-published Skills that can be installed into a repository
              and mounted into Codex, Claude Code, DoAgent, or other workers.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link href={installHref}>
                <Button size="lg" className="rounded-full">Install in console</Button>
              </Link>
              <LightAuthButtons showRegister className="flex gap-3" />
            </div>
          </div>
          <MarketplaceSignalPanel total={skills.length} visible={visibleSkills.length} />
        </div>
      </section>

      <section className="container mx-auto px-4 sm:px-6 lg:px-8 pb-20">
        <div className="surface-card p-4 sm:p-5">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--azure-text-muted)]" />
              <Input
                id="marketplace-skill-search"
                name="marketplace-skill-search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search skills, workflows, and capabilities..."
                className="h-12 rounded-full pl-11"
              />
            </div>
            <div className="flex flex-wrap gap-2">
              {categories.map((item) => (
                <button
                  key={item}
                  type="button"
                  onClick={() => setCategory(item)}
                  className={`rounded-full border px-4 py-2 text-xs font-semibold uppercase tracking-[0.16em] transition ${
                    category === item
                      ? "border-[var(--azure-cyan)] bg-[var(--azure-cyan)]/15 text-[var(--azure-cyan)]"
                      : "border-border/70 text-[var(--azure-text-muted)] hover:border-[var(--azure-cyan)]/60 hover:text-foreground"
                  }`}
                >
                  {item}
                </button>
              ))}
            </div>
          </div>
        </div>

        {loadError ? (
          <MarketplaceMessage
            title="Marketplace is temporarily unavailable"
            description={loadError}
          />
        ) : visibleSkills.length === 0 ? (
          <MarketplaceMessage
            title="No matching skills"
            description="Try a different keyword or clear the category filter."
          />
        ) : (
          <div className="mt-8 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {visibleSkills.map((skill) => (
              <MarketplaceSkillCard key={skill.id} skill={skill} installHref={installHref} />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
