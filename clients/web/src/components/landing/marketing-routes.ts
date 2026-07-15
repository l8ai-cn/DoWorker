export const expertStoryRoutes = [
  { id: "solutions", href: "/solutions", labelKey: "landing.nav.scenarios" },
  { id: "how-it-works", href: "/how-it-works", labelKey: "landing.nav.workflow" },
  { id: "capabilities", href: "/capabilities", labelKey: "landing.nav.capabilities" },
] as const;

export const marketingRoutes = [
  { id: "home", href: "/", labelKey: "landing.nav.home" },
  ...expertStoryRoutes,
  { id: "marketplace", href: "/marketplace", labelKey: "landing.nav.marketplace" },
  { id: "docs", href: "/docs", labelKey: "landing.nav.docs" },
] as const;

export function isMarketingRouteActive(pathname: string, href: string) {
  return href === "/"
    ? pathname === "/"
    : pathname === href || pathname.startsWith(`${href}/`);
}
