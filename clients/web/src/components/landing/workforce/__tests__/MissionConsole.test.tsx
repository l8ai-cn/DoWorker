import { act, fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { hydrateRoot } from 'react-dom/client'
import { renderToString } from 'react-dom/server'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MissionConsole } from '../MissionConsole'
import { workforceScenarios, type WorkforceScenario } from '../workforce-scenarios'

const labels: Record<string, string> = {
  'landing.workforce.mission.goal': 'Goal',
  'landing.workforce.mission.workers': 'Workers',
  'landing.workforce.mission.activity': 'Activity',
  'landing.workforce.mission.checkpoint': 'Human checkpoint',
  'landing.workforce.mission.checkpointDetail': 'Review required before delivery',
  'landing.workforce.mission.deliverable': 'Final deliverable',
  'landing.workforce.mission.pause': 'Pause mission',
  'landing.workforce.mission.resume': 'Resume mission',
  'landing.workforce.mission.nextStep': 'Next step',
  'landing.workforce.mission.replay': 'Replay mission',
  'landing.workforce.scenarioLabels.research': 'Research',
  'landing.workforce.scenarioLabels.content': 'Content',
  'landing.workforce.scenarioLabels.operations': 'Operations',
  'landing.workforce.scenarioLabels.sales': 'Sales',
  'landing.workforce.scenarioLabels.knowledge': 'Knowledge',
  'landing.workforce.scenarioLabels.product': 'Product',
  'landing.workforce.scenarios.research.goal': 'Map the market and recommend a direction',
  'landing.workforce.scenarios.research.workers.scout': 'Research Scout',
  'landing.workforce.scenarios.research.workers.analyst': 'Market Analyst',
  'landing.workforce.scenarios.research.workers.editor': 'Brief Editor',
  'landing.workforce.scenarios.research.steps.scope': 'Scope the question',
  'landing.workforce.scenarios.research.steps.gather': 'Gather evidence',
  'landing.workforce.scenarios.research.steps.synthesize': 'Synthesize findings',
  'landing.workforce.scenarios.research.steps.review': 'Review recommendation',
  'landing.workforce.scenarios.research.deliverable': 'Market direction brief',
  'landing.workforce.scenarios.content.goal': 'Create a campaign from one brief',
  'landing.workforce.scenarios.content.workers.strategist': 'Content Strategist',
  'landing.workforce.scenarios.content.workers.writer': 'Campaign Writer',
  'landing.workforce.scenarios.content.workers.editor': 'Content Editor',
  'landing.workforce.scenarios.content.steps.brief': 'Shape the brief',
  'landing.workforce.scenarios.content.steps.research': 'Research the audience',
  'landing.workforce.scenarios.content.steps.draft': 'Draft the campaign',
  'landing.workforce.scenarios.content.steps.approve': 'Approve the package',
  'landing.workforce.scenarios.content.deliverable': 'Campaign package',
}

const t = (key: string) => labels[key] ?? key
const researchScenario = workforceScenarios[0]

function ControlledConsole({ onScenarioChange = vi.fn() }: { onScenarioChange?: (id: string) => void }) {
  const [scenario, setScenario] = useState<WorkforceScenario>(researchScenario)

  const handleScenarioChange = (id: WorkforceScenario['id']) => {
    setScenario(workforceScenarios.find((candidate) => candidate.id === id) ?? researchScenario)
    onScenarioChange(id)
  }

  return <MissionConsole scenario={scenario} onScenarioChange={handleScenarioChange} t={t} />
}

function installMatchMedia(reducedMotion = false) {
  const addEventListener = vi.fn()
  const removeEventListener = vi.fn()
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    value: vi.fn(() => ({
      matches: reducedMotion,
      media: '(prefers-reduced-motion: reduce)',
      onchange: null,
      addEventListener,
      removeEventListener,
      dispatchEvent: vi.fn(),
    })),
  })
  return { addEventListener, removeEventListener }
}

