export function dedupeRepeatedText(text: string): string {
  const t = text.trim();
  if (t.length < 12) return text;

  const half = Math.floor(t.length / 2);
  if (t.slice(0, half) === t.slice(half, half + half) && t.slice(half + half).trim() === "") {
    return t.slice(0, half);
  }

  const parts = t.split(/\n{2,}/).map((p) => p.trim()).filter(Boolean);
  if (parts.length >= 2 && parts.every((p) => p === parts[0])) {
    return parts[0];
  }

  return text;
}
