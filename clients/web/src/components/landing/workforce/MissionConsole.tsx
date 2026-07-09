'use client'

import { forwardRef, useImperativeHandle, useRef } from 'react'
import { type WorkforceScenario, type WorkforceScenarioId } from './workforce-scenarios'
import {
  ActivityTimeline,
  ScenarioControls,
  WorkerRoster,
  type MissionTranslate,
} from './MissionConsoleSections'
import { useMissionPlayback } from './useMissionPlayback'

interface MissionConsoleProps {
  scenario: WorkforceScenario
  onScenarioChange: (id: WorkforceScenarioId) => void
  t: MissionTranslate
}

export interface MissionConsoleHandle {
  restartAndFocus: () => void
}

export const MissionConsole = forwardRef<MissionConsoleHandle, MissionConsoleProps>(function MissionConsole(
  { scenario, onScenarioChange, t },
  ref,
) {
  const regionRef = useRef<HTMLDivElement>(null)
  const playback = useMissionPlayback(scenario.steps.length)
  const isDelivered = playback.currentStep === scenario.steps.length - 1
  const replay = playback.replay

  const selectScenario = (id: WorkforceScenarioId) => {
    playback.resetAndPause()
    onScenarioChange(id)
  }

  useImperativeHandle(ref, () => ({
    restartAndFocus() {
      replay()
      regionRef.current?.scrollIntoView({ block: 'center' })
      regionRef.current?.focus()
    },
  }), [replay])

  return (
    <div
      ref={regionRef}
      role="region"
      tabIndex={-1}
      aria-label={t('landing.workforce.mission.console')}
      className="relative rounded-[1.75rem] border border-[var(--azure-outline-variant)] bg-[var(--azure-bg-card)] p-4 shadow-[var(--shadow-soft)] sm:p-5"
    >
      <div className="mb-4 flex items-center justify-between border-b border-[var(--azure-outline-variant)] pb-3">
        <p className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
          {t('landing.workforce.mission.goal')}
        </p>
        <div className="flex items-center gap-1">
          {playback.reducedMotion ? (
            <button
              type="button"
              aria-label={t('landing.workforce.mission.nextStep')}
              onClick={playback.nextStep}
              className="min-h-10 rounded-lg border border-[var(--azure-outline-variant)] px-3 text-xs text-[var(--azure-text-muted)] hover:text-foreground focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)]"
            >
              {t('landing.workforce.mission.nextStep')}
            </button>
          ) : (
            <button
              type="button"
              aria-label={t(playback.isPaused ? 'landing.workforce.mission.resume' : 'landing.workforce.mission.pause')}
              onClick={playback.togglePlayback}
              className="min-h-10 rounded-lg border border-[var(--azure-outline-variant)] px-3 text-xs text-[var(--azure-text-muted)] hover:text-foreground focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)]"
            >
              {t(playback.isPaused ? 'landing.workforce.mission.resume' : 'landing.workforce.mission.pause')}
            </button>
          )}
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
})
