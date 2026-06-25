import { useTranslations } from "next-intl";
import { ExternalLink as ExternalIcon } from "lucide-react";
import type { ReleaseSummary } from "@/lib/download/asset-types";

interface Props {
  release: ReleaseSummary;
}

const DATE_FMT: Intl.DateTimeFormatOptions = { year: "numeric", month: "short", day: "numeric" };

function formatReleaseDate(iso: string): string {
  return iso ? new Date(iso).toLocaleDateString(undefined, DATE_FMT) : "";
}

export function DownloadHero({ release }: Props) {
  const t = useTranslations();
  const releaseDate = formatReleaseDate(release.publishedAt);

  return (
    <section className="relative pt-36 pb-16 px-4 overflow-hidden">
      <div className="absolute inset-0 azure-mesh-bg pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[900px] h-[900px] bg-[var(--azure-cyan)]/10 blur-[180px] rounded-full pointer-events-none" />

      <div className="container mx-auto max-w-4xl text-center relative z-10">
        <div className="inline-flex items-center gap-2 px-4 py-1.5 mb-6 rounded-full border border-[var(--azure-cyan)]/30 bg-[var(--azure-cyan)]/5 text-[var(--azure-cyan)]">
          <span className="w-1.5 h-1.5 rounded-full bg-[var(--azure-cyan)] animate-pulse" />
          <span className="text-[10px] font-headline uppercase tracking-[0.25em] font-semibold">
            {t("landing.download.hero.badge")} v{release.version}
          </span>
        </div>

        <h1 className="font-headline text-5xl md:text-6xl font-bold mb-6 leading-[1.05] tracking-tight">
          {t("landing.download.hero.title")}{" "}
          <span className="azure-gradient-text">{t("landing.download.hero.titleHighlight")}</span>
        </h1>
        <p className="text-lg md:text-xl text-[var(--azure-text-muted)] max-w-2xl mx-auto mb-10 font-light">
          {t("landing.download.hero.subtitle")}
        </p>

        <p className="mb-6 text-xs text-[var(--azure-text-muted)]/70 uppercase tracking-[0.18em]">
          {t("landing.download.hero.releasedOn", { date: releaseDate })}
        </p>

        <a
          href={release.htmlUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 text-sm text-[var(--azure-text-muted)] hover:text-[var(--azure-cyan)] transition-colors"
        >
          {t("landing.download.hero.viewReleaseNotes")}
          <ExternalIcon className="w-4 h-4" />
        </a>
      </div>
    </section>
  );
}
