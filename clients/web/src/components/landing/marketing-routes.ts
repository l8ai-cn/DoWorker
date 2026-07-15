export const productStoryRoutes = [
  { id: "product", href: "/product", labelKey: "landing.nav.product" },
  { id: "solutions", href: "/solutions", labelKey: "landing.nav.solutions" },
] as const;

export const marketingRoutes = [
  { id: "home", href: "/", labelKey: "landing.nav.home" },
  ...productStoryRoutes,
  { id: "marketplace", href: "/marketplace", labelKey: "landing.nav.marketplace" },
  { id: "docs", href: "/docs", labelKey: "landing.nav.docs" },
] as const;

export function isMarketingRouteActive(pathname: string, href: string) {
  return href === "/"
    ? pathname === "/"
    : pathname === href || pathname.startsWith(`${href}/`);
}
