const SAFE_IMAGE_MEDIA_TYPES = new Set([
  "image/avif",
  "image/gif",
  "image/jpeg",
  "image/png",
  "image/webp",
]);

export function isSafeArtifactImage(mediaType: string): boolean {
  return SAFE_IMAGE_MEDIA_TYPES.has(mediaType.toLowerCase());
}
