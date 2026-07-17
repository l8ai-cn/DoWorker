import { expectTypeOf, vi } from "vitest";

import { ArtifactController } from "./ArtifactController";
import type {
  ArtifactDescriptor,
  ArtifactRuntime,
} from "./ArtifactRuntime";
import { artifactAction } from "./artifactAction";

interface Receipt {
  clientActionId: string;
  revision: bigint;
}

function descriptor(
  patch: Partial<ArtifactDescriptor> = {},
): ArtifactDescriptor {
  return {
    artifactId: "image-1",
    filename: "image.png",
    grants: ["read", "download", "edit_image"],
    mimeType: "image/png",
    representationId: "source",
    revision: 3n,
    ...patch,
  };
}

function createRuntime() {
  const loadRepresentation = vi.fn(
    async (
      artifactId: string,
      representationId: string,
      revision: bigint,
    ) => new Blob([`${artifactId}:${representationId}:${revision}`]),
  );
  const executeAction = vi.fn(
    async (command: { clientActionId: string }): Promise<Receipt> => ({
      clientActionId: command.clientActionId,
      revision: 4n,
    }),
  );
  const runtime: ArtifactRuntime<Blob, Receipt> = {
    download: vi.fn(async () => undefined),
    executeAction,
    loadRepresentation,
    subscribe: vi.fn(() => () => undefined),
  };
  return { executeAction, loadRepresentation, runtime };
}

describe("ArtifactController", () => {
  it("caches representations by artifact, representation, and revision", async () => {
    const { loadRepresentation, runtime } = createRuntime();
    const controller = new ArtifactController(runtime);

    const first = controller.loadRepresentation("artifact-1", "source", 1n);
    const duplicate = controller.loadRepresentation(
      "artifact-1",
      "source",
      1n,
    );
    const otherArtifact = controller.loadRepresentation(
      "artifact-2",
      "source",
      1n,
    );
    const otherRepresentation = controller.loadRepresentation(
      "artifact-1",
      "preview",
      1n,
    );
    const otherRevision = controller.loadRepresentation(
      "artifact-1",
      "source",
      2n,
    );

    expect(first).toBe(duplicate);
    await Promise.all([
      first,
      duplicate,
      otherArtifact,
      otherRepresentation,
      otherRevision,
    ]);
    expect(loadRepresentation).toHaveBeenCalledTimes(4);
  });

  it("rejects stale actions without calling the runtime", async () => {
    const { executeAction, runtime } = createRuntime();
    const controller = new ArtifactController(runtime);
    controller.updateDescriptor(descriptor({ revision: 4n }));

    await expect(
      controller.execute(
        artifactAction({
          actionType: "edit_image",
          artifactId: "image-1",
          baseRevision: 3n,
          clientActionId: "action-1",
          payload: {},
          representationId: "source",
        }),
      ),
    ).rejects.toThrow("artifact_revision_conflict");
    expect(executeAction).not.toHaveBeenCalled();
  });

  it("reuses the runtime result for the same client action id", async () => {
    const { executeAction, runtime } = createRuntime();
    const controller = new ArtifactController(runtime);
    controller.updateDescriptor(descriptor());
    const command = artifactAction({
      actionType: "edit_image",
      artifactId: "image-1",
      baseRevision: 3n,
      clientActionId: "action-1",
      payload: { instruction: "remove background" },
      representationId: "source",
    });

    const first = controller.execute(command);
    const duplicate = controller.execute(
      artifactAction({
        ...command,
        payload: { instruction: "remove background" },
      }),
    );
    expect(first).toBe(duplicate);
    expect(await duplicate).toEqual({
      clientActionId: "action-1",
      revision: 4n,
    });

    controller.updateDescriptor(descriptor({ revision: 4n }));
    expect(await controller.execute(command)).toBe(await first);
    expect(executeAction).toHaveBeenCalledTimes(1);
  });

  it("rejects a changed command that reuses a client action id", async () => {
    const { executeAction, runtime } = createRuntime();
    const controller = new ArtifactController(runtime);
    controller.updateDescriptor(descriptor());
    const command = artifactAction({
      actionType: "edit_image",
      artifactId: "image-1",
      baseRevision: 3n,
      clientActionId: "action-1",
      payload: { instruction: "remove background" },
      representationId: "source",
    });

    await controller.execute(command);

    await expect(
      controller.execute(
        artifactAction({
          ...command,
          payload: { instruction: "crop image" },
        }),
      ),
    ).rejects.toThrow("artifact_action_id_conflict");
    expect(executeAction).toHaveBeenCalledTimes(1);
  });

  it("waits for an external descriptor update after an action", async () => {
    const { runtime } = createRuntime();
    const controller = new ArtifactController(runtime);
    const original = descriptor();
    controller.updateDescriptor(original);

    await controller.execute(
      artifactAction({
        actionType: "edit_image",
        artifactId: "image-1",
        baseRevision: 3n,
        clientActionId: "action-1",
        payload: {},
        representationId: "source",
      }),
    );

    const current = controller.getDescriptor("image-1", "source");
    expect(current).toBe(original);
    expect(current?.revision).toBe(3n);
    expectTypeOf(current?.revision).toEqualTypeOf<bigint | undefined>();
  });
});
