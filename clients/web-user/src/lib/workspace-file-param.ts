function stripTrailingEquals(path: string): string {
  return path.endsWith("=") ? path.slice(0, -1) : path;
}

export function parseWorkspaceFileParam(searchParams: URLSearchParams): string | null {
  const direct = searchParams.get("file");
  if (direct !== null && direct !== "") {
    return stripTrailingEquals(direct);
  }

  for (const [key, value] of searchParams.entries()) {
    if (key === "file" && value !== "") {
      return stripTrailingEquals(value);
    }
    const embedded = /^file=(.+)$/.exec(key);
    if (embedded?.[1]) {
      return stripTrailingEquals(embedded[1]);
    }
  }

  return null;
}

export function normalizeWorkspaceFileSearch(search: string): string {
  const raw = search.startsWith("?") ? search.slice(1) : search;
  if (!raw) return "";

  const params = new URLSearchParams(raw);
  const filePath = parseWorkspaceFileParam(params);
  if (!filePath) return search.startsWith("?") ? search : `?${raw}`;

  const next = new URLSearchParams();
  for (const [key, value] of params.entries()) {
    if (key === "file" || key.startsWith("file=")) continue;
    next.append(key, value);
  }
  next.set("file", filePath);

  const qs = next.toString();
  return qs ? `?${qs}` : "";
}

export function isHtmlWorkspacePath(path: string): boolean {
  return path.toLowerCase().endsWith(".html") || path.toLowerCase().endsWith(".htm");
}
