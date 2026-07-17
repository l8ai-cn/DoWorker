import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import { SessionConfigurationSchema } from "@do-worker/proto/agent_workbench/v2/configuration_pb";
import {
  AgentEventSchema,
  ConfigurationChangedSchema,
  EventEnvelopeSchema,
  SessionSnapshotSchema,
  SupportCapabilitiesSchema,
} from "@do-worker/proto/agent_workbench/v2/session_pb";
import { SessionStatus } from "@do-worker/proto/agent_workbench/v2/session_state_pb";

import { projectGeneratedSessionSnapshot } from "./projectGeneratedSessionSnapshot";
import { applyAgentEvent } from "./agentSessionEventReducer";

describe("generated session configuration projection", () => {
  it("projects only explicit current values with advertised options", () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      status: SessionStatus.IDLE,
      capabilities: create(SupportCapabilitiesSchema, {
        models: ["gpt-5.4", "gpt-5.4-mini"],
        permissionModes: ["ask_dangerous", "bypass"],
      }),
      configuration: create(SessionConfigurationSchema, {
        model: "gpt-5.4",
        permissionMode: "ask_dangerous",
      }),
    });

    const projected = projectGeneratedSessionSnapshot(snapshot, {
      title: "Session",
      agentLabel: "Codex",
      connection: "connected",
      interactionMode: "acp",
      hasOlderItems: false,
    });

    expect(projected.configuration).toEqual([
      {
        id: "model",
        label: "Model",
        value: "gpt-5.4",
        options: [
          { value: "gpt-5.4", label: "gpt-5.4" },
          { value: "gpt-5.4-mini", label: "gpt-5.4-mini" },
        ],
      },
      {
        id: "permission_mode",
        label: "Permissions",
        value: "ask_dangerous",
        options: [
          { value: "ask_dangerous", label: "ask_dangerous" },
          { value: "bypass", label: "bypass" },
        ],
      },
    ]);
  });

  it("does not invent a current value from the first supported option", () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      status: SessionStatus.IDLE,
      capabilities: create(SupportCapabilitiesSchema, {
        models: ["gpt-5.4"],
      }),
    });

    const projected = projectGeneratedSessionSnapshot(snapshot, {
      title: "Session",
      agentLabel: "Codex",
      connection: "connected",
      interactionMode: "acp",
      hasOlderItems: false,
    });

    expect(projected.configuration).toEqual([
      {
        id: "model",
        label: "Model",
        value: "",
        options: [{ value: "gpt-5.4", label: "gpt-5.4" }],
      },
    ]);
  });

  it("applies configuration changed events to the canonical snapshot", () => {
    const snapshot = create(SessionSnapshotSchema, {
      sessionId: "session-1",
      streamEpoch: "epoch-1",
      status: SessionStatus.IDLE,
    });
    const event = create(AgentEventSchema, {
      envelope: create(EventEnvelopeSchema, {
        sessionId: "session-1",
        streamEpoch: "epoch-1",
        revision: 1n,
        sequence: 1n,
        itemId: "configuration:1",
        createdAt: "2026-07-16T00:00:00Z",
      }),
      event: {
        case: "configurationChanged",
        value: create(ConfigurationChangedSchema, {
          configuration: create(SessionConfigurationSchema, {
            model: "gpt-5.4",
            permissionMode: "bypass",
          }),
        }),
      },
    });

    applyAgentEvent(snapshot, event);

    expect(snapshot.configuration).toMatchObject({
      model: "gpt-5.4",
      permissionMode: "bypass",
    });
  });
});
