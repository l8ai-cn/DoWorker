export type ReconnectState = "host_offline" | "local_stranded";

const CLAUDE_NATIVE_WRAPPER = "claude-code-native-ui";

export function buildReconnectCommand({
  conversationId,
  serverUrl,
  wrapper,
  state,
}: {
  conversationId: string;
  serverUrl: string;
  wrapper?: string | null;
  state: ReconnectState;
}): string {
  if (state === "host_offline") {
    return [
      "runner run \\",
      `  # server: ${serverUrl}`,
      "  # uses ~/.agent-cloud/config.yaml from runner register",
    ].join("\n");
  }
  if (wrapper === CLAUDE_NATIVE_WRAPPER) {
    return [
      "runner run \\",
      `  # resume session ${conversationId}`,
      `  # server: ${serverUrl}`,
    ].join("\n");
  }
  return [
    "runner run \\",
    `  # resume session ${conversationId}`,
    `  # server: ${serverUrl}`,
    "  # or re-open the session from Agent Cloud web UI",
  ].join("\n");
}
