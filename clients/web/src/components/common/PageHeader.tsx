"use client";

import Link from "next/link";
import { LanguageSwitcher } from "@/components/i18n";
import { LightAuthButtons as AuthButtons } from "./LightAuthButtons";
import { Logo } from "./Logo";

export function PageHeader() {
  return (
    <header className="bg-surface/80 backdrop-blur-md panel-lift sticky top-0 z-40">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2 motion-interactive">
          <div className="w-8 h-8 rounded-lg overflow-hidden">
            <Logo />
          </div>
          <span className="text-xl font-bold">Do Worker</span>
        </Link>
        <div className="flex items-center gap-4">
          <LanguageSwitcher variant="icon" />
          <AuthButtons consoleVariant="primary" />
        </div>
      </div>
    </header>
  );
}
