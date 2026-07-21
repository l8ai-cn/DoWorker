"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

export function PageFooter() {
  const t = useTranslations();

  return (
    <footer className="mt-16 bg-surface-muted/30 panel-lift">
      <div className="container mx-auto px-4 py-8">
        <div className="flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-sm text-muted-foreground">
            &copy; {new Date().getFullYear()} Agent Cloud.{" "}
            {t("common.allRightsReserved")}
          </p>
          <div className="flex gap-6">
            <Link
              href="/privacy"
              className="text-sm text-muted-foreground hover:text-primary motion-interactive"
            >
              {t("landing.footer.legal.privacy")}
            </Link>
            <Link
              href="/terms"
              className="text-sm text-muted-foreground hover:text-primary motion-interactive"
            >
              {t("landing.footer.legal.terms")}
            </Link>
            <Link
              href="/docs"
              className="text-sm text-muted-foreground hover:text-primary motion-interactive"
            >
              {t("landing.footer.resources.documentation")}
            </Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
