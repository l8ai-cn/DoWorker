import { getSlashQuery } from "@/lib/slashQuery";

/** Configurable slash-command registry for Worker / create-Worker composers. */
export type WorkerSlashCommandDef = {
  /** Command name without leading slash, e.g. `compact`. */
  id: string;
  /** When true, selecting the command leaves the textarea at `/cmd ` for args. */
  hasArg?: boolean;
};

/**
 * Built-in maintenance / control commands forwarded to the native agent CLI
 * as plain prompt text. Extend this list to add org- or agent-specific cmds.
 */
export const WORKER_SLASH_COMMAND_DEFS: WorkerSlashCommandDef[] = [
  { id: "compact" },
  { id: "context" },
  { id: "help" },
  { id: "model", hasArg: true },
  { id: "effort", hasArg: true },
];

export type WorkerSlashCommand = WorkerSlashCommandDef & {
  label: string;
  hint: string;
};

export function buildWorkerSlashCommands(
  t: (key: string) => string,
  extra: WorkerSlashCommandDef[] = [],
): WorkerSlashCommand[] {
  return [...WORKER_SLASH_COMMAND_DEFS, ...extra].map((def) => ({
    ...def,
    label: `/${def.id}`,
    hint: t(`workerSlash.commands.${def.id}`),
  }));
}

export function filterWorkerSlashCommands(
  commands: WorkerSlashCommand[],
  query: string,
): WorkerSlashCommand[] {
  const q = query.toLowerCase();
  if (!q) return commands;
  return commands.filter((c) => c.id.startsWith(q));
}

export function parseWorkerSlashInput(
  text: string,
  commands: WorkerSlashCommand[],
): { command: WorkerSlashCommand; arg: string } | null {
  const trimmed = text.trim();
  if (!trimmed.startsWith("/")) return null;
  const sp = trimmed.indexOf(" ");
  const name = (sp < 0 ? trimmed.slice(1) : trimmed.slice(1, sp)).toLowerCase();
  const arg = sp < 0 ? "" : trimmed.slice(sp + 1).trim();
  const command = commands.find((c) => c.id === name);
  return command ? { command, arg } : null;
}

export function isWorkerSlashCommandText(text: string): boolean {
  return /^\/[A-Za-z0-9][\w:-]*(\s|$)/.test(text.trim());
}

export { getSlashQuery };
