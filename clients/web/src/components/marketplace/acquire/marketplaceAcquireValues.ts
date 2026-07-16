import { MarketplaceAcquireError } from "@/lib/marketplace/acquire-api";

export type MarketplaceAcquireStep =
  | "select"
  | "confirm"
  | "installing"
  | "success";

export function marketplaceAcquireErrorMessage(cause: unknown): string {
  if (cause instanceof MarketplaceAcquireError) return cause.message;
  if (cause instanceof Error) return cause.message;
  return "启用失败，请稍后重试。";
}

export function numericToolModelIDs(
  values: Record<string, string>,
): Record<string, number> {
  return Object.fromEntries(
    Object.entries(values).map(([role, id]) => [role, Number(id)]),
  );
}
