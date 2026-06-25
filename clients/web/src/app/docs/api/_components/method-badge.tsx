"use client";

import { useTranslations } from "next-intl";

const METHOD_TONE: Record<string, string> = {
  GET: "text-info",
  POST: "text-success",
  PUT: "text-warning",
  PATCH: "text-warning",
  DELETE: "text-danger",
};

export function MethodBadge({
  method,
  size = "sm",
}: {
  method: string;
  size?: "sm" | "md";
}) {
  const padding = size === "md" ? "px-2 py-1" : "px-1";
  const tone = METHOD_TONE[method] ?? "text-foreground";
  return <code className={`bg-muted ${padding} rounded ${tone}`}>{method}</code>;
}

export function RequiredBadge() {
  const t = useTranslations();
  return (
    <span className="text-xs bg-danger-bg text-danger px-2 py-0.5 rounded">
      {t("docs.api.common.requiredBadge")}
    </span>
  );
}

export function OptionalBadge() {
  const t = useTranslations();
  return (
    <span className="text-xs bg-muted text-muted-foreground px-2 py-0.5 rounded">
      {t("docs.api.common.optionalBadge")}
    </span>
  );
}
