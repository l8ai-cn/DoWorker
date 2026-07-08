import { describe, expect, it } from "vitest";
import type { Host } from "@/hooks/useHosts";
import { hostDisplayLabel } from "./hostDisplayLabel";

function host(partial: Partial<Host> & Pick<Host, "host_id" | "name">): Host {
  return {
    owner: "me",
    status: "online",
    ...partial,
  };
}

describe("hostDisplayLabel", () => {
  it("hides internal admin-workspace runner host names", () => {
    expect(
      hostDisplayLabel(
        host({ host_id: "host_admin-workspace-do-agent", name: "admin-workspace-do-agent" }),
      ),
    ).toBe("Workspace");
  });

  it("prefers this machine for the local desktop host", () => {
    expect(
      hostDisplayLabel(host({ host_id: "host_laptop", name: "laptop" }), {
        thisMachineHostId: "host_laptop",
      }),
    ).toBe("This machine");
  });
});
