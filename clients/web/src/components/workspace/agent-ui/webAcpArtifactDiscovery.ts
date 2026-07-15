import type { AgentArtifactItem } from "@do-worker/agent-ui";

import type { WebAcpRuntimeDeps } from "./webAcpRuntimeTypes";

export class WebAcpArtifactDiscovery {
  private artifacts: AgentArtifactItem[] = [];
  private currentState = "idle";
  private inFlight: Promise<void> | null = null;
  error: string | null = null;

  constructor(
    private readonly podKey: string,
    private readonly deps: () => WebAcpRuntimeDeps,
    private readonly onChange: () => void,
  ) {}

  items(): AgentArtifactItem[] {
    return this.artifacts;
  }

  async open(state: string): Promise<void> {
    this.currentState = state;
    await this.refresh();
  }

  observe(state: string): void {
    const completedTurn = this.currentState !== "idle" && state === "idle";
    this.currentState = state;
    if (completedTurn) void this.refresh();
  }

  private refresh(): Promise<void> {
    if (this.inFlight) return this.inFlight;
    const refresh = this.deps()
      .listWorkspaceArtifacts(this.podKey)
      .then((artifacts) => {
        this.artifacts = artifacts;
        this.error = null;
        this.onChange();
      })
      .catch((cause: unknown) => {
        this.error = cause instanceof Error ? cause.message : String(cause);
        this.onChange();
      })
      .finally(() => {
        if (this.inFlight === refresh) this.inFlight = null;
      });
    this.inFlight = refresh;
    return refresh;
  }
}