describe('MissionConsole', () => {
  beforeEach(() => {
    installMatchMedia()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the research goal and exactly three worker roles', () => {
    render(<ControlledConsole />)

    expect(screen.getByText('Map the market and recommend a direction')).toBeInTheDocument()
    expect(screen.getByText('Research Scout')).toBeInTheDocument()
    expect(screen.getByText('Market Analyst')).toBeInTheDocument()
    expect(screen.getByText('Brief Editor')).toBeInTheDocument()
    expect(screen.getAllByTestId('mission-worker')).toHaveLength(3)
  })

  it('switches content and reports the selected scenario', () => {
    const onScenarioChange = vi.fn()
    render(<ControlledConsole onScenarioChange={onScenarioChange} />)

    fireEvent.click(screen.getByRole('button', { name: 'Content' }))

    expect(onScenarioChange).toHaveBeenCalledWith('content')
    expect(screen.getByText('Create a campaign from one brief')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Content' })).toHaveAttribute('aria-pressed', 'true')
    expect(
      screen.getAllByRole('button').filter((button) => button.hasAttribute('aria-pressed')),
    ).toHaveLength(6)
    expect(screen.getByRole('group', { name: 'landing.workforce.mission.scenarios' })).toBeInTheDocument()
  })

  it('replay returns the activity timeline to step zero', () => {
    vi.useFakeTimers()
    render(<ControlledConsole />)

    act(() => vi.advanceTimersByTime(1800))
    expect(screen.getByText('Gather evidence')).toHaveAttribute('aria-current', 'step')

    fireEvent.click(screen.getByRole('button', { name: 'Replay mission' }))

    expect(screen.getByText('Scope the question')).toHaveAttribute('aria-current', 'step')
  })

  it('exposes accessible pause and replay controls', () => {
    render(<ControlledConsole />)

    expect(screen.getByRole('button', { name: 'Pause mission' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Replay mission' })).toBeInTheDocument()
  })

  it('pauses autoplay after direct scenario interaction', () => {
    vi.useFakeTimers()
    render(<ControlledConsole />)

    fireEvent.click(screen.getByRole('button', { name: 'Content' }))
    act(() => vi.advanceTimersByTime(3600))

    expect(screen.getByText('Shape the brief')).toHaveAttribute('aria-current', 'step')
    expect(screen.getByRole('button', { name: 'Resume mission' })).toBeInTheDocument()
  })

  it('disables autoplay and cleans up media listeners for reduced motion', () => {
    vi.useFakeTimers()
    const listeners = installMatchMedia(true)
    const { unmount } = render(<ControlledConsole />)

    act(() => vi.advanceTimersByTime(3600))
    expect(screen.getByText('Scope the question')).toHaveAttribute('aria-current', 'step')
    expect(listeners.addEventListener).toHaveBeenCalledWith('change', expect.any(Function))

    unmount()
    expect(listeners.removeEventListener).toHaveBeenCalledWith('change', expect.any(Function))
  })

  it('hydrates with the same playback label before detecting reduced motion', async () => {
    const browserWindow = window
    Reflect.deleteProperty(globalThis, 'window')
    const serverMarkup = renderToString(<ControlledConsole />)
    Object.defineProperty(globalThis, 'window', { configurable: true, value: browserWindow })
    installMatchMedia(true)
    const container = document.createElement('div')
    container.innerHTML = serverMarkup
    document.body.appendChild(container)
    const recoverableErrors: unknown[] = []

    let root: ReturnType<typeof hydrateRoot>
    await act(async () => {
      root = hydrateRoot(container, <ControlledConsole />, {
        onRecoverableError: (error) => recoverableErrors.push(error),
      })
    })

    const hasReducedMotionLabel = container.querySelector('[aria-label="Next step"]') !== null
    act(() => root.unmount())
    container.remove()
    expect(recoverableErrors).toHaveLength(0)
    expect(hasReducedMotionLabel).toBe(true)
  })

  it('offers manual progression and keeps replay paused under reduced motion', () => {
    vi.useFakeTimers()
    installMatchMedia(true)
    render(<ControlledConsole />)

    fireEvent.click(screen.getByRole('button', { name: 'Next step' }))
    expect(screen.getByText('Gather evidence')).toHaveAttribute('aria-current', 'step')
    fireEvent.click(screen.getByRole('button', { name: 'Replay mission' }))

    expect(screen.getByText('Scope the question')).toHaveAttribute('aria-current', 'step')
    expect(screen.getByRole('button', { name: 'Next step' })).toBeInTheDocument()
    act(() => vi.advanceTimersByTime(3600))
    expect(screen.getByText('Scope the question')).toHaveAttribute('aria-current', 'step')
  })

  it('uses collision-safe labels and announces the current activity', () => {
    const { container } = render(
      <>
        <ControlledConsole />
        <ControlledConsole />
      </>,
    )

    const workerHeadings = Array.from(container.querySelectorAll('[id]'))
      .map((element) => element.id)
      .filter((id) => id.includes('mission-workers'))
    expect(new Set(workerHeadings).size).toBe(2)
    expect(screen.getAllByRole('status')).toHaveLength(2)
  })
})
