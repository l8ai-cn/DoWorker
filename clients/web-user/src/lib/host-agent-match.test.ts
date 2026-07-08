import { describe, expect, it } from "vitest";
import type { AvailableAgent } from "@/hooks/useAvailableAgents";
import type { Host } from "@/hooks/useHosts";
import { hostSupportsAgent, pickOnlineHostForAgent, collapseInternalWorkspaceHostsForPicker } from "./host-agent-match";

const agent = (id: string, harness: string): AvailableAgent => ({
  id,
  name: id,
  display_name: id,
  description: null,
  harness,
  skills: [],
});

const host = (
  id: string,
  configured?: Record<string, boolean | string> | null,
): Host => ({
  host_id: `host_${id}`,
  name: id,
  owner: "org",
  status: "online",
  configured_harnesses: configured,
});

describe("hostSupportsAgent", () => {
  it("matches by agent id when configured_harnesses lists runner slugs", () => {
    const h = host("admin-workspace-runner", { "e2e-echo": true });
    expect(hostSupportsAgent(h, agent("e2e-echo", "e2e-mock-agent"))).toBe(true);
    expect(hostSupportsAgent(h, agent("do-agent", "do-agent"))).toBe(false);
  });

  it("assumes support when configured_harnesses is absent", () => {
    expect(hostSupportsAgent(host("legacy"), agent("do-agent", "do-agent"))).toBe(true);
  });
});

describe("pickOnlineHostForAgent", () => {
  const hosts = [
    host("admin-workspace-runner", { "e2e-echo": true }),
    host("admin-workspace-do-agent", { "do-agent": true }),
  ];

  it("picks the host that advertises the agent slug", () => {
    const picked = pickOnlineHostForAgent(hosts, agent("do-agent", "do-agent"));
    expect(picked?.host_id).toBe("host_admin-workspace-do-agent");
  });

  it("returns undefined when no online host supports the agent", () => {
    const picked = pickOnlineHostForAgent(hosts, agent("codex-cli", "codex"));
    expect(picked).toBeUndefined();
  });
});

describe("collapseInternalWorkspaceHostsForPicker", () => {
  it("shows one Workspace row when several admin-workspace runners are online", () => {
    const collapsed = collapseInternalWorkspaceHostsForPicker([
      host("admin-workspace-runner", { "e2e-echo": true }),
      host("admin-workspace-do-agent", { "do-agent": true }),
      host("dev-runner-codex", { "codex-cli": true }),
    ]);
    expect(collapsed).toHaveLength(2);
    expect(collapsed.map((h) => h.name).sort()).toEqual(
      ["admin-workspace-runner", "dev-runner-codex"].sort(),
    );
  });

  it("picks the do-agent runner when that agent is selected", () => {
    const collapsed = collapseInternalWorkspaceHostsForPicker(
      [
        host("admin-workspace-runner", { "e2e-echo": true }),
        host("admin-workspace-do-agent", { "do-agent": true }),
      ],
      agent("do-agent", "do-agent"),
    );
    expect(collapsed).toHaveLength(1);
    expect(collapsed[0]?.host_id).toBe("host_admin-workspace-do-agent");
  });
});
