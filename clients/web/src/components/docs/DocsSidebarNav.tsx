import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { docsNavSections } from "@/lib/docs-navigation";
import { cn } from "@/lib/utils";

interface DocsSidebarNavProps {
  onNavigate?: () => void;
}

export function DocsSidebarNav({ onNavigate }: DocsSidebarNavProps) {
  const pathname = usePathname();
  const t = useTranslations();

  return (
    <nav className="space-y-8">
      {docsNavSections.map((section) => (
        <div key={section.titleKey}>
          <h3 className="text-[11px] font-semibold mb-3 uppercase tracking-[0.14em] text-[var(--azure-light-ink-soft)]">
            {t(section.titleKey)}
          </h3>
          <ul className="space-y-1">
            {section.items.map((item) => {
              const active = pathname === item.href;
              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    onClick={onNavigate}
                    className={cn(
                      "text-sm block px-3 py-1.5 rounded-full transition-colors",
                      active
                        ? "bg-[var(--azure-light-cyan-soft)] text-[var(--azure-light-cyan-ink)] font-semibold"
                        : "text-[var(--azure-light-ink-muted)] hover:text-[var(--azure-light-ink)] hover:bg-[var(--azure-light-surface-high)]"
                    )}
                  >
                    {t(item.titleKey)}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </nav>
  );
}
