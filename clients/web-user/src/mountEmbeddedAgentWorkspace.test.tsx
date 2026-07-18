import { act } from "react";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";

const renderedClients = vi.hoisted(() => [] as unknown[]);
const actEnvironment = globalThis as typeof globalThis & {
  IS_REACT_ACT_ENVIRONMENT?: boolean;
};
const previousActEnvironment = actEnvironment.IS_REACT_ACT_ENVIRONMENT;

beforeAll(() => {
  actEnvironment.IS_REACT_ACT_ENVIRONMENT = true;
});

afterAll(() => {
  actEnvironment.IS_REACT_ACT_ENVIRONMENT = previousActEnvironment;
});

vi.mock("./embed-session/EmbeddedAgentWorkspace", () => ({
  EmbeddedAgentWorkspace: ({
    client,
    sessionId,
  }: {
    client: unknown;
    sessionId: string;
  }) => {
    renderedClients.push(client);
    return <output data-session-id={sessionId}>{sessionId}</output>;
  },
}));

import type { EmbedSessionClient } from "./embed-session-api";
import { mountEmbeddedAgentWorkspace } from "./mountEmbeddedAgentWorkspace";

describe("mountEmbeddedAgentWorkspace", () => {
  it("mounts two isolated atomic workspaces on the same page", () => {
    const firstElement = document.createElement("div");
    const secondElement = document.createElement("div");
    document.body.append(firstElement, secondElement);
    const firstClient = {} as EmbedSessionClient;
    const secondClient = {} as EmbedSessionClient;
    let first: ReturnType<typeof mountEmbeddedAgentWorkspace>;
    let second: ReturnType<typeof mountEmbeddedAgentWorkspace>;

    act(() => {
      first = mountEmbeddedAgentWorkspace(firstElement, {
        client: firstClient,
        sessionId: "session-1",
      });
      second = mountEmbeddedAgentWorkspace(secondElement, {
        client: secondClient,
        sessionId: "session-2",
      });
    });

    expect(firstElement).toHaveTextContent("session-1");
    expect(secondElement).toHaveTextContent("session-2");
    expect(renderedClients).toContain(firstClient);
    expect(renderedClients).toContain(secondClient);

    act(() => {
      first.unmount();
      second.unmount();
    });
    firstElement.remove();
    secondElement.remove();
  });
});
