import { cn } from "@/lib/utils";

interface AgentCloudLogoProps {
  className?: string;
  /** Defaults to decorative; set for branded hero marks. */
  "aria-hidden"?: boolean | "true" | "false";
  title?: string;
}

export function AgentCloudLogo({ className, title, ...rest }: AgentCloudLogoProps) {
  const labelled = title != null && title !== "";
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 400 400"
      className={cn("h-full w-full text-primary", className)}
      role={labelled ? "img" : undefined}
      aria-hidden={labelled ? undefined : rest["aria-hidden"] ?? true}
      aria-label={labelled ? title : undefined}
    >
      <rect width="400" height="400" rx="32" fill="currentColor" />
      <g stroke="#FFFFFF" strokeWidth="22" strokeLinecap="round">
        <line x1="118" y1="118" x2="282" y2="118" />
        <line x1="118" y1="282" x2="282" y2="282" />
        <line x1="118" y1="118" x2="118" y2="282" />
        <line x1="282" y1="118" x2="282" y2="282" />
        <line x1="118" y1="118" x2="282" y2="282" />
        <line x1="282" y1="118" x2="118" y2="282" />
      </g>
      <circle cx="200" cy="200" r="34" fill="#5EEAD4" />
      <circle cx="118" cy="118" r="26" fill="#CCFBF1" />
      <circle cx="282" cy="118" r="26" fill="#FFFFFF" />
      <circle cx="118" cy="282" r="26" fill="#FFFFFF" />
      <circle cx="282" cy="282" r="26" fill="#CCFBF1" />
    </svg>
  );
}
