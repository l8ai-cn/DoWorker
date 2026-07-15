"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { Menu, X } from "lucide-react";
import { LanguageSwitcher } from "@/components/i18n";
import { LightAuthButtons as AuthButtons } from "@/components/common/LightAuthButtons";
import { Logo } from "@/components/common/Logo";
import { useTranslations } from "next-intl";
import { isMarketingRouteActive, marketingRoutes } from "./marketing-routes";

const GITHUB_URL = "https://github.com/l8ai-cn/DoWorker";

function GithubIcon() {
  return (
    <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
      <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
    </svg>
  );
}

export function Navbar() {
  const pathname = usePathname();
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const t = useTranslations();

  const navLinks = marketingRoutes.map(({ href, labelKey }) => ({
    href,
    label: t(labelKey),
  }));

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 10);
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const surfaceStyle = isMobileMenuOpen
    ? "bg-[var(--expert-bg)]"
    : isScrolled
      ? "bg-[var(--expert-bg)]/95 backdrop-blur-xl"
      : "bg-[var(--expert-bg)]/85 backdrop-blur-md";

  return (
    <nav className={`fixed left-0 right-0 top-0 z-50 border-b transition-colors ${surfaceStyle} ${isScrolled || isMobileMenuOpen ? "border-white/10" : "border-transparent"}`}>
      <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg overflow-hidden">
              <Logo />
            </div>
            <span className="text-lg font-semibold text-white">
              Do Worker
            </span>
          </Link>

          <div className="hidden items-center gap-5 xl:flex">
            {navLinks.map((link) => {
              const active = isMarketingRouteActive(pathname, link.href);
              const className = `border-b py-1 text-xs font-semibold transition-colors ${
                active
                  ? "border-[var(--expert-action)] text-white"
                  : "border-transparent text-[var(--expert-muted)] hover:border-[var(--expert-action)] hover:text-white"
              }`;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? "page" : undefined}
                  className={className}
                >
                  {link.label}
                </Link>
              );
            })}
          </div>

          <div className="hidden items-center gap-4 xl:flex">
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[var(--expert-muted)] transition-colors hover:text-white"
              aria-label="GitHub"
            >
              <GithubIcon />
            </a>
            <LanguageSwitcher variant="icon" />
            <AuthButtons size="sm" showRegister className="flex items-center gap-3" />
          </div>

          <button
            className="flex h-11 w-11 items-center justify-center text-[var(--expert-muted)] xl:hidden"
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            aria-label={t("landing.nav.toggleMenu")}
            aria-expanded={isMobileMenuOpen}
          >
            {isMobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>

        {isMobileMenuOpen && (
          <div className="mt-4 flex max-h-[calc(100dvh-5rem)] flex-col gap-1 overflow-y-auto border-t border-white/10 pt-3 xl:hidden">
            {navLinks.map((link) => {
              const active = isMarketingRouteActive(pathname, link.href);
              const className = `border-b border-white/8 py-3 text-sm font-semibold transition-colors ${
                active
                  ? "text-[var(--expert-action)]"
                  : "text-[var(--expert-muted)] hover:text-white"
              }`;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? "page" : undefined}
                  className={className}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {link.label}
                </Link>
              );
            })}
            <div className="mt-2 flex flex-col gap-3 pt-3">
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-[var(--expert-muted)] transition-colors hover:text-white"
              >
                <GithubIcon />
                GitHub
              </a>
              <div className="flex items-center justify-between py-1">
                <span className="text-sm text-[var(--expert-muted)]">{t("landing.nav.language")}</span>
                <LanguageSwitcher variant="full" />
              </div>
              <AuthButtons
                size="sm"
                showRegister
                onClick={() => setIsMobileMenuOpen(false)}
                className="flex flex-col gap-2 [&_button]:w-full"
              />
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}
