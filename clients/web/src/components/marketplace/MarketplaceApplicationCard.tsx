import {
  CheckCircle2,
  Clapperboard,
  Film,
  GitCompareArrows,
  Network,
  Rocket,
  Scissors,
  type LucideIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import type { PublicMarketApplication } from "@/lib/public-market-api";
import { MarketplaceInstallButton } from "./MarketplaceInstallButton";

const icons: Record<PublicMarketApplication["icon"], LucideIcon> = {
  rocket: Rocket,
  network: Network,
  "git-compare": GitCompareArrows,
  clapperboard: Clapperboard,
  scissors: Scissors,
  film: Film,
};

export function MarketplaceApplicationCard({
  application,
}: {
  application: PublicMarketApplication;
}) {
  const Icon = icons[application.icon];

  return (
    <article className="flex min-h-[420px] flex-col rounded-md border border-border bg-card p-6 shadow-sm transition hover:border-border-strong hover:shadow-md">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Icon className="h-5 w-5" />
          </span>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-lg font-semibold text-card-foreground">{application.name}</h2>
              {application.featured ? <Badge>精选</Badge> : null}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              {application.category} · {application.agent_slug}
            </p>
          </div>
        </div>
        <Badge variant="outline">v{application.version}</Badge>
      </div>

      <p className="mt-5 text-sm font-medium leading-6 text-card-foreground">
        {application.summary}
      </p>
      <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted-foreground">
        {application.description}
      </p>

      <div className="mt-5 space-y-2">
        {application.outcomes.map((outcome) => (
          <div key={outcome} className="flex items-start gap-2 text-sm text-card-foreground">
            <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-success" />
            <span>{outcome}</span>
          </div>
        ))}
      </div>

      <div className="mt-6 border-t border-border pt-4">
        <p className="text-xs font-medium text-muted-foreground">
          内置能力组件 · {application.skill_slugs.length} 项 Skills
        </p>
        <div className="mt-2 flex flex-wrap gap-2">
          {application.skill_slugs.map((skill) => (
            <Badge key={skill} variant="secondary" className="font-mono font-normal">
              {skill}
            </Badge>
          ))}
        </div>
      </div>

      <div className="mt-auto pt-6">
        <MarketplaceInstallButton
          applicationSlug={application.slug}
          agentSlug={application.agent_slug}
        />
      </div>
    </article>
  );
}
