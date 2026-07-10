import { useTranslations } from 'next-intl'

type StageId = 'goal' | 'coordinate' | 'review' | 'deliver'
type LifecycleKey =
  | `landing.workforce.lifecycle.${'eyebrow' | 'title' | 'description'}`
  | `landing.workforce.lifecycle.stages.${StageId}.${'title' | 'description'}`
  | `landing.workforce.lifecycle.fragments.ticket.${'id' | 'status' | 'title'}`
  | `landing.workforce.lifecycle.fragments.channel.${'name' | 'activity'}`
  | `landing.workforce.lifecycle.fragments.checkpoint.${'status' | 'decision' | 'action'}`
  | `landing.workforce.lifecycle.fragments.evidence.${'pod' | 'status' | 'artifact' | 'detail'}`
type Translate = (key: LifecycleKey) => string

const stages: readonly StageId[] = ['goal', 'coordinate', 'review', 'deliver']

function GoalFragment({ t }: { t: Translate }) {
  return (
    <div className="border-l-2 border-[var(--azure-mint)] pl-3">
      <div className="flex items-center justify-between gap-2">
        <span className="font-mono text-[10px] text-[var(--azure-mint)]">
          {t('landing.workforce.lifecycle.fragments.ticket.id')}
        </span>
        <span className="rounded-full border border-[var(--azure-outline)] px-2 py-0.5 text-[9px] text-[var(--azure-text-muted)]">
          {t('landing.workforce.lifecycle.fragments.ticket.status')}
        </span>
      </div>
      <p className="mt-3 text-xs font-medium leading-relaxed text-foreground">
        {t('landing.workforce.lifecycle.fragments.ticket.title')}
      </p>
    </div>
  )
}

function CoordinateFragment({ t }: { t: Translate }) {
  return (
    <div>
      <p className="font-mono text-[10px] text-[var(--azure-mint)]">
        {t('landing.workforce.lifecycle.fragments.channel.name')}
      </p>
      <div className="mt-3 flex items-center gap-2" aria-hidden="true">
        <span className="h-7 w-7 rounded-full border border-[var(--azure-mint)] bg-[var(--azure-mint)]/10" />
        <span className="h-px flex-1 bg-[var(--azure-outline)]" />
        <span className="h-7 w-7 rounded-full border border-[var(--azure-cyan)] bg-[var(--azure-cyan)]/10" />
        <span className="h-px flex-1 bg-[var(--azure-outline)]" />
        <span className="h-7 w-7 rounded-full border border-[var(--azure-outline)] bg-[var(--azure-bg-highest)]" />
      </div>
      <p className="mt-3 text-[10px] text-[var(--azure-text-muted)]">
        {t('landing.workforce.lifecycle.fragments.channel.activity')}
      </p>
    </div>
  )
}

function ReviewFragment({ t }: { t: Translate }) {
  return (
    <div>
      <div className="flex items-center gap-2 text-[10px] text-warning">
        <span className="h-2 w-2 rounded-full bg-warning" aria-hidden="true" />
        {t('landing.workforce.lifecycle.fragments.checkpoint.status')}
      </div>
      <p className="mt-3 text-xs font-medium leading-relaxed text-foreground">
        {t('landing.workforce.lifecycle.fragments.checkpoint.decision')}
      </p>
      <span className="mt-3 inline-block border-b border-warning pb-1 font-headline text-[9px] font-bold uppercase tracking-[0.14em] text-warning">
        {t('landing.workforce.lifecycle.fragments.checkpoint.action')}
      </span>
    </div>
  )
}

function DeliverFragment({ t }: { t: Translate }) {
  return (
    <div>
      <div className="flex items-center justify-between gap-2 font-mono text-[10px]">
        <span className="text-[var(--azure-mint)]">
          {t('landing.workforce.lifecycle.fragments.evidence.pod')}
        </span>
        <span className="text-[var(--azure-text-muted)]">
          {t('landing.workforce.lifecycle.fragments.evidence.status')}
        </span>
      </div>
      <div className="mt-3 border-t border-[var(--azure-outline-variant)] pt-3">
        <p className="text-xs font-medium text-foreground">
          {t('landing.workforce.lifecycle.fragments.evidence.artifact')}
        </p>
        <p className="mt-1 text-[10px] text-[var(--azure-text-muted)]">
          {t('landing.workforce.lifecycle.fragments.evidence.detail')}
        </p>
      </div>
    </div>
  )
}

const fragments: Record<StageId, (props: { t: Translate }) => React.ReactNode> = {
  goal: GoalFragment,
  coordinate: CoordinateFragment,
  review: ReviewFragment,
  deliver: DeliverFragment,
}

function LifecycleStage({ id, index, t }: { id: StageId; index: number; t: Translate }) {
  const Fragment = fragments[id]
  return (
    <li className="relative grid gap-5 border-t border-[var(--azure-outline-variant)] py-6 lg:grid-rows-[auto_1fr] lg:px-5">
      {index < stages.length - 1 ? (
        <span
          aria-hidden="true"
          className="absolute left-3 top-full z-10 h-6 w-px bg-[var(--azure-mint)] lg:left-full lg:top-9 lg:h-px lg:w-5"
        />
      ) : null}
      <div>
        <span className="font-mono text-[10px] text-[var(--azure-mint)]">
          {String(index + 1).padStart(2, '0')}
        </span>
        <h3 className="mt-3 font-headline text-xl font-bold text-foreground">
          {t(`landing.workforce.lifecycle.stages.${id}.title`)}
        </h3>
        <p className="mt-2 text-sm leading-relaxed text-[var(--azure-text-muted)]">
          {t(`landing.workforce.lifecycle.stages.${id}.description`)}
        </p>
      </div>
      <div className="self-end bg-[var(--azure-bg-high)]/60 p-4">
        <Fragment t={t} />
      </div>
    </li>
  )
}

export function WorkLifecycle() {
  const t = useTranslations()
  return (
    <section id="lifecycle" className="scroll-mt-24 bg-[var(--azure-bg-highest)]/40 px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-6 lg:grid-cols-[0.65fr_1.35fr] lg:items-end">
          <p className="font-headline text-xs font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
            {t('landing.workforce.lifecycle.eyebrow')}
          </p>
          <div>
            <h2 className="font-headline text-3xl font-bold tracking-[-0.03em] text-foreground sm:text-5xl">
              {t('landing.workforce.lifecycle.title')}
            </h2>
            <p className="mt-4 max-w-2xl text-base leading-relaxed text-[var(--azure-text-muted)]">
              {t('landing.workforce.lifecycle.description')}
            </p>
          </div>
        </div>
        <ol className="mt-14 grid gap-6 lg:grid-cols-4 lg:gap-5">
          {stages.map((id, index) => (
            <LifecycleStage key={id} id={id} index={index} t={t} />
          ))}
        </ol>
      </div>
    </section>
  )
}
