import { Image } from "lucide-react";
import { fireEvent, render, screen } from "@testing-library/react";

import type { AgentToolRendererRegistration } from "./react/rendererTypes";
import { ToolRendererRegistry } from "./registry/ToolRendererRegistry";
import { ToolActivityCard } from "./ToolActivityCard";

describe("ToolActivityCard", () => {
  it("uses presentation only for the exact renderer identity", () => {
    const registry =
      new ToolRendererRegistry<AgentToolRendererRegistration>();
    registry.register(
      identity("1"),
      {
        detail: () => <div>Registered detail</div>,
        presentation: {
          icon: Image,
          inputLabel: "Prompt",
          label: "Image task",
          outputLabel: "Image result",
        },
        summary: () => <div>Registered summary</div>,
      },
      "builtin.image",
    );

    const { rerender } = render(
      <ToolActivityCard
        item={tool({ input: "Create a product image", output: "Done" })}
        renderers={registry}
      />,
    );

    expect(screen.getByText("Image task")).toBeVisible();
    expect(screen.getByText("Registered summary")).toBeVisible();
    fireEvent.click(screen.getByText("Details"));
    expect(screen.getByText("Registered detail")).toBeVisible();
    expect(screen.getByText("Prompt")).toBeVisible();
    expect(screen.getByText("Image result")).toBeVisible();

    rerender(
      <ToolActivityCard
        item={tool({ identity: identity("2"), input: "raw input" })}
        renderers={registry}
      />,
    );

    expect(screen.queryByText("Registered summary")).not.toBeInTheDocument();
    expect(screen.getByTestId("unsupported-tool-preview")).toBeVisible();
    expect(screen.getByText("agentsmesh.acp/generate_image@2")).toBeVisible();
  });

  it("does not infer specialized UI from an unregistered title", () => {
    render(
      <ToolActivityCard
        item={tool({
          identity: {
            namespace: "agentsmesh.acp",
            semanticKey: "image.generate",
            schemaVersion: "1",
          },
          input: "raw input",
          inputValue: { prompt: "a chart" },
          output: "raw output",
          results: [{ id: "result-1", kind: "data", value: { asset: "x" } }],
          title: "shell",
        })}
      />,
    );

    expect(screen.getByText("image.generate")).toBeVisible();
    expect(screen.getByTestId("unsupported-tool-preview")).toBeVisible();
    expect(screen.queryByText("Command")).not.toBeInTheDocument();

    fireEvent.click(screen.getByText("Details"));

    expect(screen.getByText("raw input")).toBeVisible();
    expect(screen.getByText("raw output")).toBeVisible();
    expect(screen.getByText("Raw tool evidence")).toBeVisible();
    expect(screen.getByText(/"title": "shell"/)).toBeVisible();
    expect(screen.getByText(/"asset": "x"/)).toBeVisible();
  });

  it("opens failed raw evidence immediately", () => {
    render(
      <ToolActivityCard
        item={tool({ input: "pnpm test", output: "exit status 1", status: "failed" })}
      />,
    );

    expect(screen.getByText("Details").closest("details")).toHaveAttribute("open");
  });
});

function identity(schemaVersion: string) {
  return {
    namespace: "agentsmesh.acp",
    semanticKey: "generate_image",
    schemaVersion,
  };
}

function tool(
  patch: Partial<Parameters<typeof ToolActivityCard>[0]["item"]> = {},
): Parameters<typeof ToolActivityCard>[0]["item"] {
  return {
    id: "tool-1",
    identity: identity("1"),
    kind: "tool",
    results: [],
    status: "completed",
    title: "generate_image",
    ...patch,
  };
}
