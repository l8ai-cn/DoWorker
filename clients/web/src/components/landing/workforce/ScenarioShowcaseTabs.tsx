import { useRef, type KeyboardEvent } from 'react'
import { workforceScenarios, type WorkforceScenarioId } from './workforce-scenarios'

export type ScenarioTabKey = `landing.workforce.scenarioLabels.${WorkforceScenarioId}`

type Props = {
  instanceId: string
  selectedId: WorkforceScenarioId
  onSelect: (id: WorkforceScenarioId) => void
  t: (key: ScenarioTabKey | 'landing.workforce.showcase.scenarioPicker') => string
}

export function ScenarioShowcaseTabs({ instanceId, selectedId, onSelect, t }: Props) {
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([])

  const moveSelection = (event: KeyboardEvent, index: number) => {
    let nextIndex: number
    if (event.key === 'ArrowRight') nextIndex = (index + 1) % workforceScenarios.length
    else if (event.key === 'ArrowLeft')
      nextIndex = (index - 1 + workforceScenarios.length) % workforceScenarios.length
    else if (event.key === 'Home') nextIndex = 0
    else if (event.key === 'End') nextIndex = workforceScenarios.length - 1
    else return

    event.preventDefault()
    onSelect(workforceScenarios[nextIndex].id)
    tabRefs.current[nextIndex]?.focus()
  }

  return (
    <div
      role="tablist"
      aria-label={t('landing.workforce.showcase.scenarioPicker')}
      className="grid grid-cols-2 gap-px border-y border-[var(--azure-outline-variant)] sm:grid-cols-3 lg:grid-cols-6"
    >
      {workforceScenarios.map(({ id }, index) => (
        <button
          key={id}
          ref={(node) => {
            tabRefs.current[index] = node
          }}
          id={`${instanceId}-scenario-tab-${id}`}
          type="button"
          role="tab"
          aria-controls={`${instanceId}-scenario-panel-${id}`}
          aria-selected={selectedId === id}
          tabIndex={selectedId === id ? 0 : -1}
          onClick={() => onSelect(id)}
          onKeyDown={(event) => moveSelection(event, index)}
          className="min-h-14 border-x border-[var(--azure-outline-variant)] px-3 py-3 text-left font-headline text-xs font-bold uppercase tracking-[0.12em] text-[var(--azure-text-muted)] transition-colors hover:text-foreground focus-visible:z-10 focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)] aria-selected:bg-[var(--azure-mint)] aria-selected:text-[var(--azure-on-cyan)]"
        >
          {t(`landing.workforce.scenarioLabels.${id}`)}
        </button>
      ))}
    </div>
  )
}
