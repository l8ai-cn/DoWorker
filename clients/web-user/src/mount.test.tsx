import { act } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("./standalone", () => ({
  AgentCloudStandaloneApp: () => <output data-testid="worker-app">mounted</output>,
}));

import { mountAgentCloudApp } from "./mount";

describe("mountAgentCloudApp", () => {
  it("mounts the Worker and removes it on unmount", () => {
    const element = document.createElement("div");
    document.body.append(element);
    let mounted: ReturnType<typeof mountAgentCloudApp>;
    act(() => {
      mounted = mountAgentCloudApp(element);
    });

    expect(element.querySelector("[data-testid=worker-app]")).toHaveTextContent("mounted");

    act(() => {
      mounted.unmount();
    });
    expect(element.innerHTML).toBe("");
    element.remove();
  });
});
