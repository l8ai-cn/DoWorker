import { describe, expect, it } from "vitest";
import { dispatchDoAgentRelayEvent } from "@/stores/doagentDispatcher";
import { useDoAgentConsoleStore } from "@/stores/doagentConsole";
import { MsgType } from "@/stores/relayProtocol";

describe("doagentDispatcher", () => {
  it("parses goal list from controlResponse", () => {
    useDoAgentConsoleStore.setState({ goalsByPod: {}, lastResponseKey: {} });
    dispatchDoAgentRelayEvent("pod-1", MsgType.AcpEvent, {
      type: "controlResponse",
      goals: [{ id: "g1", title: "Ship feature", status: "in_progress" }],
    });
    expect(useDoAgentConsoleStore.getState().goalsByPod["pod-1"]).toEqual([
      { id: "g1", title: "Ship feature", status: "in_progress" },
    ]);
  });

  it("returns false so ACP dispatch continues", () => {
    const handled = dispatchDoAgentRelayEvent("pod-1", MsgType.AcpEvent, {
      type: "controlResponse",
      goals: [],
    });
    expect(handled).toBe(false);
  });
});
