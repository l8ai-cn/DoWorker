import type { ReactNode } from "react";
import { AgentCloudLogo } from "@/components/icons/AgentCloudLogo";

interface AuthPageShellProps {
  title: string;
  description: string;
  children: ReactNode;
  footer?: ReactNode;
}

export function AuthPageShell({ title, description, children, footer }: AuthPageShellProps) {
  return (
    <div
      className="auth-shell flex min-h-dvh items-center justify-center px-4 py-8 text-foreground"
      style={{
        paddingTop: "max(2rem, var(--agent-cloud-safe-top))",
        paddingBottom: "max(2rem, var(--agent-cloud-safe-bottom))",
      }}
    >
      <div className="w-full max-w-[26rem]">
        <div className="rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8 dark:border-border/80 dark:bg-card-solid dark:shadow-md">
          <header className="mb-6 flex flex-col items-center gap-3 text-center">
            <AgentCloudLogo className="h-11 w-11 shrink-0" title="Agent Cloud" />
            <div className="space-y-1">
              <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
              <p className="text-sm text-muted-foreground">{description}</p>
            </div>
          </header>
          {children}
          {footer ? <div className="mt-6 border-t border-border/60 pt-5">{footer}</div> : null}
        </div>
      </div>
    </div>
  );
}
