"use client";

import Link from "next/link";
import { Logo } from "@/components/common";

interface AuthShellProps {
  title: string;
  subtitle: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
}

export function AuthShell({ title, subtitle, children, footer }: AuthShellProps) {
  return (
    <div className="auth-theme min-h-screen relative overflow-hidden bg-background flex items-center justify-center px-4 py-10">
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          backgroundImage:
            "linear-gradient(rgba(15, 23, 42, 0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(15, 23, 42, 0.04) 1px, transparent 1px)",
          backgroundSize: "96px 96px",
        }}
      />

      <div className="relative z-10 w-full max-w-md">
        <div className="text-center mb-7">
          <Link href="/" className="inline-flex items-center gap-2.5">
            <div className="w-9 h-9 rounded-lg overflow-hidden shadow-sm ring-1 ring-border/60">
              <Logo />
            </div>
            <span className="font-headline text-2xl font-semibold tracking-normal text-foreground">
              AgentsMesh
            </span>
          </Link>
        </div>

        <div className="rounded-xl bg-card p-7 shadow-[0_24px_70px_rgba(15,23,42,0.10)] ring-1 ring-border/80 sm:p-9">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-semibold text-foreground mb-2">{title}</h1>
            <p className="text-sm text-muted-foreground">{subtitle}</p>
          </div>
          {children}
        </div>

        {footer && (
          <p className="mt-6 text-center text-sm text-muted-foreground">{footer}</p>
        )}
      </div>
    </div>
  );
}
