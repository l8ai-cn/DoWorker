export function taskTitleForSubmit(
  prompt: string,
  expertName: string | undefined,
  engineName: string,
): string {
  const head = expertName ? `${expertName} · ` : `${engineName} · `;
  const body = prompt.trim().slice(0, 60);
  return (head + body).slice(0, 80);
}

export function messageWithExpertContext(
  prompt: string,
  expertName?: string,
  expertHint?: string | null,
): string {
  if (!expertName) return prompt;
  const hint = expertHint?.trim();
  if (hint) return `[专家模式: ${expertName}]\n${hint}\n\n${prompt}`;
  return `[专家模式: ${expertName}]\n\n${prompt}`;
}
