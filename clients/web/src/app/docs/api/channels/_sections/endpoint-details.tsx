"use client";

import { useTranslations } from "next-intl";
import { EndpointDetail } from "../../_components/endpoint-detail";
import { detailEndpoints } from "./endpoints-data";

export function EndpointDetails() {
  const t = useTranslations();

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-6">
        {t("docs.api.channels.details.title")}
      </h2>
      <div className="space-y-8">
        {detailEndpoints.map((spec) => (
          <EndpointDetail
            key={`${spec.method}-${spec.path}`}
            spec={spec}
          />
        ))}
      </div>
    </section>
  );
}
