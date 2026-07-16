import type { AgentCommand } from "./contracts";

export function ComposerCommandMenu({
  commands,
  onSelect,
  query,
}: {
  commands: AgentCommand[];
  onSelect: (command: AgentCommand) => void;
  query: string | null;
}) {
  if (query === null) return null;
  const matches = commands.filter((command) =>
    command.name.toLowerCase().startsWith(query.toLowerCase()),
  );
  if (matches.length === 0) return null;

  return (
    <div className="absolute inset-x-0 bottom-full z-20 mb-2 overflow-hidden rounded-md border border-border bg-popover shadow-lg">
      {matches.map((command) => (
        <button
          aria-label={command.label}
          className="flex w-full items-start gap-3 px-3 py-2 text-left hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
          key={command.name}
          onClick={() => onSelect(command)}
          type="button"
        >
          <span className="w-20 shrink-0 font-mono text-xs font-medium">
            {command.label}
          </span>
          <span className="text-xs text-muted-foreground">
            {command.description}
          </span>
        </button>
      ))}
    </div>
  );
}

export function commandQuery(value: string) {
  const match = /^\/([A-Za-z0-9_-]*)$/.exec(value);
  return match ? match[1] : null;
}

export function parseAgentCommand(value: string, commands: AgentCommand[]) {
  const trimmed = value.trim();
  const match = /^\/([A-Za-z0-9_-]+)(?:\s+(.*))?$/.exec(trimmed);
  if (!match) return null;
  const command = commands.find((candidate) => candidate.name === match[1]);
  if (!command) return null;
  return { command, arguments: match[2]?.trim() ?? "" };
}
