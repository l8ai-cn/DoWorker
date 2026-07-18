import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  SemanticChangeOperation,
  SemanticChangeSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { render, screen } from "@/test/test-utils";
import { ResourceSemanticDiff } from "./ResourceSemanticDiff";

describe("ResourceSemanticDiff", () => {
  it("shows digest prefixes and redacts embedded values", () => {
    render(
      <ResourceSemanticDiff
        emptyLabel="No changes"
        changes={[
          create(SemanticChangeSchema, {
            operation: SemanticChangeOperation.REPLACE,
            path: "/spec/modelRef",
            before: {
              value: {
                case: "digest",
                value: "sha256:0123456789abcdef",
              },
            },
            after: {
              value: {
                case: "redactedJson",
                value: new TextEncoder().encode('{"token":"do-not-render"}'),
              },
            },
          }),
        ]}
      />,
    );

    expect(screen.getByText("/spec/modelRef")).toBeInTheDocument();
    expect(screen.getByText(/sha256:012.*redacted/)).toBeInTheDocument();
    expect(screen.queryByText(/do-not-render/)).not.toBeInTheDocument();
  });

  it("renders the provided empty state", () => {
    render(<ResourceSemanticDiff changes={[]} emptyLabel="No semantic changes" />);

    expect(screen.getByText("No semantic changes")).toBeInTheDocument();
  });

  it("keeps the full semantic path available", () => {
    const path = "/spec/workspace/configDocumentBindings/0/configBundleRef";
    render(
      <ResourceSemanticDiff
        emptyLabel="No changes"
        changes={[create(SemanticChangeSchema, {
          operation: SemanticChangeOperation.ADD,
          path,
        })]}
      />,
    );

    expect(screen.getByText(path)).toHaveAttribute("title", path);
  });
});
