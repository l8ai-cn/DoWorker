import { isSlashCommandText } from "@/components/SlashCommandMenu";
import type { HostFilesystemEntry } from "@/hooks/useHostFilesystem";
import type { Conversation } from "@/hooks/useConversations";

export function isValidWorkspace(workspace: string): boolean {
  return workspace.trim().startsWith("/");
}

export function normalizeWorkspacePath(path: string): string | null {
  const trimmed = path.trim();
  if (trimmed === "") return null;
  const stripped = trimmed.replace(/\/+$/, "");
  return stripped === "" ? "/" : stripped;
}

export function sessionsSharingDirectory(
  sessions: Conversation[],
  hostId: string | null,
  workspace: string,
  isRunnerOnline: (sessionId: string) => boolean,
): Conversation[] {
  if (!hostId) return [];
  const target = normalizeWorkspacePath(workspace);
  if (target === null) return [];
  return sessions.filter(
    (session) =>
      session.host_id === hostId &&
      session.workspace != null &&
      normalizeWorkspacePath(session.workspace) === target &&
      isRunnerOnline(session.id),
  );
}

export function sanitizeInitialPrompt(prompt: string): string {
  // eslint-disable-next-line no-control-regex
  return prompt.replace(/[\x00-\x08\x0b-\x1f\x7f-\x9f]/g, "").trim();
}

export function isValidSandboxRepoUrl(url: string): boolean {
  const trimmed = url.trim();
  return (
    /^https:\/\/[^\s#/]+\/[^\s#]+$/.test(trimmed) ||
    /^git@[^\s#:]+:[^\s#]+$/.test(trimmed)
  );
}

export function composeSandboxWorkspace(url: string, branch: string): string | undefined {
  const repo = url.trim();
  if (repo === "") return undefined;
  const ref = branch.trim();
  return ref === "" ? repo : `${repo}#${ref}`;
}

export function deriveRepoName(url: string): string | null {
  const trimmed = url.trim().replace(/\/+$/, "");
  if (trimmed === "") return null;
  const last = trimmed.split(/[/:]/).pop() ?? "";
  const name = last.endsWith(".git") ? last.slice(0, -4) : last;
  return name === "" ? null : name;
}

export function matchSkillInvocation(
  text: string,
  skills: ReadonlyArray<{ name: string }>,
): { name: string; args: string } | null {
  const trimmed = text.trim();
  if (!isSlashCommandText(trimmed)) return null;
  const command = trimmed.split(/\s+/)[0]!;
  const name = command.slice(1);
  if (!skills.some((skill) => skill.name === name)) return null;
  return { name, args: trimmed.slice(command.length).trim() };
}

export function deriveHomeDir(entries: HostFilesystemEntry[]): string | null {
  const first = entries[0];
  if (!first) return null;
  const slash = first.path.lastIndexOf("/");
  if (slash < 0) return null;
  return slash === 0 ? "/" : first.path.slice(0, slash);
}
