import {
  AgentSessionRuntimeV2,
  type AgentAttachmentReference,
  type AgentSessionRuntimeV2Options,
} from "@do-worker/agent-ui";

export class EmbeddedAgentSessionRuntime extends AgentSessionRuntimeV2 {
  constructor(
    options: AgentSessionRuntimeV2Options,
    private readonly upload: (file: File) => Promise<AgentAttachmentReference>,
    private readonly embeddedSessionId: string,
  ) {
    super(options);
  }

  uploadAttachment(sessionId: string, file: File): Promise<AgentAttachmentReference> {
    if (sessionId !== this.embeddedSessionId) {
      throw new Error("agent_workbench_runtime_session_mismatch");
    }
    return this.upload(file);
  }
}
