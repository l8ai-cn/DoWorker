export function markdownImageSource(src?: string): string | undefined {
  if (!src) return undefined;
  return /^(blob:|data:image\/)/i.test(src) ? src : undefined;
}
