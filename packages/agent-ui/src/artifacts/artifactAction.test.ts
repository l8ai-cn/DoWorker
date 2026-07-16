import { expectTypeOf } from "vitest";

import type { ArtifactDescriptor } from "./ArtifactRuntime";
import { artifactAction } from "./artifactAction";

it("defines the required artifact descriptor contract", () => {
  const value: ArtifactDescriptor<"read" | "edit_image"> = {
    artifactId: "image-1",
    filename: "image.png",
    grants: ["read", "edit_image"],
    mimeType: "image/png",
    representationId: "source",
    revision: 3n,
  };

  expect(value).toEqual({
    artifactId: "image-1",
    filename: "image.png",
    grants: ["read", "edit_image"],
    mimeType: "image/png",
    representationId: "source",
    revision: 3n,
  });
  expectTypeOf(value.revision).toEqualTypeOf<bigint>();
  expectTypeOf(value.grants).toEqualTypeOf<
    readonly ("read" | "edit_image")[]
  >();
});

it("preserves typed action fields without changing the payload", () => {
  const payload = {
    instruction: "remove background",
    normalizedRegion: { height: 0.5, width: 0.5, x: 0.25, y: 0.25 },
  };
  const action = artifactAction({
    actionType: "edit_image",
    artifactId: "image-1",
    baseRevision: 3n,
    clientActionId: "action-1",
    payload,
    representationId: "source",
  });

  expect(action).toEqual({
    actionType: "edit_image",
    artifactId: "image-1",
    baseRevision: 3n,
    clientActionId: "action-1",
    payload,
    representationId: "source",
  });
  expect(action.payload).toBe(payload);
  expectTypeOf(action.actionType).toEqualTypeOf<"edit_image">();
  expectTypeOf(action.baseRevision).toEqualTypeOf<bigint>();
  expectTypeOf(action.payload).toEqualTypeOf<typeof payload>();
});
