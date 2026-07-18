import Papa from "papaparse";

import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";
import { MAX_TEXT_PREVIEW_BYTES } from "../../useArtifactBlobUrl";

const MAX_COLUMNS = 30;
const MAX_ROWS = 200;

export function ArtifactCsvPreview({
  filename,
  text,
  truncated,
}: {
  filename: string;
  text: string;
  truncated: boolean;
}) {
  const labels = useAgentWorkspaceText().artifact;
  const parsed = Papa.parse<string[]>(text, {
    skipEmptyLines: "greedy",
  });
  if (parsed.errors.length > 0 || parsed.data.length === 0) {
    return (
      <div className="border-b border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive" role="alert">
        {labels.csvPreviewFailed(filename)}
      </div>
    );
  }
  const rows = parsed.data.slice(0, MAX_ROWS + 1);
  const columnCount = Math.min(
    MAX_COLUMNS,
    rows.reduce((maximum, row) => Math.max(maximum, row.length), 0),
  );
  return (
    <div className="max-h-[32rem] overflow-auto border-b border-border">
      <table aria-label={labels.csvPreview(filename)} className="min-w-full border-collapse text-left text-xs">
        <thead className="sticky top-0 z-10 bg-muted">
          <tr>
            {rows[0].slice(0, columnCount).map((cell, index) => (
              <th className="border border-border px-2 py-1.5 font-medium" key={`${index}:${cell}`}>
                {cell || `Column ${index + 1}`}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.slice(1).map((row, rowIndex) => (
            <tr key={rowIndex}>
              {Array.from({ length: columnCount }, (_, columnIndex) => (
                <td className="max-w-72 border border-border px-2 py-1.5 align-top" key={columnIndex}>
                  <span className="line-clamp-3 break-words">{row[columnIndex] ?? ""}</span>
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {(parsed.data.length > MAX_ROWS + 1 || columnCount === MAX_COLUMNS) && (
        <div className="border-t border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
          Preview limited to {MAX_ROWS} rows and {MAX_COLUMNS} columns.
        </div>
      )}
      {truncated && (
        <div className="border-t border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
          {labels.previewLimited(MAX_TEXT_PREVIEW_BYTES)}
        </div>
      )}
    </div>
  );
}
