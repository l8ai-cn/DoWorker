"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { EndpointsTable } from "./_sections/endpoints-table";
import { EndpointDetails } from "./_sections/endpoint-details";

export default function ApiChannelsPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.api.channels.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.api.channels.description")}
      </p>

      <EndpointsTable />
      <EndpointDetails />

      <DocNavigation />
    </div>
  );
}
