import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const close = vi.hoisted(() => vi.fn());
const factory = vi.hoisted(() => vi.fn());
const rendered = vi.hoisted(() => [] as unknown[]);

vi.mock("@do-worker/agent-ui", () => ({
  AgentWorkspace: (props: unknown) => {
    rendered.push(props);
    return <div>Shared Agent Workspace</div>;
  },
  createBuiltinContentRenderers: () => "builtin-renderers",
}));

vi.mock("./createEmbeddedAgentWorkbenchRuntime", () => ({
  createEmbeddedAgentWorkbenchRuntime: factory,
}));

import { EmbeddedAgentWorkspace } from "./EmbeddedAgentWorkspace";

describe("EmbeddedAgentWorkspace", () => {
  it("通过共享 factory 创建 V2 runtime，不接收 legacy client", async () => {
    const runtime = { close };
    const terminalRuntime = {};
    factory.mockResolvedValue({ runtime, terminalRuntime });
    const access = {
      baseUrl: "https://api.example.test",
      getAccessToken: () => "token",
      orgSlug: "acme",
      sessionId: "session-1",
    };

    const view = render(<EmbeddedAgentWorkspace access={access} />);

    expect(screen.getByRole("status")).toHaveTextContent("正在连接 Agent Workspace…");
    expect(await screen.findByText("Shared Agent Workspace")).toBeInTheDocument();
    expect(factory).toHaveBeenCalledWith(access, { fetch: undefined });
    expect(rendered.at(-1)).toMatchObject({
      contentRenderers: "builtin-renderers",
      locale: "zh-CN",
      runtime,
      sessionId: "session-1",
      terminalRuntime,
    });

    view.unmount();
    await waitFor(() => expect(close).toHaveBeenCalledWith("session-1"));
  });
});
