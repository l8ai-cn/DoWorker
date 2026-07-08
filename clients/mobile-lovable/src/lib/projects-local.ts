export interface LocalProjectMeta {
  name: string;
  repo?: string;
  host?: string;
  color?: "primary" | "accent" | "info";
}

const KEY = "agentsmesh-mobile-projects";

function readAll(): LocalProjectMeta[] {
  if (typeof localStorage === "undefined") return [];
  try {
    return JSON.parse(localStorage.getItem(KEY) ?? "[]") as LocalProjectMeta[];
  } catch {
    return [];
  }
}

function writeAll(items: LocalProjectMeta[]): void {
  localStorage.setItem(KEY, JSON.stringify(items));
}

export function listLocalProjects(): LocalProjectMeta[] {
  return readAll();
}

export function saveLocalProject(meta: LocalProjectMeta): void {
  const items = readAll().filter((p) => p.name !== meta.name);
  items.unshift(meta);
  writeAll(items.slice(0, 50));
}

export function localProjectMeta(name: string): LocalProjectMeta | undefined {
  return readAll().find((p) => p.name === name);
}
