import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ScenarioShowcase } from '../ScenarioShowcase'
import { WorkLifecycle } from '../WorkLifecycle'

const labels = vi.hoisted(() => ({
  'landing.workforce.scenarioLabels.research': 'Research',
  'landing.workforce.scenarioLabels.content': 'Content',
  'landing.workforce.scenarioLabels.operations': 'Operations',
  'landing.workforce.scenarioLabels.sales': 'Sales',
  'landing.workforce.scenarioLabels.knowledge': 'Knowledge',
  'landing.workforce.scenarioLabels.product': 'Product',
  'landing.workforce.scenarios.research.goal': 'Map the market',
  'landing.workforce.scenarios.research.workers.scout': 'Scout',
  'landing.workforce.scenarios.research.workers.analyst': 'Analyst',
  'landing.workforce.scenarios.research.workers.editor': 'Editor',
  'landing.workforce.scenarios.research.steps.scope': 'Scope',
  'landing.workforce.scenarios.research.steps.gather': 'Gather',
  'landing.workforce.scenarios.research.steps.synthesize': 'Synthesize',
  'landing.workforce.scenarios.research.steps.review': 'Review',
  'landing.workforce.scenarios.research.deliverable': 'Research brief',
  'landing.workforce.scenarios.sales.goal': 'Prepare an account plan',
  'landing.workforce.scenarios.sales.workers.researcher': 'Researcher',
  'landing.workforce.scenarios.sales.workers.strategist': 'Strategist',
  'landing.workforce.scenarios.sales.workers.writer': 'Writer',
  'landing.workforce.scenarios.sales.steps.target': 'Target',
  'landing.workforce.scenarios.sales.steps.qualify': 'Qualify',
  'landing.workforce.scenarios.sales.steps.personalize': 'Personalize',
  'landing.workforce.scenarios.sales.steps.review': 'Review outreach',
  'landing.workforce.scenarios.sales.deliverable': 'Sales account brief',
  'landing.workforce.lifecycle.stages.goal.title': 'Goal',
  'landing.workforce.lifecycle.stages.coordinate.title': 'Coordinate',
  'landing.workforce.lifecycle.stages.review.title': 'Review',
  'landing.workforce.lifecycle.stages.deliver.title': 'Deliver',
  'landing.workforce.lifecycle.fragments.checkpoint.status': 'Needs review',
  'landing.workforce.lifecycle.fragments.checkpoint.action': 'Review decision',
}))

vi.mock('next-intl', () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}))

describe('ScenarioShowcase', () => {
  it('offers all scenarios and starts with research selected', () => {
    render(<ScenarioShowcase />)

    expect(screen.getAllByRole('tab')).toHaveLength(6)
    expect(screen.getByRole('tab', { name: 'Research' })).toHaveAttribute('aria-selected', 'true')
  })

  it('connects every tab to a panel and hides inactive panels', () => {
    render(<ScenarioShowcase />)

    for (const tab of screen.getAllByRole('tab')) {
      const panelId = tab.getAttribute('aria-controls')
      expect(panelId).toBeTruthy()
      const panel = document.getElementById(panelId!)
      expect(panel).not.toBeNull()
      expect(panel).toHaveAttribute('role', 'tabpanel')
      expect(panel).toHaveAttribute('aria-labelledby', tab.id)
      expect(panel).toHaveProperty('hidden', tab.getAttribute('aria-selected') !== 'true')
      expect(panel).toHaveAttribute('tabindex', tab.getAttribute('aria-selected') === 'true' ? '0' : '-1')
    }
  })

  it('activates tabs with horizontal arrows and wraps at both ends', () => {
    render(<ScenarioShowcase />)
    const research = screen.getByRole('tab', { name: 'Research' })
    const content = screen.getByRole('tab', { name: 'Content' })
    const product = screen.getByRole('tab', { name: 'Product' })

    research.focus()
    fireEvent.keyDown(research, { key: 'ArrowLeft' })
    expect(product).toHaveFocus()
    expect(product).toHaveAttribute('aria-selected', 'true')

    fireEvent.keyDown(product, { key: 'ArrowRight' })
    expect(research).toHaveFocus()
    expect(research).toHaveAttribute('aria-selected', 'true')

    fireEvent.keyDown(research, { key: 'ArrowRight' })
    expect(content).toHaveFocus()
    expect(content).toHaveAttribute('aria-selected', 'true')
  })

  it('activates the first and last tabs with Home and End', () => {
    render(<ScenarioShowcase />)
    const research = screen.getByRole('tab', { name: 'Research' })
    const operations = screen.getByRole('tab', { name: 'Operations' })
    const product = screen.getByRole('tab', { name: 'Product' })

    operations.focus()
    fireEvent.keyDown(operations, { key: 'End' })
    expect(product).toHaveFocus()
    expect(product).toHaveAttribute('aria-selected', 'true')

    fireEvent.keyDown(product, { key: 'Home' })
    expect(research).toHaveFocus()
    expect(research).toHaveAttribute('aria-selected', 'true')
  })

  it('leaves vertical arrows available for page scrolling', () => {
    render(<ScenarioShowcase />)
    const research = screen.getByRole('tab', { name: 'Research' })
    research.focus()

    expect(fireEvent.keyDown(research, { key: 'ArrowUp' })).toBe(true)
    expect(fireEvent.keyDown(research, { key: 'ArrowDown' })).toBe(true)
    expect(research).toHaveFocus()
    expect(research).toHaveAttribute('aria-selected', 'true')
  })

  it('creates collision-safe relationships for multiple showcases', () => {
    render(
      <>
        <ScenarioShowcase />
        <ScenarioShowcase />
      </>,
    )
    const researchTabs = screen.getAllByRole('tab', { name: 'Research' })

    expect(researchTabs[0].id).not.toBe(researchTabs[1].id)
    for (const tab of researchTabs) {
      expect(document.getElementById(tab.getAttribute('aria-controls')!)).not.toBeNull()
    }
  })

  it('updates the scenario story when sales is selected', () => {
    render(<ScenarioShowcase />)

    fireEvent.click(screen.getByRole('tab', { name: 'Sales' }))

    expect(screen.getByRole('tab', { name: 'Sales' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByText('Prepare an account plan')).toBeVisible()
    expect(within(screen.getByTestId('scenario-workers')).getAllByRole('listitem')).toHaveLength(3)
    expect(within(screen.getByTestId('scenario-workflow')).getAllByRole('listitem')).toHaveLength(4)
    expect(screen.getByText('Sales account brief')).toBeVisible()
  })
})

describe('WorkLifecycle', () => {
  it('renders four stages and marks the human checkpoint as a warning', () => {
    render(<WorkLifecycle />)

    expect(screen.getAllByRole('listitem')).toHaveLength(4)
    expect(screen.getByRole('heading', { name: 'Goal' })).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Coordinate' })).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Review' })).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Deliver' })).toBeVisible()
    expect(screen.getByText('Needs review')).toHaveClass('text-warning')
    expect(screen.getByText('Review decision')).toHaveClass('text-warning', 'border-warning')
  })
})
