/**
 * Interaction mode constants — canonical source for frontend.
 * Mirrors agentfile.ModePTY / agentfile.ModeACP on the backend.
 */
export const POD_MODE_PTY = "pty" as const;
export const POD_MODE_ACP = "acp" as const;

// PodMode moved to @do-worker/service-interface (used by the shared PodData
// view-model); re-exported here to keep `@/lib/pod-modes` import paths.
export type { PodMode } from "@do-worker/service-interface";
