"use client";

import { useTranslations } from "next-intl";
import { DocsTable } from "@/components/docs/DocsTable";
import { docsLabel, docsMono, twoColumnHeaders } from "@/components/docs/docs-table-helpers";

const EXPRESSION_ROWS = [
  ["strings", '"hello" + " world"'],
  ["numbers", "42, 3.14"],
  ["booleans", "true, false"],
  ["dotAccess", "config.model, sandbox.root"],
  ["operators", "+, ==, !=, and, or, not"],
  ["functions", "json(...), str_replace(...), env(...)"],
] as const;

export function BuildLogicExpressions() {
  const t = useTranslations();
  const prefix = "docs.concepts.agentfile.expressions";

  return (
    <>
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.concepts.agentfile.buildLogic.title")}
        </h2>
        <p className="text-muted-foreground leading-relaxed mb-6">
          {t("docs.concepts.agentfile.buildLogic.description")}
        </p>
        <pre className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 text-sm overflow-x-auto mb-4">
          <code>{`# Variable assignment
model_flag = "--model " + config.model

# arg — append CLI argument
arg model_flag

# file — write a file
file ".env" "NODE_ENV=production"

# mkdir — create a directory
mkdir sandbox.work_dir + "/output"

# if / else
if config.verbose {
  arg "--verbose"
} else {
  arg "--quiet"
}

# for / in
for server in mcp.servers {
  arg "--mcp-server " + server
}

# when — shorthand conditional
when config.verbose arg "--verbose"`}</code>
        </pre>
      </section>

      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">{t(`${prefix}.title`)}</h2>
        <p className="text-muted-foreground leading-relaxed mb-6">{t(`${prefix}.description`)}</p>
        <DocsTable
          columns={twoColumnHeaders(t, prefix, "typeHeader", "exampleHeader")}
          rows={EXPRESSION_ROWS.map(([typeKey, example]) => ({
            cells: [docsLabel(t(`${prefix}.types.${typeKey}`)), docsMono(example)],
          }))}
        />
      </section>
    </>
  );
}
