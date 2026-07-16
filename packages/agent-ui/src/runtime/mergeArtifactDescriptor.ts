import { clone } from "@bufbuild/protobuf";

import {
  ArtifactDescriptorSchema,
  ArtifactRevisionSchema,
  type ArtifactDescriptor,
  type ArtifactRevision,
} from "@do-worker/proto/agent_workbench/v2/artifact_pb";

export function mergeArtifactDescriptor(
  current: ArtifactDescriptor | undefined,
  value: ArtifactDescriptor,
): ArtifactDescriptor {
  const next = clone(ArtifactDescriptorSchema, value);
  if (!current) return next;

  const revisions = new Map<string, ArtifactRevision>();
  for (const revision of [...current.revisions, ...value.revisions]) {
    revisions.set(
      revision.revision.toString(),
      clone(ArtifactRevisionSchema, revision),
    );
  }
  next.revisions = [...revisions.values()].sort((left, right) =>
    left.revision < right.revision ? -1 : left.revision > right.revision ? 1 : 0,
  );
  return next;
}
