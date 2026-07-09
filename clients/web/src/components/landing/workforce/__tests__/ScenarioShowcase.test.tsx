import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ScenarioShowcase } from '../ScenarioShowcase'

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
