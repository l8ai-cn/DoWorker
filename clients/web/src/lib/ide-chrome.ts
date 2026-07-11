import { pathnameHidesIdeSidebar } from "@/lib/ide-route";

export function hideIdeSidebar(pathname: string): boolean {
  return pathnameHidesIdeSidebar(pathname);
}

export function hideIdeChrome(pathname: string): boolean {
  return pathname.includes("/do-agent/")
    || pathname.includes("/loopal/")
    || pathname.includes("/mobile/pods/")
    || pathname.includes("/mobile/workers/");
}

export function hideMobileTabBar(pathname: string): boolean {
  if (/\/mobile\/workers\/?$/.test(pathname)) return false;
  return hideIdeSidebar(pathname) || hideIdeChrome(pathname);
}
