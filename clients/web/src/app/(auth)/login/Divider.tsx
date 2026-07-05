export function Divider({ text }: { text: string }) {
  return (
    <div className="my-6 space-y-4">
      <div className="soft-separator" />
      <p className="text-center text-[10px] font-medium uppercase tracking-[0.16em] text-muted-foreground">
        {text}
      </p>
    </div>
  );
}
