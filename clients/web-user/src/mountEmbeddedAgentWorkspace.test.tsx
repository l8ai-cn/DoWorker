import { act } from "react";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";

const renderedAccess = vi.hoisted(() => [] as unknown[]);
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
  EmbeddedAgentWorkspace: ({ access }: { access: { sessionId: string } }) => {
    renderedAccess.push(access);
    return <output data-session-id={access.sessionId}>{access.sessionId}</output>;
  },
}));

import { mountEmbeddedAgentWorkspace } from "./mountEmbeddedAgentWorkspace";

describe("mountEmbeddedAgentWorkspace", () => {
  it("用显式 access 挂载两个隔离工作区", () => {
    const firstElement = document.createElement("div");
    const secondElement = document.createElement("div");
    document.body.append(firstElement, secondElement);
    const firstAccess = {
      baseUrl: "https://api.example.test",
      getAccessToken: () => "token-1",
      orgSlug: "acme",
      sessionId: "session-1",
    };
    const secondAccess = {
      baseUrl: "https://api.example.test",
      getAccessToken: () => "token-2",
      orgSlug: "acme",
      sessionId: "session-2",
    };
    let first: ReturnType<typeof mountEmbeddedAgentWorkspace>;
    let second: ReturnType<typeof mountEmbeddedAgentWorkspace>;

    act(() => {
      first = mountEmbeddedAgentWorkspace(firstElement, {
        access: firstAccess,
      });
      second = mountEmbeddedAgentWorkspace(secondElement, {
        access: secondAccess,
      });
    });

    expect(firstElement).toHaveTextContent("session-1");
    expect(secondElement).toHaveTextContent("session-2");
    expect(firstElement).toHaveClass("agent-cloud-app");
    expect(secondElement).toHaveClass("agent-cloud-app");
    expect(renderedAccess).toEqual([firstAccess, secondAccess]);

    act(() => {
      first.unmount();
      second.unmount();
    });
    expect(firstElement).not.toHaveClass("agent-cloud-app");
    expect(secondElement).not.toHaveClass("agent-cloud-app");
    firstElement.remove();
    secondElement.remove();
  });
});
