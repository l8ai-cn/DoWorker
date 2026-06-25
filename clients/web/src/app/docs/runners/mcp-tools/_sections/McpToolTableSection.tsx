import { DocsTable } from "@/components/docs/DocsTable";
import type { DocsTableRow } from "@/components/docs/DocsTable";

interface McpToolTableSectionProps {
  title: string;
  description?: string;
  columns: Array<{ header: string; className?: string }>;
  rows: DocsTableRow[];
}

export function McpToolTableSection({
  title,
  description,
  columns,
  rows,
}: McpToolTableSectionProps) {
  return (
    <section className="mb-12">
      <h2 className="text-2xl font-semibold mb-4 text-foreground">{title}</h2>
      {description && (
        <p className="text-muted-foreground mb-4">{description}</p>
      )}
      <DocsTable columns={columns} rows={rows} />
    </section>
  );
}
