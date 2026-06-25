export function Divider({ text }: { text: string }) {
  return (
    <div className="my-6 space-y-4">
      <div className="soft-separator" />
      <p className="text-center text-[10px] font-headline tracking-[0.2em] uppercase text-[var(--azure-text-muted)]">
        {text}
      </p>
    </div>
  );
}
