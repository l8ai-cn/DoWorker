import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
}

export function Logo({ className }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 512 512"
      className={cn("w-full h-full", className)}
      aria-hidden="true"
      focusable="false"
      shapeRendering="geometricPrecision"
    >
      <rect data-logo-background width="512" height="512" rx="88" fill="#0B0F14" />
      <path
        data-logo-module
        d="M120 80H292V224H80V120Z"
        fill="#F4F7F6"
      />
      <path
        data-logo-module
        d="M308 80H392L432 120V224H308Z"
        fill="#F4F7F6"
      />
      <path
        data-logo-module
        d="M80 240H216V432H120L80 392Z"
        fill="#F4F7F6"
      />
      <path
        data-logo-module
        data-logo-active-module
        d="M232 240H432V392L392 432H232Z"
        fill="#4FD1C5"
      />
    </svg>
  );
}
