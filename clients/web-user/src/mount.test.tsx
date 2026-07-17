import { act } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("./standalone", () => ({
  DoWorkerStandaloneApp: () => <output data-testid="worker-app">mounted</output>,
}));

import { mountDoWorkerApp } from "./mount";

describe("mountDoWorkerApp", () => {
  it("mounts the Worker and removes it on unmount", () => {
    const element = document.createElement("div");
    document.body.append(element);
    let mounted: ReturnType<typeof mountDoWorkerApp>;
    act(() => {
      mounted = mountDoWorkerApp(element);
    });

    expect(element.querySelector("[data-testid=worker-app]")).toHaveTextContent("mounted");

    act(() => {
      mounted.unmount();
    });
    expect(element.innerHTML).toBe("");
    element.remove();
  });
});
