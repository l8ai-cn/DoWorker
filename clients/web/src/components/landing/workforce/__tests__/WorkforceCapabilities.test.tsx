import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { AgentLogos } from '../../AgentLogos'
import { TrustDeployment, type TrustDeploymentKey } from '../TrustDeployment'
import {
  WorkforceCapabilities,
  type WorkforceCapabilityKey,
} from '../WorkforceCapabilities'

const agentLabels = vi.hoisted(() => ({
  'landing.agentLogos.title': 'Works with your preferred agents',
  'landing.agentLogos.descriptions.anthropic': 'Anthropic',
  'landing.agentLogos.descriptions.openai': 'OpenAI',
  'landing.agentLogos.descriptions.google': 'Google',
  'landing.agentLogos.descriptions.openSource': 'Open source',
  'landing.agentLogos.descriptions.community': 'Community',
  'landing.agentLogos.descriptions.anysphere': 'Anysphere',
  'landing.agentLogos.descriptions.selfBuilt': 'Self-built',
  'landing.agentLogos.descriptions.yourOwn': 'Your own',
}))
const translationRequests = vi.hoisted(() => [] as string[])

vi.mock('next-intl', () => ({
  useTranslations: () => (key: keyof typeof agentLabels) => {
    translationRequests.push(key)
    if (!(key in agentLabels)) throw new Error(`Missing existing translation: ${key}`)
    return agentLabels[key]
  },
}))

const capabilityCopy = {
  'landing.workforce.capabilities.eyebrow': 'Platform capabilities',
  'landing.workforce.capabilities.title': 'Direct every layer of the work',
  'landing.workforce.capabilities.description': 'One system for coordinated work',
  'landing.workforce.capabilities.items.organize.title': 'Organize',
  'landing.workforce.capabilities.items.organize.description': 'Assign the work',
  'landing.workforce.capabilities.items.organize.primitives.roles': 'Specialized roles',
  'landing.workforce.capabilities.items.organize.primitives.tasks': 'Owned tasks',
  'landing.workforce.capabilities.items.organize.primitives.ownership': 'Clear ownership',
  'landing.workforce.capabilities.items.organize.primitives.context': 'Shared context',
  'landing.workforce.capabilities.items.observe.title': 'Observe',
  'landing.workforce.capabilities.items.observe.description': 'Follow progress',
  'landing.workforce.capabilities.items.observe.primitives.activity': 'Live activity',
  'landing.workforce.capabilities.items.observe.primitives.evidence': 'Evidence',
  'landing.workforce.capabilities.items.observe.primitives.status': 'Delivery status',
  'landing.workforce.capabilities.items.observe.primitives.deliverables': 'Deliverables',
  'landing.workforce.capabilities.items.control.title': 'Control',
  'landing.workforce.capabilities.items.control.description': 'Set boundaries',
  'landing.workforce.capabilities.items.control.primitives.permissions': 'Permissions',
  'landing.workforce.capabilities.items.control.primitives.checkpoints': 'Checkpoints',
  'landing.workforce.capabilities.items.control.primitives.credentials': 'Credentials',
  'landing.workforce.capabilities.items.control.primitives.audit': 'Audit history',
  'landing.workforce.capabilities.items.operate.title': 'Operate',
  'landing.workforce.capabilities.items.operate.description': 'Run repeatable work',
  'landing.workforce.capabilities.items.operate.primitives.execution': 'Self-hosted execution',
  'landing.workforce.capabilities.items.operate.primitives.schedules': 'Schedules',
  'landing.workforce.capabilities.items.operate.primitives.workflows': 'Reusable workflows',
} satisfies Record<WorkforceCapabilityKey, string>

const trustCopy = {
  'landing.workforce.trust.eyebrow': 'Deployment and trust',
  'landing.workforce.trust.title': 'Adopt an AI workforce on your terms',
  'landing.workforce.trust.description': 'Keep execution inside your controls',
  'landing.workforce.trust.safeguards.selfHosting.title': 'Self-hosted control',
  'landing.workforce.trust.safeguards.selfHosting.description': 'Run on your infrastructure',
  'landing.workforce.trust.safeguards.selfHosting.status': 'Controlled',
  'landing.workforce.trust.safeguards.workspaces.title': 'Isolated AgentPod workspaces',
  'landing.workforce.trust.safeguards.workspaces.description': 'Separate every assignment',
  'landing.workforce.trust.safeguards.workspaces.status': 'Isolated',
  'landing.workforce.trust.safeguards.credentials.title': 'Managed credentials',
  'landing.workforce.trust.safeguards.credentials.description': 'Scope access to the work',
  'landing.workforce.trust.safeguards.credentials.status': 'Scoped',
  'landing.workforce.trust.safeguards.audit.title': 'Audit history',
  'landing.workforce.trust.safeguards.audit.description': 'Retain inspectable history',
  'landing.workforce.trust.safeguards.audit.status': 'Recorded',
  'landing.workforce.trust.compatibility.title': 'Bring the agents your team trusts',
  'landing.workforce.trust.compatibility.description': 'Use one workforce control layer',
} satisfies Record<TrustDeploymentKey, string>

function strictTranslator<Key extends string>(copy: Record<Key, string>) {
  return (key: Key) => {
    if (!(key in copy)) throw new Error(`Missing workforce translation: ${key}`)
    return copy[key]
  }
}

describe('WorkforceCapabilities', () => {
  it('maps four outcomes to their product primitives', () => {
    render(<WorkforceCapabilities translate={strictTranslator(capabilityCopy)} />)

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
    for (const ordinal of screen.getAllByText(/^0[1-4]$/)) {
      expect(ordinal).toHaveAttribute('aria-hidden', 'true')
      expect(ordinal).toHaveClass('text-xs')
    }
  })
})

describe('TrustDeployment', () => {
  it('presents deployment safeguards before agent compatibility', () => {
    render(<TrustDeployment translate={strictTranslator(trustCopy)} />)

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
    for (const agent of [
      'Claude Code',
      'Codex CLI',
      'Gemini CLI',
      'Aider',
      'OpenCode',
      'Cursor CLI',
      'Loopal',
      'Custom Agent',
    ]) {
      expect(screen.getByText(agent)).toBeVisible()
    }
    for (const description of Object.values(agentLabels).slice(1)) {
      expect(screen.getByText(description)).toHaveClass('text-xs')
    }
    for (const status of ['Controlled', 'Isolated', 'Scoped', 'Recorded']) {
      expect(screen.getByText(status)).toHaveClass('text-xs')
    }
    for (const ordinal of screen.getAllByText(/^0[1-4]$/)) {
      expect(ordinal).toHaveAttribute('aria-hidden', 'true')
      expect(ordinal).toHaveClass('text-xs')
    }
  })
})

describe('AgentLogos', () => {
  it('keeps standalone rendering on the existing translation namespace', () => {
    translationRequests.length = 0

    render(<AgentLogos />)

    expect(screen.getByRole('heading', { name: 'Works with your preferred agents' })).toBeVisible()
    expect(translationRequests).toHaveLength(9)
    expect(translationRequests.every((key) => key.startsWith('landing.agentLogos.'))).toBe(true)
  })
})
