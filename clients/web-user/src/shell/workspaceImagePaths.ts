export function isWorkspaceRelativeSrc(src: string): boolean {
  return !/^([a-z][a-z0-9+.-]*:|\/\/)/i.test(src);
}

export function resolveWorkspacePath(filePath: string, src: string): string {
  const cleanSrc = src.split(/[?#]/)[0];
  const segments = cleanSrc.startsWith("/")
    ? []
    : filePath.split("/").slice(0, -1);

  for (const rawSegment of cleanSrc.split("/")) {
    let segment = rawSegment;
    try {
      const decoded = decodeURIComponent(rawSegment);
      if (!decoded.includes("/") && !decoded.includes("\\")) {
        segment = decoded;
      }
    } catch {
      segment = rawSegment;
    }
    if (segment === "" || segment === ".") continue;
    if (segment === "..") segments.pop();
    else segments.push(segment);
  }

  return segments.join("/");
}
