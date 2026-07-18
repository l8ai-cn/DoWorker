import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { PodSchema } from "@proto/pod/v1/pod_pb";
import { podToCache } from "./pod";

describe("podToCache", () => {
  it("preserves the canonical agent session id", () => {
    const pod = create(PodSchema, {
      podKey: "agent-pod",
      sessionId: "session-123",
      status: "running",
    });

    expect(podToCache(pod).session_id).toBe("session-123");
  });

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

  it("preserves worker skill slugs", () => {
    const pod = create(PodSchema, {
      podKey: "seedance-worker",
      status: "running",
      workerSkillSlugs: ["seedance-expert"],
    });

    expect(podToCache(pod).worker_skill_slugs).toEqual(["seedance-expert"]);
  });
});
