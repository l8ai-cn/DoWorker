import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { PublicMarketSkill } from "@/lib/public-market-api";

export function MarketplaceSkillCard({
  skill,
  installHref,
}: {
  skill: PublicMarketSkill;
  installHref: string;
}) {
  return (
    <article className="surface-card-interactive flex min-h-[260px] flex-col p-6">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-foreground">
            {skill.display_name || skill.slug}
          </h2>
          <p className="mt-1 font-mono text-xs text-[var(--azure-cyan)]">{skill.slug}</p>
        </div>
        <Badge variant="outline">v{skill.version}</Badge>
      </div>
      <p className="line-clamp-4 flex-1 text-sm leading-6 text-[var(--azure-text-muted)]">
        {skill.description || "No description provided yet."}
      </p>
      <div className="mt-5 flex flex-wrap gap-2">
        <Badge variant="secondary">{skill.category || "workflow"}</Badge>
        {skill.license ? <Badge variant="secondary">{skill.license}</Badge> : null}
      </div>
      <Link href={installHref} className="mt-6">
        <Button className="w-full rounded-full" variant="outline">
          Open install flow
        </Button>
      </Link>
    </article>
  );
}

export function MarketplaceMessage({ title, description }: { title: string; description: string }) {
  return (
    <div className="mt-8 surface-card p-10 text-center">
      <h2 className="text-2xl font-bold text-foreground">{title}</h2>
      <p className="mx-auto mt-3 max-w-xl text-sm leading-6 text-[var(--azure-text-muted)]">
        {description}
      </p>
    </div>
  );
}
