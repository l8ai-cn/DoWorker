import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
}

export function Logo({ className }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 400 400"
      className={cn("w-full h-full", className)}
      aria-hidden
    >
      <defs>
        <linearGradient id="am-logo-bg" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#FF6A24" />
          <stop offset="100%" stopColor="#FF5701" />
        </linearGradient>
      </defs>
      <rect width="400" height="400" rx="32" fill="url(#am-logo-bg)" />
      <g stroke="#FFFFFF" strokeWidth="22" strokeLinecap="round">
        <line x1="118" y1="118" x2="282" y2="118" />
        <line x1="118" y1="282" x2="282" y2="282" />
        <line x1="118" y1="118" x2="118" y2="282" />
        <line x1="282" y1="118" x2="282" y2="282" />
        <line x1="118" y1="118" x2="282" y2="282" />
        <line x1="282" y1="118" x2="118" y2="282" />
      </g>
      <circle cx="200" cy="200" r="34" fill="#FFB020" />
      <circle cx="118" cy="118" r="26" fill="#FFF1EA" />
      <circle cx="282" cy="118" r="26" fill="#FFFFFF" />
      <circle cx="118" cy="282" r="26" fill="#FFFFFF" />
      <circle cx="282" cy="282" r="26" fill="#FFF1EA" />
    </svg>
  );
}
