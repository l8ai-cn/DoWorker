import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { TrustDeployment } from '../TrustDeployment'
import { WorkforceCapabilities } from '../WorkforceCapabilities'

const labels = vi.hoisted(() => ({
  'landing.workforce.capabilities.title': 'Direct every layer of the work',
  'landing.workforce.capabilities.items.organize.title': 'Organize',
  'landing.workforce.capabilities.items.organize.primitives.roles': 'Specialized roles',
  'landing.workforce.capabilities.items.organize.primitives.tasks': 'Owned tasks',
  'landing.workforce.capabilities.items.organize.primitives.ownership': 'Clear ownership',
  'landing.workforce.capabilities.items.organize.primitives.context': 'Shared context',
  'landing.workforce.capabilities.items.observe.title': 'Observe',
  'landing.workforce.capabilities.items.observe.primitives.activity': 'Live activity',
  'landing.workforce.capabilities.items.observe.primitives.evidence': 'Evidence',
  'landing.workforce.capabilities.items.observe.primitives.status': 'Delivery status',
  'landing.workforce.capabilities.items.observe.primitives.deliverables': 'Deliverables',
  'landing.workforce.capabilities.items.control.title': 'Control',
  'landing.workforce.capabilities.items.control.primitives.permissions': 'Permissions',
  'landing.workforce.capabilities.items.control.primitives.checkpoints': 'Checkpoints',
  'landing.workforce.capabilities.items.control.primitives.credentials': 'Credentials',
  'landing.workforce.capabilities.items.control.primitives.audit': 'Audit history',
  'landing.workforce.capabilities.items.operate.title': 'Operate',
  'landing.workforce.capabilities.items.operate.primitives.execution': 'Self-hosted execution',
  'landing.workforce.capabilities.items.operate.primitives.schedules': 'Schedules',
  'landing.workforce.capabilities.items.operate.primitives.workflows': 'Reusable workflows',
  'landing.workforce.trust.title': 'Adopt an AI workforce on your terms',
  'landing.workforce.trust.safeguards.selfHosting.title': 'Self-hosted control',
  'landing.workforce.trust.safeguards.workspaces.title': 'Isolated AgentPod workspaces',
  'landing.workforce.trust.safeguards.credentials.title': 'Managed credentials',
  'landing.workforce.trust.safeguards.audit.title': 'Audit history',
  'landing.workforce.trust.compatibility.title': 'Bring the agents your team trusts',
}))

vi.mock('next-intl', () => ({
  useTranslations: () => (key: keyof typeof labels) => labels[key] ?? key,
}))

describe('WorkforceCapabilities', () => {
  it('maps four outcomes to their product primitives', () => {
    render(<WorkforceCapabilities />)

    expect(screen.getByRole('heading', { name: 'Direct every layer of the work' })).toBeVisible()
    const organize = screen.getByRole('article', { name: 'Organize' })
    expect(within(organize).getByText('Specialized roles')).toBeVisible()
    expect(within(organize).getByText('Owned tasks')).toBeVisible()
    expect(within(organize).getByText('Clear ownership')).toBeVisible()
    expect(within(organize).getByText('Shared context')).toBeVisible()

    const observe = screen.getByRole('article', { name: 'Observe' })
    expect(within(observe).getByText('Live activity')).toBeVisible()
    expect(within(observe).getByText('Evidence')).toBeVisible()
    expect(within(observe).getByText('Delivery status')).toBeVisible()
    expect(within(observe).getByText('Deliverables')).toBeVisible()

    const control = screen.getByRole('article', { name: 'Control' })
    expect(within(control).getByText('Permissions')).toBeVisible()
    expect(within(control).getByText('Checkpoints')).toBeVisible()
    expect(within(control).getByText('Credentials')).toBeVisible()
    expect(within(control).getByText('Audit history')).toBeVisible()

    const operate = screen.getByRole('article', { name: 'Operate' })
    expect(within(operate).getByText('Self-hosted execution')).toBeVisible()
    expect(within(operate).getByText('Schedules')).toBeVisible()
    expect(within(operate).getByText('Reusable workflows')).toBeVisible()
  })
})

describe('TrustDeployment', () => {
  it('presents deployment safeguards before agent compatibility', () => {
    render(<TrustDeployment />)

    expect(screen.getByRole('heading', { name: 'Adopt an AI workforce on your terms' })).toBeVisible()
    for (const safeguard of [
      'Self-hosted control',
      'Isolated AgentPod workspaces',
      'Managed credentials',
      'Audit history',
    ]) {
      expect(screen.getByRole('heading', { name: safeguard })).toBeVisible()
    }
    expect(screen.getByRole('heading', { name: 'Bring the agents your team trusts' })).toBeVisible()
    expect(screen.getByText('Claude Code')).toBeVisible()
    expect(screen.getByText('Codex CLI')).toBeVisible()
  })
})
