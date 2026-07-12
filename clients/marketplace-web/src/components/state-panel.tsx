import { AlertCircle, Inbox, PauseCircle } from "lucide-react";
import Link from "next/link";

interface StatePanelProps {
  kind: "empty" | "error" | "suspended";
  title: string;
  description: string;
  action?: { href: string; label: string };
}

const icons = {
  empty: Inbox,
  error: AlertCircle,
  suspended: PauseCircle,
};

export function StatePanel({
  kind,
  title,
  description,
  action,
}: StatePanelProps) {
  const Icon = icons[kind];
  return (
    <section className={`state-panel state-${kind}`}>
      <span className="state-icon">
        <Icon aria-hidden="true" size={24} />
      </span>
      <h2>{title}</h2>
      <p>{description}</p>
      {action && (
        <Link className="button button-secondary" href={action.href}>
          {action.label}
        </Link>
      )}
    </section>
  );
}
