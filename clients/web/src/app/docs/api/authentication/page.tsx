"use client";

import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";
import {
  buildTripleKeyRows,
  docsMono,
  threeColumnHeaders,
  twoColumnHeaders,
} from "@/components/docs/docs-table-helpers";

export default function ApiAuthenticationPage() {
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.api.authentication.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.api.authentication.description")}
      </p>

      {/* Authentication Methods */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.api.authentication.methods.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.api.authentication.methods.description")}
        </p>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.api.authentication.methods.headerMethod")}
            </h3>
            <p className="text-sm text-muted-foreground mb-3">
              {t("docs.api.authentication.methods.headerMethodDesc")}
            </p>
            <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm">
              <pre className="text-success">{`curl -H "X-API-Key: amk_your_api_key_here" \\
  https://your-domain.com/api/v1/ext/orgs/my-org/pods`}</pre>
            </div>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.api.authentication.methods.bearerMethod")}
            </h3>
            <p className="text-sm text-muted-foreground mb-3">
              {t("docs.api.authentication.methods.bearerMethodDesc")}
            </p>
            <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm">
              <pre className="text-success">{`curl -H "Authorization: Bearer amk_your_api_key_here" \\
  https://your-domain.com/api/v1/ext/orgs/my-org/pods`}</pre>
            </div>
          </div>
        </div>
      </section>

      {/* Scopes */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.api.authentication.scopes.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.api.authentication.scopes.description")}
        </p>
        <DocsTable
          columns={threeColumnHeaders(
            t,
            "docs.api.authentication.scopes",
            "scopeHeader",
            "descriptionHeader",
            "endpointsHeader",
          )}
          rows={buildTripleKeyRows(t, "docs.api.authentication.scopes", [
            ["podsRead", "podsReadDesc", "podsReadEndpoints"],
            ["podsWrite", "podsWriteDesc", "podsWriteEndpoints"],
            ["ticketsRead", "ticketsReadDesc", "ticketsReadEndpoints"],
            ["ticketsWrite", "ticketsWriteDesc", "ticketsWriteEndpoints"],
            ["channelsRead", "channelsReadDesc", "channelsReadEndpoints"],
            ["channelsWrite", "channelsWriteDesc", "channelsWriteEndpoints"],
            ["runnersRead", "runnersReadDesc", "runnersReadEndpoints"],
            ["reposRead", "reposReadDesc", "reposReadEndpoints"],
          ], { monoFirst: true })}
        />
      </section>

      {/* Error Handling */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.api.authentication.errors.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.api.authentication.errors.description")}
        </p>
        <DocsTable
          columns={twoColumnHeaders(t, "docs.api.authentication.errors", "codeHeader", "descriptionHeader")}
          rows={[
            { cells: [docsMono("400"), t("docs.api.authentication.errors.badRequest")] },
            { cells: [docsMono("401"), t("docs.api.authentication.errors.unauthorized")] },
            { cells: [docsMono("403"), t("docs.api.authentication.errors.forbidden")] },
            { cells: [docsMono("404"), t("docs.api.authentication.errors.notFound")] },
          ]}
        />
      </section>

      <DocNavigation />
    </div>
  );
}
