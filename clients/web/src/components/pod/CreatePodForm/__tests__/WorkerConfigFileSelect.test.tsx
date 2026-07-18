import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createEnvBundle } from "@/lib/api/facade/envBundleConnect";
import { WorkerConfigFileSelect } from "../WorkerConfigFileSelect";

vi.mock("@/lib/api/facade/envBundleConnect", () => ({
  createEnvBundle: vi.fn(),
}));

const mockCreateEnvBundle = vi.mocked(createEnvBundle);

describe("WorkerConfigFileSelect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateEnvBundle.mockResolvedValue({
      id: 9n,
      name: "openclaw.json",
      agentSlug: "openclaw",
      kind: "config",
      kindPrimary: false,
    } as never);
  });

  it("binds the selected config bundle to the declared document", () => {
    const onChange = vi.fn();
    render(
      <WorkerConfigFileSelect
        agentSlug="openclaw"
        requirements={[{
          document_id: "openclaw-json",
          format: "json",
          target_path: "openclaw-home/.openclaw/openclaw.json",
        }]}
        bundles={[
          bundle(1, "base.json"),
          bundle(2, "overlay.json"),
        ]}
        bindings={[{
          document_id: "openclaw-json",
          config_bundle_id: 1,
        }]}
        onChange={onChange}
        t={(key) => key}
      />,
    );

    fireEvent.click(screen.getByRole("checkbox", { name: /overlay\.json/ }));

    expect(onChange).toHaveBeenCalledWith([{
      document_id: "openclaw-json",
      config_bundle_id: 2,
    }]);
  });

  it("validates and stores an uploaded JSON object", async () => {
    const onChange = vi.fn();
    render(
      <WorkerConfigFileSelect
        agentSlug="openclaw"
        requirements={[{
          document_id: "openclaw-json",
          format: "json",
          target_path: "openclaw-home/.openclaw/openclaw.json",
        }]}
        bundles={[]}
        bindings={[]}
        onChange={onChange}
        t={(key) => key}
      />,
    );

    fireEvent.change(screen.getByLabelText("ide.createPod.uploadConfigFile"), {
      target: {
        files: [
          new File(
            ['{"gateway":{"enabled":true}}'],
            "openclaw.json",
            { type: "application/json" },
          ),
        ],
      },
    });

    await waitFor(() => expect(mockCreateEnvBundle).toHaveBeenCalledWith({
      agentSlug: "openclaw",
      name: "openclaw.json",
      description: "ide.createPod.configFileUploaded",
      kind: "config",
      data: { __json: '{"gateway":{"enabled":true}}' },
    }));
    expect(onChange).toHaveBeenCalledWith([{
      document_id: "openclaw-json",
      config_bundle_id: 9,
    }]);
  });
});

function bundle(id: number, name: string) {
  return {
    id,
    name,
    agent_slug: "openclaw",
    kind: "config",
    kind_primary: false,
  };
}
