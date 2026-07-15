"use client";

import type { ReactNode } from "react";
import { useTranslations } from "next-intl";

interface WorkerRuntimeCardProps {
  worker: {
    name: string;
    slug: string;
    executable: string;
    adapterId: string;
    interactionModes: string[];
    modelRequirement: {
      required: boolean;
      protocol_adapters: string[];
    };
    credentialBindings: Array<{
      sourceKind: string;
      sourceRef: string;
      environmentVariable: string;
    }>;
    configFields: Array<{
      name: string;
      kind: string;
      options: string[];
      defaultValue: string;
    }>;
    configDocuments: Array<{
      id: string;
      format: string;
      target_path: string;
    }>;
    runtimeImage: {
      name: string;
      reference: string;
      availability: string;
    } | null;
    validationStatus: string;
  };
}

export function WorkerRuntimeCard({ worker }: WorkerRuntimeCardProps) {
  const t = useTranslations("docs.workerRuntime");

  return (
    <article className="surface-card rounded-xl p-5 sm:p-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-foreground">{worker.name}</h2>
          <p className="mt-1 font-mono text-xs text-muted-foreground">{worker.slug}</p>
        </div>
        <span className={statusClass(worker.validationStatus)}>
          {t(`status.${worker.validationStatus}`)}
        </span>
      </header>

      <dl className="mt-5 grid gap-x-6 gap-y-4 text-sm sm:grid-cols-2">
        <Detail label={t("labels.executable")} value={<code>{worker.executable}</code>} />
        <Detail label={t("labels.adapter")} value={<code>{worker.adapterId}</code>} />
        <Detail
          label={t("labels.modes")}
          value={worker.interactionModes.map((mode) => (
            <code key={mode}>{mode}</code>
          ))}
        />
        <Detail
          label={t("labels.modelResource")}
          value={
            worker.modelRequirement.required
              ? t("model.required", {
                  protocols: worker.modelRequirement.protocol_adapters.join(", "),
                })
              : t("model.notRequired")
          }
        />
        <Detail
          label={t("labels.runtimeImage")}
          value={
            worker.runtimeImage
              ? `${worker.runtimeImage.name} (${t(`runtime.${worker.runtimeImage.availability}`)})`
              : t("runtime.blocked_no_published_digest")
          }
        />
        <Detail
          label={t("labels.credentials")}
          value={
            worker.credentialBindings.length === 0
              ? t("none")
              : worker.credentialBindings.map((binding) => (
                  <span key={`${binding.sourceKind}-${binding.environmentVariable}`}>
                    {t(`binding.${binding.sourceKind}`)} <code>{binding.sourceRef}</code>
                    {" -> "}
                    <code>{binding.environmentVariable}</code>
                  </span>
                ))
          }
        />
      </dl>

      <WorkerDetails
        title={t("labels.configFields")}
        emptyLabel={t("noConfigFields")}
        values={worker.configFields.map((field) => (
          <span key={field.name}>
            <code>{field.name}</code> <span className="text-muted-foreground">({field.kind})</span>
            {field.options.length > 0 && (
              <span className="text-muted-foreground">: {field.options.join(", ")}</span>
            )}
          </span>
        ))}
      />
      <WorkerDetails
        title={t("labels.configDocuments")}
        emptyLabel={t("noConfigDocuments")}
        values={worker.configDocuments.map((document) => (
          <span key={document.id}>
            <code>{document.id}</code> ({document.format}) {" -> "}
            <code>{document.target_path}</code>
          </span>
        ))}
      />
    </article>
  );
}

function Detail({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
        {label}
      </dt>
      <dd className="mt-1.5 flex flex-wrap gap-x-2 gap-y-1 text-foreground">{value}</dd>
    </div>
  );
}

function WorkerDetails({
  title,
  emptyLabel,
  values,
}: {
  title: string;
  emptyLabel: string;
  values: ReactNode[];
}) {
  return (
    <section className="mt-5 border-t border-border pt-4">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      {values.length === 0 ? (
        <p className="mt-2 text-sm text-muted-foreground">{emptyLabel}</p>
      ) : (
        <div className="mt-2 flex flex-wrap gap-2 text-sm text-foreground">
          {values.map((value, index) => (
            <span key={index} className="rounded-md bg-muted px-2 py-1">
              {value}
            </span>
          ))}
        </div>
      )}
    </section>
  );
}

function statusClass(status: string): string {
  const classes: Record<string, string> = {
    verified_local_dev: "rounded-full bg-success/15 px-2.5 py-1 text-xs font-semibold text-success",
    local_evidence_release_blocked: "rounded-full bg-warning/15 px-2.5 py-1 text-xs font-semibold text-warning",
    requires_model_resource: "rounded-full bg-warning/15 px-2.5 py-1 text-xs font-semibold text-warning",
    runtime_ready_unverified: "rounded-full bg-info/15 px-2.5 py-1 text-xs font-semibold text-info",
    invalid_published_runtime: "rounded-full bg-destructive/15 px-2.5 py-1 text-xs font-semibold text-destructive",
    runtime_image_unavailable: "rounded-full bg-muted px-2.5 py-1 text-xs font-semibold text-muted-foreground",
  };
  return classes[status] ?? classes.runtime_image_unavailable;
}
