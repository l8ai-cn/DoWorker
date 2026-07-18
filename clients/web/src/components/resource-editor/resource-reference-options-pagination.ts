import { listResources } from "@/lib/api/facade/orchestrationResource";
import type { EnvironmentBundlePurpose } from "@proto/orchestration_resource/v1/orchestration_resource_queries_pb";

const REFERENCE_PAGE_SIZE = 100;

interface EnvironmentBundleFilter {
  purpose: EnvironmentBundlePurpose;
  workerType: string;
  targetName?: string;
}

type ResourceFilterInput = {
  kind: string;
  environmentBundleFilter?: EnvironmentBundleFilter;
};

type ListResourceResponse<T> = {
  items: readonly T[];
  total?: bigint | number;
  limit?: bigint | number;
  offset?: bigint | number;
  appliedEnvironmentBundleFilter?: {
    purpose: EnvironmentBundlePurpose;
    workerType: string;
    targetName: string;
  };
};

export async function collectResourceReferenceOptions<T>(
  orgSlug: string,
  query: ResourceFilterInput,
  toOption: (resource: {
    identity?: { target?: { name?: string } };
    displayName: string;
    revision: bigint;
  }) => T,
  assertFilterApplied: (
    requested: EnvironmentBundleFilter | undefined,
    applied: {
      purpose: EnvironmentBundlePurpose;
      workerType: string;
      targetName: string;
    } | undefined,
  ) => void,
): Promise<T[]> {
  let offset = 0;
  const options: T[] = [];

  for (;;) {
    const response = await listResources(
      orgSlug,
      { kind: query.kind, limit: REFERENCE_PAGE_SIZE, offset, environmentBundleFilter: query.environmentBundleFilter },
    ) as ListResourceResponse<{
      identity?: { target?: { name?: string } };
      displayName: string;
      revision: bigint;
    }>;
    assertFilterApplied(
      query.environmentBundleFilter,
      response.appliedEnvironmentBundleFilter,
    );
    options.push(...response.items.map(toOption));

    const responseOffset = Number(response.offset ?? offset);
    const responseLimit = Number(response.limit ?? REFERENCE_PAGE_SIZE);
    const total = Number(response.total ?? 0);
    const nextOffset = responseOffset + responseLimit;
    if (
      !Number.isFinite(responseOffset) ||
      !Number.isFinite(responseLimit) ||
      !Number.isFinite(total) ||
      responseLimit <= 0 ||
      nextOffset >= total
    ) {
      return options;
    }
    if (nextOffset <= responseOffset) {
      throw new Error("The control plane returned an invalid pagination response.");
    }
    offset = nextOffset;
  }
}
