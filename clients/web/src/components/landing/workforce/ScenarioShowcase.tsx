'use client'

import { useId, useState } from 'react'
import { useTranslations } from 'next-intl'
import {
  workforceScenarios,
  type WorkforceScenario,
  type WorkforceScenarioId,
  type WorkforceTranslationKey,
} from './workforce-scenarios'
import { ScenarioShowcaseTabs, type ScenarioTabKey } from './ScenarioShowcaseTabs'

type ScenarioShowcaseKey =
  | WorkforceTranslationKey
  | ScenarioTabKey
  | `landing.workforce.showcase.${ShowcaseCopyKey}`

type Translate = (key: ScenarioShowcaseKey) => string
type ShowcaseCopyKey =
  | 'scenarioPicker'
  | 'goal'
  | 'workers'
  | 'workflow'
  | 'deliverable'
  | 'eyebrow'
  | 'title'
  | 'description'

function ScenarioPanel({
  instanceId,
  scenario,
  selected,
  t,
}: {
  instanceId: string
  scenario: WorkforceScenario
  selected: boolean
  t: Translate
}) {
  return (
    <div
      id={`${instanceId}-scenario-panel-${scenario.id}`}
      role="tabpanel"
      aria-labelledby={`${instanceId}-scenario-tab-${scenario.id}`}
      tabIndex={selected ? 0 : -1}
      hidden={!selected}
      className={`gap-10 border-l-2 border-[var(--azure-mint)] py-8 pl-5 sm:pl-8 lg:grid-cols-[0.72fr_1.28fr] lg:gap-16 lg:py-12 ${
        selected ? 'grid' : 'hidden'
      }`}
    >
      <div>
        <p className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
          {t('landing.workforce.showcase.goal')}
        </p>
        <h3 className="mt-4 max-w-md font-headline text-3xl font-bold tracking-[-0.03em] text-foreground sm:text-4xl">
          {t(scenario.goalKey)}
        </h3>
        <p className="mt-9 font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-text-muted)]">
          {t('landing.workforce.showcase.workers')}
        </p>
        <ul data-testid={selected ? 'scenario-workers' : undefined} className="mt-3 space-y-2">
          {scenario.workers.map((worker) => (
            <li key={worker} className="flex items-center gap-3 text-sm text-foreground">
              <span className="h-1.5 w-1.5 rounded-full bg-[var(--azure-mint)]" aria-hidden="true" />
              {t(worker)}
            </li>
          ))}
        </ul>
      </div>

      <div className="lg:pt-6">
        <p className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-text-muted)]">
          {t('landing.workforce.showcase.workflow')}
        </p>
        <ol
          data-testid={selected ? 'scenario-workflow' : undefined}
          className="mt-4 grid gap-3 sm:grid-cols-2"
        >
          {scenario.steps.map((step, index) => (
            <li
              key={step}
              className="grid grid-cols-[auto_1fr] items-center gap-3 border-b border-[var(--azure-outline-variant)] py-3 text-sm text-foreground"
            >
              <span className="font-mono text-[10px] text-[var(--azure-mint)]">
                {String(index + 1).padStart(2, '0')}
              </span>
              {t(step)}
            </li>
          ))}
        </ol>
        <div className="mt-8 bg-[var(--azure-bg-high)] px-5 py-5 sm:ml-10">
          <p className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
            {t('landing.workforce.showcase.deliverable')}
          </p>
          <p className="mt-2 text-base font-medium text-foreground">{t(scenario.deliverableKey)}</p>
        </div>
      </div>
    </div>
  )
}

export function ScenarioShowcase() {
  const t = useTranslations()
  const instanceId = useId()
  const [selectedId, setSelectedId] = useState<WorkforceScenarioId>('research')

  return (
    <section className="bg-[var(--azure-bg-deeper)] px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="mb-10 max-w-2xl lg:ml-[18%]">
          <p className="font-headline text-xs font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
            {t('landing.workforce.showcase.eyebrow')}
          </p>
          <h2 className="mt-4 font-headline text-3xl font-bold tracking-[-0.03em] text-foreground sm:text-5xl">
            {t('landing.workforce.showcase.title')}
          </h2>
          <p className="mt-5 text-base leading-relaxed text-[var(--azure-text-muted)]">
            {t('landing.workforce.showcase.description')}
          </p>
        </div>
        <ScenarioShowcaseTabs
          instanceId={instanceId}
          selectedId={selectedId}
          onSelect={setSelectedId}
          t={t}
        />
        {workforceScenarios.map((scenario) => (
          <ScenarioPanel
            key={scenario.id}
            instanceId={instanceId}
            scenario={scenario}
            selected={scenario.id === selectedId}
            t={t}
          />
        ))}
      </div>
    </section>
  )
}
