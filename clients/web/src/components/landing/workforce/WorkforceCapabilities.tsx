type CapabilityId = 'organize' | 'observe' | 'control' | 'operate'
type PrimitiveByCapability = {
  organize: 'roles' | 'tasks' | 'ownership' | 'context'
  observe: 'activity' | 'evidence' | 'status' | 'deliverables'
  control: 'permissions' | 'checkpoints' | 'credentials' | 'audit'
  operate: 'execution' | 'schedules' | 'workflows'
}
export type WorkforceCapabilityKey =
  | `landing.workforce.capabilities.${'eyebrow' | 'title' | 'description'}`
  | `landing.workforce.capabilities.items.${CapabilityId}.${'title' | 'description'}`
  | {
      [Id in CapabilityId]: `landing.workforce.capabilities.items.${Id}.primitives.${PrimitiveByCapability[Id]}`
    }[CapabilityId]
export type WorkforceCapabilityTranslate = (key: WorkforceCapabilityKey) => string

const capabilityPrimitives: Record<CapabilityId, readonly WorkforceCapabilityKey[]> = {
  organize: [
    'landing.workforce.capabilities.items.organize.primitives.roles',
    'landing.workforce.capabilities.items.organize.primitives.tasks',
    'landing.workforce.capabilities.items.organize.primitives.ownership',
    'landing.workforce.capabilities.items.organize.primitives.context',
  ],
  observe: [
    'landing.workforce.capabilities.items.observe.primitives.activity',
    'landing.workforce.capabilities.items.observe.primitives.evidence',
    'landing.workforce.capabilities.items.observe.primitives.status',
    'landing.workforce.capabilities.items.observe.primitives.deliverables',
  ],
  control: [
    'landing.workforce.capabilities.items.control.primitives.permissions',
    'landing.workforce.capabilities.items.control.primitives.checkpoints',
    'landing.workforce.capabilities.items.control.primitives.credentials',
    'landing.workforce.capabilities.items.control.primitives.audit',
  ],
  operate: [
    'landing.workforce.capabilities.items.operate.primitives.execution',
    'landing.workforce.capabilities.items.operate.primitives.schedules',
    'landing.workforce.capabilities.items.operate.primitives.workflows',
  ],
}
const capabilities: readonly CapabilityId[] = ['organize', 'observe', 'control', 'operate']
const spans: Record<CapabilityId, string> = {
  organize: 'lg:col-span-7',
  observe: 'lg:col-span-5',
  control: 'lg:col-span-5',
  operate: 'lg:col-span-7',
}

function CapabilityFragment({
  id,
  translate,
}: {
  id: CapabilityId
  translate: WorkforceCapabilityTranslate
}) {
  return (
    <ul className="mt-6 border-t border-[var(--azure-outline-variant)] pt-4">
      {capabilityPrimitives[id].map((key, index) => (
        <li
          key={key}
          className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 border-b border-[var(--azure-outline-variant)] py-3 last:border-b-0"
        >
          <span
            aria-hidden="true"
            className={`h-2 w-2 rounded-full ${
              id === 'control' && index === 1 ? 'bg-warning' : 'bg-[var(--azure-mint)]'
            }`}
          />
          <span className="break-words text-sm font-medium text-foreground">
            {translate(key)}
          </span>
          <span
            aria-hidden="true"
            className="font-mono text-xs text-[var(--azure-text-muted)]"
          >
            {String(index + 1).padStart(2, '0')}
          </span>
        </li>
      ))}
    </ul>
  )
}

function CapabilityCard({
  id,
  index,
  translate,
}: {
  id: CapabilityId
  index: number
  translate: WorkforceCapabilityTranslate
}) {
  const title = translate(`landing.workforce.capabilities.items.${id}.title`)
  return (
    <article
      aria-label={title}
      className={`${spans[id]} min-w-0 border border-[var(--azure-outline-variant)] bg-[var(--azure-bg-high)]/45 p-6 sm:p-8`}
    >
      <div className="flex items-start justify-between gap-5">
        <div className="min-w-0">
          <h3 className="break-words font-headline text-2xl font-bold text-foreground">{title}</h3>
          <p className="mt-3 max-w-xl break-words text-sm leading-relaxed text-[var(--azure-text-muted)]">
            {translate(`landing.workforce.capabilities.items.${id}.description`)}
          </p>
        </div>
        <span
          aria-hidden="true"
          className="shrink-0 font-mono text-xs text-[var(--azure-mint)]"
        >
          {String(index + 1).padStart(2, '0')}
        </span>
      </div>
      <CapabilityFragment id={id} translate={translate} />
    </article>
  )
}

export function WorkforceCapabilities({
  translate,
}: {
  translate: WorkforceCapabilityTranslate
}) {
  return (
    <section className="bg-[var(--azure-bg-deeper)] px-4 py-20 sm:px-6 sm:py-24 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-6 lg:grid-cols-[0.7fr_1.3fr] lg:items-end">
          <p className="font-headline text-xs font-bold uppercase tracking-[0.2em] text-[var(--azure-mint)]">
            {translate('landing.workforce.capabilities.eyebrow')}
          </p>
          <div>
            <h2 className="break-words font-headline text-3xl font-bold tracking-[-0.03em] text-foreground sm:text-5xl">
              {translate('landing.workforce.capabilities.title')}
            </h2>
            <p className="mt-4 max-w-2xl break-words text-base leading-relaxed text-[var(--azure-text-muted)]">
              {translate('landing.workforce.capabilities.description')}
            </p>
          </div>
        </div>
        <div className="mt-14 grid gap-5 lg:grid-cols-12">
          {capabilities.map((id, index) => (
            <CapabilityCard key={id} id={id} index={index} translate={translate} />
          ))}
        </div>
      </div>
    </section>
  )
}
