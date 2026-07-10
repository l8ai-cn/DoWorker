import { AgentLogos } from '../AgentLogos'

type SafeguardId = 'selfHosting' | 'workspaces' | 'credentials' | 'audit'
export type TrustDeploymentKey =
  | `landing.workforce.trust.${'eyebrow' | 'title' | 'description'}`
  | `landing.workforce.trust.safeguards.${SafeguardId}.${'title' | 'description' | 'status'}`
  | `landing.workforce.trust.compatibility.${'title' | 'description'}`
export type TrustDeploymentTranslate = (key: TrustDeploymentKey) => string

const safeguards: readonly SafeguardId[] = [
  'selfHosting',
  'workspaces',
  'credentials',
  'audit',
]

function Safeguard({
  id,
  index,
  translate,
}: {
  id: SafeguardId
  index: number
  translate: TrustDeploymentTranslate
}) {
  return (
    <li className="grid min-w-0 grid-cols-[auto_minmax(0,1fr)] gap-4 border-t border-[var(--azure-outline-variant)] py-5">
      <span
        aria-hidden="true"
        className="mt-1 font-mono text-xs text-[var(--azure-mint)]"
      >
        {String(index + 1).padStart(2, '0')}
      </span>
      <div className="min-w-0">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <h3 className="break-words font-headline text-lg font-bold text-foreground">
            {translate(`landing.workforce.trust.safeguards.${id}.title`)}
          </h3>
          <span className="rounded-full border border-[var(--azure-outline)] px-2 py-0.5 font-mono text-xs text-[var(--azure-text-muted)]">
            {translate(`landing.workforce.trust.safeguards.${id}.status`)}
          </span>
        </div>
        <p className="mt-2 break-words text-sm leading-relaxed text-[var(--azure-text-muted)]">
          {translate(`landing.workforce.trust.safeguards.${id}.description`)}
        </p>
      </div>
    </li>
  )
}

export function TrustDeployment({ translate }: { translate: TrustDeploymentTranslate }) {
  return (
    <section className="bg-[var(--azure-bg-highest)] px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-10 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
          <div>
            <p className="font-headline text-xs font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
              {translate('landing.workforce.trust.eyebrow')}
            </p>
            <h2 className="mt-5 break-words font-headline text-3xl font-bold tracking-[-0.03em] text-foreground sm:text-5xl">
              {translate('landing.workforce.trust.title')}
            </h2>
            <p className="mt-5 max-w-xl break-words text-base leading-relaxed text-[var(--azure-text-muted)]">
              {translate('landing.workforce.trust.description')}
            </p>
          </div>
          <ol>
            {safeguards.map((id, index) => (
              <Safeguard key={id} id={id} index={index} translate={translate} />
            ))}
          </ol>
        </div>
        <div className="mt-14">
          <AgentLogos
            embedded
            title={translate('landing.workforce.trust.compatibility.title')}
            description={translate('landing.workforce.trust.compatibility.description')}
          />
        </div>
      </div>
    </section>
  )
}
