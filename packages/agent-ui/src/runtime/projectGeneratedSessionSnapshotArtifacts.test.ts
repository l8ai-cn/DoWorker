import { create } from "@bufbuild/protobuf";

import {
  ArtifactDescriptorSchema,
  ArtifactStatus,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";
import {
  createArtifactCatalog,
  projectArtifactDescriptor,
  projectArtifactReference,
} from "./projectGeneratedSessionSnapshotArtifacts";

describe("generated artifact projection", () => {
  it("preserves the selected representation, bigint revision, manifest, grants, and actions", () => {
    const descriptor = create(ArtifactDescriptorSchema, {
      artifactId: "image-1",
      revision: 9_007_199_254_740_993n,
      filename: "result.png",
      mediaType: "image/png",
      role: "image_edit",
      status: ArtifactStatus.READY,
      provenance: {
        producerId: "image-task-1",
        producerNamespace: "openai",
        producerType: "image.edit",
      },
      revisions: [
        {
          revision: 9_007_199_254_740_993n,
          provenance: { toolExecutionId: "workbench-publish:1" },
        },
      ],
      representations: [
        {
          representationId: "source",
          revision: 9_007_199_254_740_992n,
          mediaType: "image/png",
          role: "source",
          filename: "source.png",
          status: ArtifactStatus.READY,
          dimensions: { width: 1920, height: 1080 },
        },
        {
          representationId: "result",
          revision: 9_007_199_254_740_993n,
          mediaType: "image/webp",
          role: "result",
          filename: "result.webp",
          status: ArtifactStatus.READY,
          dimensions: { width: 1920, height: 1080 },
        },
      ],
      grants: [
        {
          grantId: "grant-edit",
          representationIds: ["source"],
          actions: ["image.edit"],
          minimumRevision: 9_007_199_254_740_993n,
          maximumRevision: 9_007_199_254_740_994n,
        },
        {
          grantId: "grant-download",
          actions: ["artifact.download"],
        },
      ],
      manifest: {
        manifest: {
          case: "imageEdit",
          value: {
            sourceRepresentationId: "source",
            resultRepresentationId: "result",
            candidateRepresentationIds: ["result"],
            sourceWidth: 1920,
            sourceHeight: 1080,
            regions: [{ x: 0.1, y: 0.2, width: 0.3, height: 0.4 }],
            annotations: [
              {
                annotationId: "selection-1",
                path: [{ x: 0.1, y: 0.2 }],
                label: "标题区域",
                style: {
                  mediaType: "application/json",
                  data: new TextEncoder().encode('{"color":"primary"}'),
                },
              },
            ],
          },
        },
      },
    });

    const projected = projectArtifactDescriptor(descriptor, "artifact-item", {
      representationId: "result",
    });

    expect(projected).toMatchObject({
      revision: 9_007_199_254_740_993n,
      selectedRepresentationId: "result",
      filename: "result.webp",
      mimeType: "image/webp",
      role: "image_edit",
      provenance: {
        publicationToolExecutionId: "workbench-publish:1",
        producerId: "image-task-1",
        producerNamespace: "openai",
        producerType: "image.edit",
      },
      actions: ["image.edit", "artifact.download"],
      representations: [
        {
          representationId: "source",
          revision: 9_007_199_254_740_992n,
          dimensions: { width: 1920, height: 1080 },
          status: "ready",
        },
        {
          representationId: "result",
          revision: 9_007_199_254_740_993n,
          status: "ready",
        },
      ],
      grants: [
        {
          grantId: "grant-edit",
          representationIds: ["source"],
          actions: ["image.edit"],
          minimumRevision: 9_007_199_254_740_993n,
          maximumRevision: 9_007_199_254_740_994n,
        },
        {
          grantId: "grant-download",
          representationIds: [],
          actions: ["artifact.download"],
        },
      ],
      manifest: {
        kind: "image_edit",
        sourceRepresentationId: "source",
        resultRepresentationId: "result",
        candidateRepresentationIds: ["result"],
        sourceDimensions: { width: 1920, height: 1080 },
        regions: [{ x: 0.1, y: 0.2, width: 0.3, height: 0.4 }],
        annotations: [
          {
            annotationId: "selection-1",
            path: [{ x: 0.1, y: 0.2 }],
            label: "标题区域",
            style: {
              mediaType: "application/json",
            },
          },
        ],
      },
    });
    expect(projected.revision).toBeTypeOf("bigint");
    expect(
      Array.from(
        projected.manifest?.kind === "image_edit"
          ? projected.manifest.annotations[0]?.style?.data ?? []
          : [],
      ),
    ).toEqual(
      Array.from(new TextEncoder().encode('{"color":"primary"}')),
    );
  });

  it("does not silently select another representation when a reference is invalid", () => {
    const descriptor = create(ArtifactDescriptorSchema, {
      artifactId: "image-1",
      revision: 1n,
      filename: "result.png",
      mediaType: "image/png",
      status: ArtifactStatus.READY,
      representations: [
        {
          representationId: "ready",
          revision: 1n,
          mediaType: "image/png",
          status: ArtifactStatus.READY,
        },
      ],
    });

    expect(
      projectArtifactReference(
        { artifactId: "image-1", representationId: "missing" },
        "artifact-item",
        createArtifactCatalog([descriptor]),
      ),
    ).toMatchObject({
      kind: "system",
      status: "failed",
      title: "Unsupported artifact",
      detail: expect.stringContaining("representationId=missing"),
    });

    expect(
      projectArtifactReference(
        { artifactId: "image-1", revision: 2n },
        "artifact-item",
        createArtifactCatalog([descriptor]),
      ),
    ).toMatchObject({
      kind: "system",
      status: "failed",
      detail: expect.stringContaining("revision=2"),
    });
  });
});
