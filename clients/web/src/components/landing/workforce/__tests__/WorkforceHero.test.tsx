import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { WorkforceHero } from '../WorkforceHero'

const labels = vi.hoisted(() => ({
  'landing.workforce.hero.watchTeam': 'Watch the team work',
  'landing.workforce.mission.console': 'Mission console',
  'landing.workforce.mission.pause': 'Pause mission',
  'landing.workforce.mission.replay': 'Replay mission',
  'landing.workforce.scenarios.research.steps.scope': 'Scope the question',
  'landing.workforce.scenarios.research.steps.gather': 'Gather evidence',
}))

vi.mock('next-intl', () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}))

describe('WorkforceHero', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn(() => ({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      })),
    })
    Element.prototype.scrollIntoView = vi.fn()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('restarts, scrolls to, and focuses the mission console', () => {
    vi.useFakeTimers()
    render(<WorkforceHero />)
    act(() => vi.advanceTimersByTime(1800))
    expect(screen.getByText('Gather evidence')).toHaveAttribute('aria-current', 'step')
    const consoleRegion = screen.getByRole('region', { name: 'Mission console' })

    fireEvent.click(screen.getByRole('button', { name: 'Watch the team work' }))

    expect(screen.getByText('Scope the question')).toHaveAttribute('aria-current', 'step')
    expect(Element.prototype.scrollIntoView).toHaveBeenCalledWith({ block: 'center' })
    expect(consoleRegion).toHaveFocus()
  })
})
