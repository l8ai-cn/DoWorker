import type { AgentConnectionStatus, AgentPermissionResolution, AgentSessionRuntime, AgentSessionSnapshot } from "@do-worker/agent-ui";
import { EMPTY_SESSION } from "@/stores/acpSessionTypes";
import { defaultWebAcpRuntimeDeps, relayConnection } from "./webAcpRuntimeDefaults";
import type { WebAcpRuntimeDeps, WebAcpSessionRuntimeInput } from "./webAcpRuntimeTypes";
import { WebAcpArtifactDiscovery } from "./webAcpArtifactDiscovery";
import { retainWebAcpPodSubscription } from "./webAcpPodSubscription";
import { projectWebAcpSnapshot } from "./webAcpSnapshot";

export type { WebAcpRuntimeDeps } from "./webAcpRuntimeTypes";

export class WebAcpSessionRuntime implements AgentSessionRuntime {
  readonly sessionId: string;
  private readonly listeners = new Set<() => void>();
  private readonly subscriptionBase: string;
  private connection: AgentConnectionStatus = "connecting";
  private error: string | null = null;
  private cleanup: (() => void) | null = null;
  private opening: Promise<void> | null = null;
  private snapshot: AgentSessionSnapshot | null = null;
  private activeOpen = 0;
  private nextSubscription = 0;
  private readonly artifacts: WebAcpArtifactDiscovery;

  constructor(private readonly input: WebAcpSessionRuntimeInput) {
    this.sessionId = `web-acp:${input.podKey}`;
    this.subscriptionBase = `agent-workspace-${input.paneId}`;
    this.artifacts = new WebAcpArtifactDiscovery(input.podKey, () => this.deps(), () => this.notify());
  }

  async open(sessionId: string): Promise<void> {
    this.assertSession(sessionId);
    if (this.cleanup || this.opening) return this.opening ?? Promise.resolve();
    const deps = this.deps();
    const openId = ++this.nextSubscription;
    const subscriptionId = `${this.subscriptionBase}-${openId}`;
    this.activeOpen = openId;
    const sharedSubscription = retainWebAcpPodSubscription(
      deps,
      this.input.podKey,
      subscriptionId,
      {
        onStatus: (status) => {
          this.connection = relayConnection(status);
          this.notify();
        },
      },
    );
    const cleanups = [
      deps.subscribeSession(() => {
        this.artifacts.observe(deps.readSession(this.input.podKey)?.state ?? "idle");
        this.notify();
      }),
      sharedSubscription.release,
    ];
    this.cleanup = () => cleanups.splice(0).forEach((cleanup) => cleanup());
    const opening = sharedSubscription.ready
      .then(async () => {
        if (this.activeOpen !== openId) return;
        this.error = null;
        await this.artifacts.open(deps.readSession(this.input.podKey)?.state ?? "idle");
      })
      .catch((cause: unknown) => {
        if (this.activeOpen !== openId) return;
        this.connection = "disconnected";
        this.error = cause instanceof Error ? cause.message : String(cause);
        this.notify();
      })
      .finally(() => {
        if (this.opening === opening) this.opening = null;
      });
    this.opening = opening;
    return opening;
  }

  close(sessionId: string): void {
    this.assertSession(sessionId);
    this.cleanup?.();
    this.cleanup = null;
    this.activeOpen = 0;
    this.opening = null;
    this.connection = "disconnected";
    this.notify();
  }

  getSnapshot(sessionId: string): AgentSessionSnapshot {
    this.assertSession(sessionId);
    this.snapshot ??= {
      ...projectWebAcpSnapshot({
        agentLabel: this.input.agentLabel,
        connection: this.connection,
        sessionId: this.sessionId,
        session: this.deps().readSession(this.input.podKey) ?? EMPTY_SESSION,
        title: this.input.title,
        workspaceArtifacts: this.artifacts.items(),
      }),
      error: this.error ?? this.artifacts.error,
    };
    return this.snapshot;
  }

  subscribe(sessionId: string, listener: () => void): () => void {
    this.assertSession(sessionId);
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  sendMessage(sessionId: string, commandId: string, input: { text: string }): Promise<void> {
    this.assertSession(sessionId);
    return this.send({ type: "prompt", prompt: input.text, requestId: commandId });
  }

  sendSlashCommand(
    sessionId: string,
    commandId: string,
    input: { name: string; arguments: string },
  ): Promise<void> {
    this.assertSession(sessionId);
    const prompt = `/${input.name}${input.arguments ? ` ${input.arguments}` : ""}`;
    return this.send({ type: "prompt", prompt, requestId: commandId });
  }

  interrupt(sessionId: string, commandId: string): Promise<void> {
    this.assertSession(sessionId);
    return this.send({ type: "interrupt", requestId: commandId });
  }

  async resolvePermission(
    sessionId: string,
    _commandId: string,
    permissionId: string,
    result: AgentPermissionResolution,
  ): Promise<void> {
    this.assertSession(sessionId);
    await this.send({
      type: "permission_response",
      requestId: permissionId,
      approved: result.action === "accept",
      ...(result.action === "accept" ? { updatedInput: result.content } : {}),
    });
    this.deps().removePermission(this.input.podKey, permissionId);
  }

  async updateConfiguration(
    sessionId: string,
    _commandId: string,
    patch: Record<string, unknown>,
  ): Promise<void> {
    this.assertSession(sessionId);
    const configuration =
      this.deps().readSession(this.input.podKey)?.configuration;
    if (!configuration?.supportedPermissionModes.length) {
      throw new Error("Agent does not expose configuration control");
    }
    if (typeof patch.permissionMode === "string") {
      await this.send({ type: "set_permission_mode", mode: patch.permissionMode });
    }
    if (typeof patch.model === "string") {
      await this.send({ type: "set_model", model: patch.model });
    }
  }

  loadOlder(): Promise<void> {
    return Promise.reject(new Error("Web ACP sessions do not expose paginated history"));
  }

  loadArtifact(sessionId: string, artifactId: string): Promise<Blob> {
    this.assertSession(sessionId);
    if (!artifactId.startsWith("workspace:")) {
      return Promise.reject(new Error("Unknown Web ACP artifact"));
    }
    return this.deps().loadWorkspaceArtifact(
      this.input.podKey,
      artifactId.slice("workspace:".length),
    );
  }

  private async send(command: Record<string, unknown>): Promise<void> {
    try {
      await this.deps().relay.sendAcpCommand(this.input.podKey, command);
    } catch (cause) {
      this.error = cause instanceof Error ? cause.message : String(cause);
      this.notify();
      throw cause;
    }
  }

  private notify() {
    this.snapshot = null;
    this.listeners.forEach((listener) => listener());
  }

  private assertSession(sessionId: string) {
    if (sessionId !== this.sessionId) throw new Error("Agent session reference mismatch");
  }

  private deps(): WebAcpRuntimeDeps {
    return this.input.deps ?? defaultWebAcpRuntimeDeps;
  }
}
