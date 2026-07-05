"use client";

import { useTranslations } from "next-intl";

export function BackgroundModes() {
  const t = useTranslations("docs.tutorials.runnerSetup.step3.background");

  return (
    <div className="mt-6">
      <h3 className="text-base font-semibold mb-3 text-[var(--azure-light-ink)]">
        {t("title")}
      </h3>
      <p className="text-sm text-muted-foreground mb-4">{t("description")}</p>
      <div className="space-y-4">
        <ModeCard
          badge="A"
          title={t("service.title")}
          description={t("service.description")}
          startLabel={t("service.startLabel")}
          startCode={`sudo do-worker-runner service install
sudo do-worker-runner service start
sudo do-worker-runner service status`}
          stopLabel={t("service.stopLabel")}
          stopCode={`sudo do-worker-runner service stop
# Remove from startup entirely:
sudo do-worker-runner service uninstall`}
          note={t("service.note")}
        />
        <ModeCard
          badge="B"
          title={t("nohup.title")}
          description={t("nohup.description")}
          startLabel={t("nohup.startLabel")}
          startCode={`nohup do-worker-runner run > ~/do-worker-runner.log 2>&1 &
echo $! > ~/.agentsmesh/runner.pid
tail -f ~/do-worker-runner.log`}
          stopLabel={t("nohup.stopLabel")}
          stopCode={`# Graceful stop by PID file
kill "$(cat ~/.agentsmesh/runner.pid)"

# Fallback if PID was lost
pkill -f do-worker-runner`}
          note={t("nohup.note")}
        />
      </div>
    </div>
  );
}

function ModeCard({
  badge,
  title,
  description,
  startLabel,
  startCode,
  stopLabel,
  stopCode,
  note,
}: {
  badge: string;
  title: string;
  description: string;
  startLabel: string;
  startCode: string;
  stopLabel: string;
  stopCode: string;
  note: string;
}) {
  return (
    <div className="azure-light-card rounded-xl p-6">
      <div className="flex items-center gap-3 mb-3">
        <span className="w-7 h-7 rounded-full azure-light-chip flex items-center justify-center text-xs font-bold">
          {badge}
        </span>
        <h4 className="text-base font-semibold text-[var(--azure-light-ink)]">
          {title}
        </h4>
      </div>
      <p className="text-sm text-[var(--azure-light-ink-muted)] leading-relaxed mb-4">
        {description}
      </p>
      <Block label={startLabel} code={startCode} />
      <div className="mt-4" />
      <Block label={stopLabel} code={stopCode} />
      <p className="mt-3 text-xs text-[var(--azure-light-ink-muted)] italic">
        {note}
      </p>
    </div>
  );
}

function Block({ label, code }: { label: string; code: string }) {
  return (
    <>
      <p className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-[var(--azure-light-ink-soft)]">
        {label}
      </p>
      <pre className="bg-[var(--azure-light-surface-high)] rounded-lg ring-1 ring-border/15 p-3 font-mono text-xs overflow-x-auto text-[var(--azure-light-cyan-ink)]">
        {code}
      </pre>
    </>
  );
}
