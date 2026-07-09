'use client'

import { workforceScenarios, type WorkforceScenario, type WorkforceScenarioId } from './workforce-scenarios'
import { useMissionPlayback } from './useMissionPlayback'

type Translate = (key: string) => string

interface MissionConsoleProps {
  scenario: WorkforceScenario
  onScenarioChange: (id: WorkforceScenarioId) => void
  t: Translate
}

function ScenarioControls({
  selectedId,
  onSelect,
  t,
}: {
  selectedId: WorkforceScenarioId
  onSelect: (id: WorkforceScenarioId) => void
  t: Translate
}) {
  return (
    <div className="grid grid-cols-3 gap-1.5" aria-label={t('landing.workforce.mission.scenarios')}>
      {workforceScenarios.map(({ id }) => (
        <button
          key={id}
          type="button"
          aria-pressed={id === selectedId}
          onClick={() => onSelect(id)}
          className="min-h-10 rounded-lg border border-[var(--azure-outline-variant)] px-2 py-2 font-headline text-[10px] font-bold uppercase tracking-[0.12em] text-[var(--azure-text-muted)] transition-colors hover:border-[var(--azure-cyan)]/60 hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--azure-cyan)] aria-pressed:border-[var(--azure-mint)] aria-pressed:bg-[var(--azure-mint)]/10 aria-pressed:text-[var(--azure-mint)]"
        >
          {t(`landing.workforce.scenarioLabels.${id}`)}
        </button>
      ))}
    </div>
  )
}

function WorkerRoster({ scenario, t }: { scenario: WorkforceScenario; t: Translate }) {
  return (
    <section aria-labelledby="mission-workers">
      <h3
        id="mission-workers"
        className="mb-2 font-headline text-[10px] font-bold uppercase tracking-[0.18em] text-[var(--azure-text-muted)]"
      >
        {t('landing.workforce.mission.workers')}
      </h3>
      <div className="grid grid-cols-3 gap-2">
        {scenario.workers.map((worker, index) => (
          <div
            key={worker}
            data-testid="mission-worker"
            className="rounded-lg border border-[var(--azure-outline-variant)] bg-[var(--azure-bg-high)]/35 p-2.5"
          >
            <span className="mb-2 block h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" />
            <p className="text-xs font-medium leading-snug text-foreground">{t(worker)}</p>
            <span className="mt-1 block text-[9px] text-[var(--azure-text-muted)]">
              {String(index + 1).padStart(2, '0')}
            </span>
          </div>
        ))}
      </div>
    </section>
  )
}

function ActivityTimeline({
  scenario,
  currentStep,
  t,
}: {
  scenario: WorkforceScenario
  currentStep: number
  t: Translate
}) {
  return (
    <section aria-labelledby="mission-activity">
      <div className="mb-2 flex items-center justify-between">
        <h3
          id="mission-activity"
          className="font-headline text-[10px] font-bold uppercase tracking-[0.18em] text-[var(--azure-text-muted)]"
        >
          {t('landing.workforce.mission.activity')}
        </h3>
        <span className="font-mono text-[10px] text-[var(--azure-mint)]">
          {currentStep + 1}/{scenario.steps.length}
        </span>
      </div>
      <ol className="space-y-1.5">
        {scenario.steps.map((step, index) => {
          const active = index === currentStep
          const complete = index < currentStep
          return (
            <li key={step} className="grid grid-cols-[1rem_1fr] items-center gap-2 rounded-md px-2 py-1.5">
              <span
                className={`h-2 w-2 rounded-full border ${
                  active || complete
                    ? 'border-[var(--azure-mint)] bg-[var(--azure-mint)]'
                    : 'border-[var(--azure-outline)]'
                }`}
              />
              <p
                aria-current={active ? 'step' : undefined}
                className={active ? 'text-sm text-foreground' : 'text-sm text-[var(--azure-text-muted)]'}
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

export function MissionConsole({ scenario, onScenarioChange, t }: MissionConsoleProps) {
  const playback = useMissionPlayback(scenario.steps.length)
  const isDelivered = playback.currentStep === scenario.steps.length - 1

  const selectScenario = (id: WorkforceScenarioId) => {
    playback.resetAndPause()
    onScenarioChange(id)
  }

  return (
    <div
      id="mission-console"
      data-accent={scenario.accent}
      className="relative rounded-[1.75rem] border border-[var(--azure-outline-variant)] bg-[var(--azure-bg-card)] p-4 shadow-[var(--shadow-soft)] sm:p-5"
    >
      <div className="mb-4 flex items-center justify-between border-b border-[var(--azure-outline-variant)] pb-3">
        <p className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
          {t('landing.workforce.mission.goal')}
        </p>
        <div className="flex items-center gap-1">
          <button
            type="button"
            aria-label={t(playback.isPaused ? 'landing.workforce.mission.resume' : 'landing.workforce.mission.pause')}
            onClick={playback.togglePlayback}
            className="min-h-10 rounded-lg border border-[var(--azure-outline-variant)] px-3 text-xs text-[var(--azure-text-muted)] hover:text-foreground focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)]"
          >
            {t(playback.isPaused ? 'landing.workforce.mission.resume' : 'landing.workforce.mission.pause')}
          </button>
          <button
            type="button"
            aria-label={t('landing.workforce.mission.replay')}
            onClick={playback.replay}
            className="min-h-10 rounded-lg border border-[var(--azure-outline-variant)] px-3 text-xs text-[var(--azure-text-muted)] hover:text-foreground focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)]"
          >
            {t('landing.workforce.mission.replay')}
          </button>
        </div>
      </div>

      <p className="mb-4 max-w-md font-headline text-lg font-semibold leading-snug text-foreground">
        {t(scenario.goalKey)}
      </p>
      <div className="space-y-4">
        <ScenarioControls selectedId={scenario.id} onSelect={selectScenario} t={t} />
        <WorkerRoster scenario={scenario} t={t} />
        <ActivityTimeline scenario={scenario} currentStep={playback.currentStep} t={t} />
        <div className="rounded-lg border border-[var(--warning)]/40 bg-[var(--warning-bg)] p-3">
          <p className="font-headline text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--warning)]">
            {t('landing.workforce.mission.checkpoint')}
          </p>
          <p className="mt-1 text-xs text-[var(--azure-text-muted)]">
            {t('landing.workforce.mission.checkpointDetail')}
          </p>
        </div>
        <progress
          className="h-1.5 w-full accent-[var(--azure-mint)]"
          aria-label={t('landing.workforce.mission.progress')}
          value={playback.currentStep + 1}
          max={scenario.steps.length}
        />
        <div
          data-state={isDelivered ? 'delivered' : 'preparing'}
          className={`rounded-xl border p-3 transition-colors ${
            isDelivered
              ? 'border-[var(--azure-mint)]/50 bg-[var(--azure-mint)]/10'
              : 'border-[var(--azure-outline-variant)] bg-[var(--azure-bg-high)]/25'
          }`}
        >
          <p className="font-headline text-[10px] font-bold uppercase tracking-[0.16em] text-[var(--azure-text-muted)]">
            {t('landing.workforce.mission.deliverable')}
          </p>
          <p className="mt-1 text-sm font-semibold text-foreground">{t(scenario.deliverableKey)}</p>
        </div>
      </div>
    </div>
  )
}
