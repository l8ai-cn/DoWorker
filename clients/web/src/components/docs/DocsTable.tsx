import type { ReactNode } from "react";
import { DocsHorizontalScroll } from "./DocsHorizontalScroll";

interface DocsTableColumn {
  header: string;
  className?: string;
}

interface DocsTableRow {
  cells: ReactNode[];
  rowClassName?: string;
}

interface DocsTableProps {
  columns: DocsTableColumn[];
  rows: DocsTableRow[];
}

export type { DocsTableRow };
export function DocsTable({ columns, rows }: DocsTableProps) {
  const tableWidth = columns.length > 2 ? "min-w-[640px]" : "w-full";

  return (
    <DocsHorizontalScroll>
      <div className="rounded-lg overflow-hidden surface-card">
        <table className={`${tableWidth} text-sm divide-y divide-border/20`}>
          <thead>
            <tr className="bg-surface-muted/50">
              {columns.map((col) => (
                <th key={col.header} className={`text-left p-3 ${col.className ?? ""}`}>
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="text-muted-foreground divide-y divide-border/20">
            {rows.map((row, i) => (
              <tr key={i} className={row.rowClassName}>
                {row.cells.map((cell, j) => (
                  <td key={j} className="p-3">
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </DocsHorizontalScroll>
  );
}
