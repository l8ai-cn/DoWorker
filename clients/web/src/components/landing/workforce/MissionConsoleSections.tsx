'use client'

import { useId } from 'react'
import { workforceScenarios, type WorkforceScenario, type WorkforceScenarioId } from './workforce-scenarios'
import type { WorkforceMessageKey } from './workforce-message-keys'

export type MissionTranslate = (key: WorkforceMessageKey) => string

export function ScenarioControls({
  selectedId,
  onSelect,
  t,
}: {
  selectedId: WorkforceScenarioId
  onSelect: (id: WorkforceScenarioId) => void
  t: MissionTranslate
}) {
  return (
    <div
      role="group"
      className="grid grid-cols-2 gap-1.5 sm:grid-cols-3"
      aria-label={t('landing.workforce.mission.scenarios')}
    >
      {workforceScenarios.map(({ id }) => (
        <button
          key={id}
          type="button"
          aria-pressed={id === selectedId}
          onClick={() => onSelect(id)}
          className="min-h-10 break-words rounded-lg border border-[var(--azure-outline-variant)] px-2 py-2 font-headline text-[10px] font-bold uppercase tracking-[0.1em] text-[var(--azure-text-muted)] transition-colors hover:border-[var(--azure-cyan)]/60 hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--azure-cyan)] aria-pressed:border-[var(--azure-mint)] aria-pressed:bg-[var(--azure-mint)]/10 aria-pressed:text-[var(--azure-mint)]"
        >
          {t(`landing.workforce.scenarioLabels.${id}`)}
        </button>
      ))}
    </div>
  )
}

export function WorkerRoster({ scenario, t }: { scenario: WorkforceScenario; t: MissionTranslate }) {
  const headingId = `${useId()}-mission-workers`

  return (
    <section aria-labelledby={headingId}>
      <h3
        id={headingId}
        className="mb-2 font-headline text-[10px] font-bold uppercase tracking-[0.18em] text-[var(--azure-text-muted)]"
      >
        {t('landing.workforce.mission.workers')}
      </h3>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
        {scenario.workers.map((worker, index) => (
          <div
            key={worker}
            data-testid="mission-worker"
            className="min-w-0 rounded-lg border border-[var(--azure-outline-variant)] bg-[var(--azure-bg-high)]/35 p-2.5"
          >
            <span className="mb-2 block h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" />
            <p className="break-words text-sm font-medium leading-relaxed text-foreground">{t(worker)}</p>
            <span className="mt-1 block text-[9px] text-[var(--azure-text-muted)]">
              {String(index + 1).padStart(2, '0')}
            </span>
          </div>
        ))}
      </div>
    </section>
  )
}

export function ActivityTimeline({
  scenario,
  currentStep,
  t,
}: {
  scenario: WorkforceScenario
  currentStep: number
  t: MissionTranslate
}) {
  const headingId = `${useId()}-mission-activity`

  return (
    <section aria-labelledby={headingId}>
      <div className="mb-2 flex items-center justify-between gap-3">
        <h3
          id={headingId}
          className="font-headline text-[10px] font-bold uppercase tracking-[0.18em] text-[var(--azure-text-muted)]"
        >
          {t('landing.workforce.mission.activity')}
        </h3>
        <span className="shrink-0 font-mono text-[10px] text-[var(--azure-mint)]">
          {currentStep + 1}/{scenario.steps.length}
        </span>
      </div>
      <ol className="space-y-1.5">
        {scenario.steps.map((step, index) => {
          const active = index === currentStep
          const complete = index < currentStep
          return (
            <li key={step} className="grid min-w-0 grid-cols-[1rem_minmax(0,1fr)] items-start gap-2 px-2 py-1.5">
              <span
                className={`mt-1.5 h-2 w-2 rounded-full border ${
                  active || complete
                    ? 'border-[var(--azure-mint)] bg-[var(--azure-mint)]'
                    : 'border-[var(--azure-outline)]'
                }`}
              />
              <p
                role={active ? 'status' : undefined}
                aria-current={active ? 'step' : undefined}
                aria-atomic={active ? true : undefined}
                className={`break-words text-sm leading-relaxed ${
                  active ? 'text-foreground' : 'text-[var(--azure-text-muted)]'
                }`}
              >
                {t(step)}
              </p>
            </li>
          )
        })}
      </ol>
    </section>
  )
}
