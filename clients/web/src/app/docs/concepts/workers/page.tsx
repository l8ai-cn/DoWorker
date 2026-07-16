"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import workerCatalog from "@/generated/worker-runtime-catalog.json";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { WorkerRuntimeCard } from "@/components/docs/WorkerRuntimeCard";

const CREATE_STEPS = ["runtime", "typeConfig", "workspace", "preflight"] as const;

export default function WorkersPage() {
  const t = useTranslations("docs.workerRuntime");
  const summary = summarizeCatalog();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-4">{t("title")}</h1>
      <p className="max-w-3xl text-muted-foreground leading-relaxed mb-8">
        {t("description")}
      </p>

      <section className="mb-10 rounded-xl border border-primary/20 bg-primary/5 p-5 sm:p-6">
        <h2 className="text-xl font-semibold text-foreground">{t("catalog.title")}</h2>
        <dl className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <CatalogStat label={t("catalog.defined")} value={summary.defined} />
          <CatalogStat label={t("catalog.deployable")} value={summary.deployable} />
          <CatalogStat label={t("catalog.localEvidence")} value={summary.localEvidence} />
          <CatalogStat label={t("catalog.releaseBlocked")} value={summary.releaseBlocked} />
        </dl>
        <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
          {t("catalog.note")}
        </p>
      </section>

      <section className="mb-10">
        <h2 className="text-2xl font-semibold text-foreground">{t("createFlow.title")}</h2>
        <p className="mt-2 max-w-3xl text-muted-foreground leading-relaxed">
          {t("createFlow.description")}
        </p>
        <div className="mt-5 grid gap-4 md:grid-cols-2">
          {CREATE_STEPS.map((step, index) => (
            <article key={step} className="surface-card rounded-xl p-5">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-primary">
                {t("createFlow.step", { number: index + 1 })}
              </p>
              <h3 className="mt-2 font-semibold text-foreground">
                {t(`createFlow.${step}.title`)}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {t(`createFlow.${step}.description`)}
              </p>
            </article>
          ))}
        </div>
      </section>

      <section className="mb-10">
        <h2 className="text-2xl font-semibold text-foreground">{t("truth.title")}</h2>
        <ul className="mt-4 space-y-3 text-muted-foreground leading-relaxed">
          <li>{t("truth.definition")}</li>
          <li>{t("truth.runtime")}</li>
          <li>{t("truth.evidence")}</li>
        </ul>
      </section>

      <section className="mb-10">
        <h2 className="text-2xl font-semibold text-foreground">{t("catalog.workersTitle")}</h2>
        <div className="mt-5 grid gap-4">
          {workerCatalog.workers.map((worker) => (
            <WorkerRuntimeCard key={worker.slug} worker={worker} />
          ))}
        </div>
      </section>

      <p className="mb-8 text-sm text-muted-foreground">
        <Link href="/docs/runners/setup" className="text-primary hover:underline">
          {t("runnerLink")}
        </Link>
      </p>
      <DocNavigation />
    </div>
  );
}

function CatalogStat({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
        {label}
      </dt>
      <dd className="mt-1 text-2xl font-semibold text-foreground">{value}</dd>
    </div>
  );
}

function summarizeCatalog() {
  return workerCatalog.workers.reduce(
    (summary, worker) => {
      summary.defined += 1;
      if (worker.validationStatus === "verified_local_dev") summary.deployable += 1;
      if (worker.validationStatus === "local_evidence_release_blocked") {
        summary.localEvidence += 1;
      }
      if (worker.validationStatus !== "verified_local_dev") summary.releaseBlocked += 1;
      return summary;
    },
    { defined: 0, deployable: 0, localEvidence: 0, releaseBlocked: 0 },
  );
}
