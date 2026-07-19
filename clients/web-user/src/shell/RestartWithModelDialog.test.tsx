import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { RestartWithModelDialog } from "./RestartWithModelDialog";

describe("RestartWithModelDialog", () => {
  it("does not offer a rejected model-only restart request", () => {
    render(
      <RestartWithModelDialog
        sessionId="conv_src"
        open
        onOpenChange={() => {}}
      />,
    );

    expect(screen.getByTestId("restart-model-unavailable-dialog")).toHaveTextContent(
      "当前协议不允许",
    );
    expect(screen.queryByTestId("restart-model-submit")).not.toBeInTheDocument();
  });
});
