interface AcquireTarget {
  market: string;
  listing: string;
  version: string;
}

export function buildAcquireLink(
  coreWebUrl: string | undefined,
  target: AcquireTarget,
): string | null {
  if (!coreWebUrl?.trim()) {
    return null;
  }

  const url = new URL("/marketplace/acquire", coreWebUrl);
  url.search = new URLSearchParams([
    ["market", target.market],
    ["listing", target.listing],
    ["version", target.version],
  ]).toString();
  return url.toString();
}
