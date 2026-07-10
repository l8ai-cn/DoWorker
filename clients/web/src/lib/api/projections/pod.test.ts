import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { PodSchema } from "@proto/pod/v1/pod_pb";
import { podToCache } from "./pod";

describe("podToCache", () => {
  it("preserves preview metadata for mobile access", () => {
    const pod = create(PodSchema, {
      podKey: "mobile-preview-pod",
      status: "running",
      previewPort: 3000,
      previewPath: "/app",
    });

    expect(podToCache(pod)).toMatchObject({
      preview_port: 3000,
      preview_path: "/app",
    });
  });
});
