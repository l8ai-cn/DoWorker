export function WorkforceBackdrop() {
  return (
    <>
      <div
        aria-hidden="true"
        className="workforce-grid pointer-events-none absolute inset-0 opacity-[0.35]"
      />
      <div
        aria-hidden="true"
        className="workforce-orb pointer-events-none absolute -right-24 top-16 h-[28rem] w-[28rem] rounded-full bg-[var(--azure-mint)]/12 blur-[120px]"
      />
      <div
        aria-hidden="true"
        className="workforce-orb workforce-orb-delayed pointer-events-none absolute -left-16 bottom-0 h-80 w-80 rounded-full bg-[var(--azure-cyan)]/10 blur-[100px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-[var(--azure-mint)]/40 to-transparent"
      />
    </>
  )
}
