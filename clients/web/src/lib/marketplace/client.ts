import { readLightAuthToken, resolveLightBaseUrl } from "@/lib/light-session";

export class MarketplaceRequestError extends Error {
  constructor(
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "MarketplaceRequestError";
  }
}

export async function marketplaceRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const baseURL = resolveLightBaseUrl();
  const token = readLightAuthToken(baseURL);
  const response = await fetch(`${baseURL}/api/marketplace/v1${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  });
  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    const error = payload?.error;
    throw new MarketplaceRequestError(
      error?.code ?? "MARKETPLACE_REQUEST_FAILED",
      error?.message ?? "市场服务暂时不可用",
    );
  }
  return payload as T;
}
