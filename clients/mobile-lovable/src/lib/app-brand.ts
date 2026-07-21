export const APP_NAME = "Agent Cloud";
export const APP_TAGLINE = "AI Agent Workforce Platform";

export function pageTitle(segment?: string): string {
  return segment ? `${segment} — ${APP_NAME}` : APP_NAME;
}
