import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ExpertMarketOperations } from "./ExpertMarketOperations";

vi.mock("./ExpertMarketSubmissionPanel", () => ({
  ExpertMarketSubmissionPanel: () => <div>User submission operations</div>,
}));
vi.mock("./ExpertMarketUpgradePanel", () => ({
  ExpertMarketUpgradePanel: () => <div>Installed expert operations</div>,
}));

describe("ExpertMarketOperations", () => {
  it("shows upgrade operations for an expert installed from the market", () => {
    renderOperations(true);

    expect(screen.getByText("Installed expert operations")).toBeInTheDocument();
  });

  it("shows submission operations for an authored expert", () => {
    renderOperations(false);

    expect(screen.getByText("User submission operations")).toBeInTheDocument();
  });
});

function renderOperations(installedFromMarket: boolean) {
  render(
    <ExpertMarketOperations
      expertID={1}
      expertSlug="video-editor"
      installedFromMarket={installedFromMarket}
      submissionReady
      onUpgraded={vi.fn()}
    />,
  );
}
