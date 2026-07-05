import { pathnameHidesIdeSidebar } from "@/lib/ide-route";

export function hideIdeSidebar(pathname: string): boolean {
  return pathnameHidesIdeSidebar(pathname);
}

export function hideIdeChrome(pathname: string): boolean {
  return pathname.includes("/do-agent/") || pathname.includes("/loopal/");
}

export function hideMobileTabBar(pathname: string): boolean {
  return hideIdeSidebar(pathname) || hideIdeChrome(pathname);
}
