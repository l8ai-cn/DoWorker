'use client'

import { useRef, useState } from 'react'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { DemoVideoModal } from '../HeroSection/DemoVideoModal'
import { MissionConsole, type MissionConsoleHandle } from './MissionConsole'
import { WorkforceBackdrop } from './WorkforceBackdrop'
import {
  workforceScenarios,
  type WorkforceScenario,
  type WorkforceScenarioId,
} from './workforce-scenarios'

export function WorkforceHero() {
  const t = useTranslations()
  const missionConsoleRef = useRef<MissionConsoleHandle>(null)
  const [scenario, setScenario] = useState<WorkforceScenario>(workforceScenarios[0])
  const [videoOpen, setVideoOpen] = useState(false)

  const selectScenario = (id: WorkforceScenarioId) => {
    const nextScenario = workforceScenarios.find((candidate) => candidate.id === id)
    if (nextScenario) setScenario(nextScenario)
  }

  const focusMission = () => {
    missionConsoleRef.current?.restartAndFocus()
  }

  return (
    <section className="relative overflow-hidden bg-[var(--azure-bg-deeper)] px-4 pb-20 pt-28 sm:px-6 sm:pb-24 sm:pt-36 lg:min-h-[92vh] lg:px-8 lg:py-32">
      <WorkforceBackdrop />
      <div className="pointer-events-none absolute inset-y-0 left-[42%] hidden w-px bg-[var(--azure-outline-variant)]/50 lg:block" />
      <div className="relative mx-auto grid max-w-7xl items-center gap-14 lg:grid-cols-[0.78fr_1.22fr] lg:gap-20">
        <div className="max-w-xl lg:-translate-y-8">
          <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-[var(--azure-mint)]/30 bg-[var(--azure-mint)]/10 px-3 py-1.5 backdrop-blur-sm">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[var(--azure-mint)] opacity-60" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-[var(--azure-mint)]" />
            </span>
            <span className="font-headline text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
              {t('landing.workforce.hero.badge')}
            </span>
          </div>

          <h1 className="font-headline text-4xl font-bold leading-[0.98] tracking-[-0.04em] text-foreground sm:text-5xl lg:text-7xl">
            {t('landing.workforce.hero.titleLead')}
            <span className="mt-2 block bg-gradient-to-r from-[var(--azure-mint)] to-[var(--azure-cyan-soft)] bg-clip-text text-transparent">
              {t('landing.workforce.hero.titleEmphasis')}
            </span>
          </h1>
          <p className="mt-7 max-w-lg text-base font-light leading-relaxed text-[var(--azure-text-muted)] sm:text-lg">
            {t('landing.workforce.hero.description')}
          </p>

          <div className="mt-9 flex flex-col items-stretch gap-3 sm:flex-row sm:items-center">
            <button
              type="button"
              onClick={focusMission}
              className="min-h-12 rounded-full bg-[var(--azure-mint)] px-7 font-headline text-xs font-black uppercase tracking-[0.16em] text-[var(--azure-on-cyan)] shadow-[0_0_32px_rgba(20,184,166,0.35)] transition-transform motion-safe:hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--azure-mint)]"
            >
              {t('landing.workforce.hero.watchTeam')}
            </button>
            <button
              type="button"
              onClick={() => setVideoOpen(true)}
              className="min-h-12 rounded-full border border-[var(--azure-outline-variant)] px-7 font-headline text-xs font-bold uppercase tracking-[0.16em] text-foreground transition-colors hover:border-[var(--azure-mint)]/60 focus-visible:outline-2 focus-visible:outline-[var(--azure-cyan)]"
            >
              {t('landing.workforce.hero.fullDemo')}
            </button>
          </div>
          <Link
            href="/register"
            className="mt-4 inline-block font-headline text-xs font-bold uppercase tracking-[0.16em] text-[var(--azure-text-muted)] underline decoration-[var(--azure-outline)] underline-offset-8 transition-colors hover:text-[var(--azure-mint)]"
          >
            {t('landing.workforce.hero.startFree')}
          </Link>
        </div>

        <div id="mission" className="relative scroll-mt-28 lg:translate-y-8">
          <div className="pointer-events-none absolute -left-5 top-10 hidden h-24 w-px bg-[var(--azure-mint)] lg:block" />
          <div className="pointer-events-none absolute -inset-4 -z-10 rounded-[2rem] bg-[var(--azure-mint)]/5 blur-2xl" />
          <MissionConsole
            ref={missionConsoleRef}
            scenario={scenario}
            onScenarioChange={selectScenario}
            t={t}
          />
        </div>
      </div>

      <DemoVideoModal
        open={videoOpen}
        onClose={() => setVideoOpen(false)}
        iframeTitle={t('landing.demoVideo.iframeTitle')}
      />
    </section>
  )
}
