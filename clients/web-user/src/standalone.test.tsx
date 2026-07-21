import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("./embed", () => ({
  AgentCloudApp: ({ basename }: { basename?: string }) => (
    <output data-testid="worker-app">{basename ?? "root"}</output>
  ),
}));

import { AgentCloudStandaloneApp } from "./standalone";

describe("AgentCloudStandaloneApp", () => {
  it("provides a router and forwards Worker props", () => {
    render(<AgentCloudStandaloneApp basename="/worker" isDarkMode />);

    expect(screen.getByTestId("worker-app")).toHaveTextContent("/worker");
  });
});
