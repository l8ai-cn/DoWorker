export function WorkspaceViewTab({
  active,
  icon,
  id,
  label,
  onClick,
  panelId,
}: {
  active: boolean;
  icon: React.ReactNode;
  id: string;
  label: string;
  onClick: () => void;
  panelId: string;
}) {
  return (
    <button
      aria-controls={panelId}
      aria-selected={active}
      className={`flex h-11 items-center gap-1.5 rounded-md px-3 text-xs ${
        active ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted/60"
      }`}
      id={id}
      onClick={onClick}
      role="tab"
      tabIndex={active ? 0 : -1}
      type="button"
    >
      {icon}
      {label}
    </button>
  );
}
