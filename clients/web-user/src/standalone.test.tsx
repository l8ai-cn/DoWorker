import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("./embed", () => ({
  DoWorkerApp: ({ basename }: { basename?: string }) => (
    <output data-testid="worker-app">{basename ?? "root"}</output>
  ),
}));

import { DoWorkerStandaloneApp } from "./standalone";

describe("DoWorkerStandaloneApp", () => {
  it("provides a router and forwards Worker props", () => {
    render(<DoWorkerStandaloneApp basename="/worker" isDarkMode />);

    expect(screen.getByTestId("worker-app")).toHaveTextContent("/worker");
  });
});
