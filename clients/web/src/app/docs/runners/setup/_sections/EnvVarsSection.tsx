"use client";

import { useTranslations } from "next-intl";
import { DocsTable } from "@/components/docs/DocsTable";
import { docsMono, threeColumnHeaders } from "@/components/docs/docs-table-helpers";

interface EnvVarsSectionProps {
  serverUrl: string;
}

export function EnvVarsSection({ serverUrl }: EnvVarsSectionProps) {
  const t = useTranslations();
  const prefix = "docs.runners.setup.envVars";

  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">{t(`${prefix}.title`)}</h2>
      <DocsTable
        columns={threeColumnHeaders(t, prefix, "variableHeader", "descriptionHeader", "defaultHeader")}
        rows={[
          { cells: [docsMono("AGENTSMESH_TOKEN"), t(`${prefix}.tokenDesc`), "-"] },
          { cells: [docsMono("AGENTSMESH_URL"), t(`${prefix}.urlDesc`), docsMono(serverUrl)] },
          { cells: [docsMono("MAX_CONCURRENT_PODS"), t(`${prefix}.maxPodsDesc`), "5"] },
          { cells: [docsMono("WORKSPACE_DIR"), t(`${prefix}.workspaceDirDesc`), docsMono("/data/workspaces")] },
          { cells: [docsMono("MCP_PORT"), t(`${prefix}.mcpPortDesc`), "19000"] },
        ]}
      />
    </section>
  );
}
