"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import ArchitectureDiagram from "@/components/docs/ArchitectureDiagram";

const AGENTS = [
  "Claude Code (Anthropic)",
  "Codex CLI (OpenAI)",
  "Gemini CLI (Google)",
  "Aider",
  "OpenCode",
];

const CAPABILITIES = ["orchestrate", "remoteWorkstation", "taskDriven", "selfHosted"] as const;

const QUICK_LINKS = [
  { href: "/docs/getting-started", titleKey: "quickStart", descKey: "quickStartDesc" },
  { href: "/docs/features/agentpod", titleKey: "agentpod", descKey: "agentpodDesc" },
  { href: "/docs/features/channels", titleKey: "agentsmesh", descKey: "agentsmeshDesc" },
  { href: "/docs/runners/mcp-tools", titleKey: "mcpTools", descKey: "mcpToolsDesc" },
] as const;

export default function DocsPage() {
  const t = useTranslations();

  return (
    <div>
      <div className="mb-10 sm:mb-14">
        <span className="inline-flex items-center gap-2 rounded-full bg-primary/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
          <span className="h-1.5 w-1.5 rounded-full bg-primary" />
          {t("docs.title")}
        </span>
        <h1 className="mt-4 text-3xl sm:text-4xl md:text-5xl font-semibold leading-tight tracking-tight text-foreground">
          {t("docs.intro.title")}
        </h1>
        <p className="mt-4 max-w-2xl text-base sm:text-lg leading-relaxed text-muted-foreground">
          {t("docs.intro.description")}
        </p>
      </div>

      <section className="mb-12 sm:mb-16">
        <div className="surface-card rounded-2xl p-5 sm:p-7">
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-primary">
            {t("docs.intro.supportedAgents")}
          </p>
          <ul className="mt-4 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2">
            {AGENTS.map((agent) => (
              <li key={agent} className="flex items-center gap-2 text-sm text-foreground">
                <span className="h-1.5 w-1.5 rounded-full bg-success" />
                {agent}
              </li>
            ))}
            <li className="flex items-center gap-2 text-sm text-muted-foreground">
              <span className="h-1.5 w-1.5 rounded-full bg-success" />
              {t("docs.intro.customAgents")}
            </li>
          </ul>
        </div>
      </section>

      <section className="mb-12 sm:mb-16">
        <div className="mb-5 sm:mb-6">
          <h2 className="text-2xl sm:text-3xl font-semibold tracking-tight text-foreground">
            {t("docs.whatYouCanDo.title")}
          </h2>
          <p className="mt-2 text-muted-foreground leading-relaxed max-w-2xl">
            {t("docs.whatYouCanDo.description")}
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {CAPABILITIES.map((key) => (
            <div key={key} className="surface-card-interactive rounded-xl p-5 sm:p-6 motion-interactive">
              <h3 className="text-base font-semibold text-foreground">
                {t(`docs.whatYouCanDo.${key}.title`)}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {t(`docs.whatYouCanDo.${key}.description`)}
              </p>
            </div>
          ))}
        </div>
      </section>

      <section className="mb-12 sm:mb-16">
        <div className="mb-2">
          <h2 className="text-2xl sm:text-3xl font-semibold tracking-tight text-foreground">
            {t("docs.architecture.title")}
          </h2>
          <p className="mt-2 text-muted-foreground leading-relaxed max-w-2xl">
            {t("docs.architecture.description")}
          </p>
        </div>
        <ArchitectureDiagram />
      </section>

      <section className="mb-12">
        <h2 className="text-2xl sm:text-3xl font-semibold mb-5 sm:mb-6 tracking-tight text-foreground">
          {t("docs.quickLinks.title")}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {QUICK_LINKS.map(({ href, titleKey, descKey }) => (
            <QuickLinkCard
              key={href}
              href={href}
              title={`${t(`docs.quickLinks.${titleKey}`)} →`}
              description={t(`docs.quickLinks.${descKey}`)}
            />
          ))}
        </div>
      </section>
    </div>
  );
}

function QuickLinkCard({
  href,
  title,
  description,
}: {
  href: string;
  title: string;
  description: string;
}) {
  return (
    <Link href={href} className="surface-card-interactive rounded-xl p-5 sm:p-6 block group motion-interactive">
      <h3 className="text-base font-semibold text-foreground group-hover:text-primary transition-colors">
        {title}
      </h3>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{description}</p>
    </Link>
  );
}
