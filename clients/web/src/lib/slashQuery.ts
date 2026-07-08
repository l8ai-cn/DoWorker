/** Detect a leading `/command` token while the cursor is still in the name (no space yet). */
export function getSlashQuery(
  text: string,
  cursor: number,
): { query: string; startIndex: number } | null {
  if (!text.startsWith("/")) return null;
  const head = text.slice(0, cursor);
  const m = head.match(/^\/(\S*)$/);
  return m ? { query: m[1] ?? "", startIndex: 0 } : null;
}
