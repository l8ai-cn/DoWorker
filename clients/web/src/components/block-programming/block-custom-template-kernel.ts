const BLOCK_PARAMETER = /{{\s*([a-z0-9]+(?:-[a-z0-9]+)*)\s*}}/g;

export function extractBlockTemplateParameters(templates: readonly string[]): string[] {
  const parameters: string[] = [];
  for (const template of templates) {
    for (const match of template.matchAll(BLOCK_PARAMETER)) {
      if (!parameters.includes(match[1])) parameters.push(match[1]);
    }
  }
  return parameters;
}

export function expandBlockTemplate(
  template: string,
  values: Record<string, string>,
): { value: string; missing: string[] } {
  const missing = extractBlockTemplateParameters([template]).filter(
    (parameter) => !values[parameter]?.trim(),
  );
  return {
    missing,
    value: template.replace(BLOCK_PARAMETER, (_, parameter: string) => values[parameter] ?? ""),
  };
}

export function matchBlockTemplate(
  template: string,
  value: string,
): Record<string, string> | undefined {
  const parameters = extractBlockTemplateParameters([template]);
  const parts: Array<string | { parameter: string }> = [];
  let offset = 0;
  for (const match of template.matchAll(BLOCK_PARAMETER)) {
    parts.push(template.slice(offset, match.index));
    parts.push({ parameter: match[1] });
    offset = match.index + match[0].length;
  }
  parts.push(template.slice(offset));
  let pattern = "^";
  for (const part of parts) {
    pattern += typeof part === "string" ? escapeRegExp(part) : "(.+?)";
  }
  const match = value.match(new RegExp(`${pattern}$`));
  if (!match) return undefined;
  return Object.fromEntries(parameters.map((parameter, index) => [parameter, match[index + 1]]));
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
