export type SessionInteractionMode = "acp" | "pty";

export interface AvailableAgent {
  id: string;
  workerTypeSlug?: string;
  supportedModes?: SessionInteractionMode[];
  requiresModelResource?: boolean;
  name: string;
  display_name: string;
  description: string | null;
  harness: string | null;
  skills: { name: string; description: string }[];
  builtin?: boolean;
  created_at?: number | null;
}
