export interface EmbeddedAgentWorkbenchAccess {
  baseUrl: string;
  getAccessToken: () => Promise<string> | string;
  orgSlug: string;
  sessionId: string;
}
