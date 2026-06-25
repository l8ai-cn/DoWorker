"use client";

import { useTranslations } from "next-intl";
import { DocsTable } from "@/components/docs/DocsTable";
import { docsMono, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

const BUILTIN_VARS = [
  ["config.*", "configDesc"],
  ["sandbox.root", "sandboxRoot"],
  ["sandbox.work_dir", "sandboxWorkDir"],
  ["mcp.enabled", "mcpEnabled"],
  ["mcp.servers", "mcpServers"],
  ["mcp.format", "mcpFormat"],
  ["mode", "modeDesc"],
] as const;

export function BuiltinVarsEscaping() {
  const t = useTranslations();
  const prefix = "docs.concepts.agentfile.builtinVars";

  return (
    <>
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">{t(`${prefix}.title`)}</h2>
        <p className="text-muted-foreground leading-relaxed mb-6">{t(`${prefix}.description`)}</p>
        <DocsTable
          columns={twoColumnHeaders(t, prefix, "variableHeader", "descHeader")}
          rows={BUILTIN_VARS.map(([variable, descKey]) => ({
            cells: [docsMono(variable), t(`${prefix}.${descKey}`)],
          }))}
        />
      </section>

      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.concepts.agentfile.escaping.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          {t("docs.concepts.agentfile.escaping.description")}
        </p>
        <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-3 text-sm overflow-x-auto">
          <code>{`\\\\  →  backslash
\\"  →  double quote
\\n  →  newline
\\t  →  tab`}</code>
        </pre>
      </section>
    </>
  );
}
