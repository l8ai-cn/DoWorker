import type {
  ContentBlock,
  TableContent,
} from "@do-worker/proto/agent_workbench/v2/content_pb";

import {
  decodeStructuredPayload,
  formatStructuredPayload,
  stringifyValue,
} from "./projectGeneratedSessionSnapshotPayload";

export function contentBlockText(block: ContentBlock): string | undefined {
  const content = block.content;
  if (content.case === "text") return content.value.text;
  if (content.case === "markdown") return content.value.markdown;
  if (content.case === "code") {
    const label = [content.value.language, content.value.filename]
      .filter(Boolean)
      .join(" ");
    return `\`\`\`${label}\n${content.value.code}\n\`\`\``;
  }
  if (content.case === "json") {
    return formatStructuredPayload(content.value.value) ?? "JSON payload missing";
  }
  if (content.case === "table") return formatTable(content.value);
  if (content.case === "diff") {
    const heading = [
      content.value.path,
      content.value.language,
      content.value.truncated ? "truncated" : undefined,
    ]
      .filter(Boolean)
      .join(" ");
    return `\`\`\`diff${heading ? ` ${heading}` : ""}\n${content.value.patch}\n\`\`\``;
  }
  if (content.case === "command") {
    return [
      content.value.cwd ? `cwd: ${content.value.cwd}` : undefined,
      `$ ${content.value.command}`,
      content.value.exitCode === undefined
        ? undefined
        : `exitCode: ${content.value.exitCode}`,
    ]
      .filter(Boolean)
      .join("\n");
  }
  if (content.case === "log") {
    const fields = formatStructuredPayload(content.value.fields);
    return [
      `[${content.value.level || "log"}] ${content.value.message}`,
      content.value.createdAt,
      fields,
    ]
      .filter(Boolean)
      .join("\n");
  }
  if (content.case === "progress") {
    const progress = content.value;
    const count =
      progress.current !== undefined || progress.total !== undefined
        ? `${progress.current?.toString() ?? "?"}/${progress.total?.toString() ?? "?"}${progress.unit ? ` ${progress.unit}` : ""}`
        : undefined;
    const fraction =
      progress.fraction === undefined
        ? undefined
        : `${Math.round(progress.fraction * 100)}%`;
    return [progress.stage, progress.message, count, fraction]
      .filter(Boolean)
      .join(" - ");
  }
  if (content.case === "error") {
    const details = formatStructuredPayload(content.value.details);
    return [
      `[${content.value.code || "content_error"}] ${content.value.message}`,
      details,
    ]
      .filter(Boolean)
      .join("\n");
  }
  if (content.case === "html" && content.value.payload.case === "source") {
    return [
      `HTML security profile: ${content.value.securityProfile}`,
      `\`\`\`html\n${content.value.payload.value}\n\`\`\``,
    ].join("\n");
  }
  if (content.case === "link") {
    const link = content.value.label
      ? `[${content.value.label}](${content.value.url})`
      : content.value.url;
    return [
      link,
      content.value.mediaType
        ? `Media type: ${content.value.mediaType}`
        : undefined,
    ]
      .filter(Boolean)
      .join("\n");
  }
  if (content.case === "citation") {
    return [
      `Citation id: ${content.value.citationId}`,
      content.value.url
        ? `[${content.value.label}](${content.value.url})`
        : content.value.label,
      content.value.excerpt,
    ]
      .filter(Boolean)
      .join("\n");
  }
  if (
    content.case === "image" ||
    content.case === "video" ||
    content.case === "audio" ||
    content.case === "pdf" ||
    content.case === "presentation" ||
    content.case === "spreadsheet" ||
    content.case === "file"
  ) {
    return content.value.altText;
  }
  return undefined;
}

export function contentBlockData(block: ContentBlock): unknown {
  const content = block.content;
  if (content.case === "json") {
    return decodeStructuredPayload(content.value.value)?.value;
  }
  if (content.case === "table") {
    return {
      columns: content.value.columns,
      rows: content.value.rows.map((row) =>
        row.cells.map((cell) => decodeStructuredPayload(cell)?.value),
      ),
    };
  }
  if (content.case === "progress") return content.value;
  return stringifyValue(content.value);
}

function formatTable(table: TableContent): string {
  const labels = table.columns.map((column) => {
    const label = column.label || column.key;
    const metadata = [
      column.key !== label ? column.key : undefined,
      column.valueType,
    ]
      .filter(Boolean)
      .join(": ");
    return metadata ? `${label} (${metadata})` : label;
  });
  if (labels.length === 0) return stringifyValue(table.rows);
  const header = `| ${labels.join(" | ")} |`;
  const separator = `| ${labels.map(() => "---").join(" | ")} |`;
  const rows = table.rows.map((row) => {
    const cells = row.cells.map(
      (cell) => formatStructuredPayload(cell)?.replaceAll("|", "\\|") ?? "",
    );
    return `| ${cells.join(" | ")} |`;
  });
  return [header, separator, ...rows].join("\n");
}
