'use client'

import { useTranslations } from 'next-intl'
import { WorkforceHero } from './WorkforceHero'
import { ScenarioShowcase } from './ScenarioShowcase'
import { WorkLifecycle } from './WorkLifecycle'
import { WorkforceCapabilities } from './WorkforceCapabilities'
import { TrustDeployment } from './TrustDeployment'

export function WorkforceLanding() {
  const t = useTranslations()

  return (
    <>
      <WorkforceHero />
      <ScenarioShowcase />
      <WorkLifecycle />
      <WorkforceCapabilities translate={t} />
      <TrustDeployment translate={t} />
    </>
  )
}
