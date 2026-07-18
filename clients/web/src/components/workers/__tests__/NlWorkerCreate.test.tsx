import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NlWorkerCreate } from "../NlWorkerCreate";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("NlWorkerCreate", () => {
  it("disables AI fill with an explicit state when no generation model is compatible", () => {
    render(
      <NlWorkerCreate
        prompt="Configure the worker"
        filling={false}
        generationModelResourceId={0}
        generationModels={{ status: "ready", data: [] }}
        onGenerationModelChange={vi.fn()}
        onPromptChange={vi.fn()}
        onFill={vi.fn()}
      />,
    );

    expect(screen.getByText("workers.create.nl.noGenerationModels")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "workers.create.nl.submit" }),
    ).toBeDisabled();
  });
});
