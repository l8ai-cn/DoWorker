"use client";

import { AuthShell } from "./AuthShell";
import { CenteredSpinner } from "@/components/ui/spinner";

type AuthStatusVariant = "loading" | "success" | "error" | "none";

interface AuthStatusShellProps {
  title: string;
  subtitle?: string;
  variant?: AuthStatusVariant;
  children?: React.ReactNode;
  footer?: React.ReactNode;
}

function StatusIcon({ variant }: { variant: Exclude<AuthStatusVariant, "none"> }) {
  if (variant === "loading") {
    return (
      <div className="flex justify-center">
        <CenteredSpinner size="lg" />
      </div>
    );
  }

  const isSuccess = variant === "success";
  return (
    <div className="flex justify-center">
      <div className={`w-16 h-16 rounded-full flex items-center justify-center ${isSuccess ? "bg-success-bg" : "bg-danger-bg"}`}>
        <svg className={`w-8 h-8 ${isSuccess ? "text-success" : "text-danger"}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          {isSuccess ? (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          ) : (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          )}
        </svg>
      </div>
    </div>
  );
}

export function AuthStatusShell({
  title,
  subtitle,
  variant = "none",
  children,
  footer,
}: AuthStatusShellProps) {
  return (
    <AuthShell title={title} subtitle={subtitle ?? ""} footer={footer}>
      {variant !== "none" && <StatusIcon variant={variant} />}
      {children}
    </AuthShell>
  );
}
