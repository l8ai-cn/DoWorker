"use client"

import { useId } from 'react'
import { useTranslations } from 'next-intl'

const agentConfigs = [
  {
    name: "Claude Code",
    descriptionKey: 'landing.agentLogos.descriptions.anthropic',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  {
    name: "Codex CLI",
    descriptionKey: 'landing.agentLogos.descriptions.openai',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2" fill="none" />
        <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  {
    name: "Gemini CLI",
    descriptionKey: 'landing.agentLogos.descriptions.google',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <polygon points="12,2 22,8.5 22,15.5 12,22 2,15.5 2,8.5" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  {
    name: "Aider",
    descriptionKey: 'landing.agentLogos.descriptions.openSource',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" strokeWidth="2" fill="none" />
        <path d="M9 9l6 6M15 9l-6 6" stroke="currentColor" strokeWidth="2" />
      </svg>
    ),
  },
  {
    name: "OpenCode",
    descriptionKey: 'landing.agentLogos.descriptions.community',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <path d="M12 2a10 10 0 110 20 10 10 0 010-20z" stroke="currentColor" strokeWidth="2" fill="none" />
        <path d="M8 12l2 2 4-4" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  {
    name: "Cursor CLI",
    descriptionKey: 'landing.agentLogos.descriptions.anysphere',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <path d="M5 3l14 7-6 2-2 6-6-15z" stroke="currentColor" strokeWidth="2" fill="none" strokeLinejoin="round" />
      </svg>
    ),
  },
  {
    name: "Loopal",
    descriptionKey: 'landing.agentLogos.descriptions.selfBuilt',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" fill="none" />
        <path d="M12 6l4 4-4 4-4-4z" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  },
  {
    name: "Custom Agent",
    descriptionKey: 'landing.agentLogos.descriptions.yourOwn',
    icon: (
      <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor" aria-hidden="true">
        <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2" />
      </svg>
    ),
  },
] as const

type AgentLogosProps =
  | { embedded?: false; title?: never; description?: never }
  | { embedded: true; title: string; description: string }

export function AgentLogos(props: AgentLogosProps = {}) {
  const t = useTranslations()
  const embedded = props.embedded === true
  const title = embedded ? props.title : t('landing.agentLogos.title')
  const headingId = `${useId()}-agent-compatibility`
  const Root = embedded ? 'div' : 'section'
  const Heading = embedded ? 'h3' : 'h2'

  return (
    <Root
      aria-labelledby={headingId}
      className={
        embedded
          ? 'border-t border-[var(--azure-outline-variant)] pt-10'
          : 'bg-[var(--azure-bg-deeper)] px-4 py-16 sm:px-6 lg:px-8'
      }
    >
      <div className={embedded ? '' : 'mx-auto max-w-7xl'}>
        <div className="max-w-2xl">
          <Heading
            id={headingId}
            className="break-words font-headline text-xl font-bold text-foreground sm:text-2xl"
          >
            {title}
          </Heading>
          {embedded ? (
            <p className="mt-2 break-words text-sm leading-relaxed text-[var(--azure-text-muted)]">
              {props.description}
            </p>
          ) : null}
        </div>
        <ul className="mt-6 grid grid-cols-1 gap-px bg-[var(--azure-outline-variant)] sm:grid-cols-2 lg:grid-cols-4">
          {agentConfigs.map((agent) => (
            <li
              key={agent.name}
              className="group flex min-w-0 items-center gap-3 bg-[var(--azure-bg-deeper)] px-4 py-4 transition-colors hover:bg-[var(--azure-bg-high)] focus-within:bg-[var(--azure-bg-high)] motion-reduce:transition-none"
            >
              <span className="shrink-0 text-[var(--azure-text-muted)] transition-colors group-hover:text-[var(--azure-mint)] motion-reduce:transition-none">
                {agent.icon}
              </span>
              <span className="min-w-0">
                <span className="block break-words font-headline text-sm font-semibold text-foreground">
                  {agent.name}
                </span>
                <span className="block break-words text-xs text-[var(--azure-text-muted)]">
                  {t(agent.descriptionKey)}
                </span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </Root>
  )
}
