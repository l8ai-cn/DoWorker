const IDENTIFIER_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

function invalidPreviewResponse(): never {
  throw new Error("Invalid preview response");
}

function parsePublicOrigin(rawOrigin: string): URL {
  let origin: URL;
  try {
    origin = new URL(rawOrigin);
  } catch {
    return invalidPreviewResponse();
  }

  if (
    !["http:", "https:"].includes(origin.protocol) ||
    origin.pathname !== "/" ||
    origin.search ||
    origin.hash ||
    origin.username ||
    origin.password
  ) {
    return invalidPreviewResponse();
  }
  return origin;
}

function validatePodKey(podKey: string): void {
  if (
    podKey.length < 2 ||
    podKey.length > 100 ||
    !IDENTIFIER_PATTERN.test(podKey)
  ) {
    invalidPreviewResponse();
  }
}

export function getPodPreviewOrigin(
  podKey: string,
  publicOrigin: string,
): string {
  validatePodKey(podKey);
  const base = parsePublicOrigin(publicOrigin);
  base.hostname = `${podKey}.${base.hostname}`;
  return base.origin;
}

export function parsePreviewSessionUrl(
  rawUrl: string,
  podKey: string,
  publicOrigin: string,
): URL {
  let sessionUrl: URL;
  try {
    sessionUrl = new URL(rawUrl);
  } catch {
    return invalidPreviewResponse();
  }

  const tokenValues = sessionUrl.searchParams.getAll("token");
  if (
    sessionUrl.origin !== getPodPreviewOrigin(podKey, publicOrigin) ||
    sessionUrl.pathname !== `/preview/${podKey}/__session` ||
    sessionUrl.hash ||
    sessionUrl.username ||
    sessionUrl.password ||
    tokenValues.length !== 1 ||
    !tokenValues[0] ||
    [...sessionUrl.searchParams.keys()].some((key) => key !== "token")
  ) {
    return invalidPreviewResponse();
  }
  return sessionUrl;
}

export function parsePreviewBaseUrl(
  rawUrl: string,
  podKey: string,
  publicOrigin: string,
): URL {
  let previewUrl: URL;
  try {
    previewUrl = new URL(rawUrl);
  } catch {
    return invalidPreviewResponse();
  }

  if (
    previewUrl.origin !== getPodPreviewOrigin(podKey, publicOrigin) ||
    previewUrl.pathname !== `/preview/${podKey}/` ||
    previewUrl.search ||
    previewUrl.hash ||
    previewUrl.username ||
    previewUrl.password
  ) {
    return invalidPreviewResponse();
  }
  return previewUrl;
}

export function requirePreviewPublicOrigin(): string {
  const origin = import.meta.env.VITE_PREVIEW_PUBLIC_ORIGIN;
  if (!origin) {
    throw new Error("VITE_PREVIEW_PUBLIC_ORIGIN is required");
  }
  return parsePublicOrigin(origin).origin;
}
