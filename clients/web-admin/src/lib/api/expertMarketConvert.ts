import type { ExpertMarketRelease as ProtoRelease } from "@proto/admin/v1/admin_pb";

import type {
  ExpertMarketRelease,
  ExpertMarketReleaseStatus,
} from "./adminExpertMarketTypes";

export function fromProtoExpertMarketRelease(
  release: ProtoRelease,
): ExpertMarketRelease {
  return {
    id: Number(release.id),
    application_id: Number(release.applicationId),
    source_expert_id: Number(release.sourceExpertId),
    publisher_organization_id: Number(release.publisherOrganizationId),
    publisher_user_id: Number(release.publisherUserId),
    version: release.version,
    status: release.status as ExpertMarketReleaseStatus,
    name: release.name,
    summary: release.summary,
    description: release.description,
    category: release.category,
    icon: release.icon,
    tags: release.tags,
    outcomes: release.outcomes,
    featured: release.featured,
    expert_snapshot_json: release.expertSnapshotJson,
    worker_spec_snapshot_json: release.workerSpecSnapshotJson,
    skill_dependencies_json: release.skillDependenciesJson,
    reviewer_user_id: optionalNumber(release.reviewerUserId),
    rejection_reason: release.rejectionReason,
    submitted_at: release.submittedAt,
    reviewed_at: release.reviewedAt,
    published_at: release.publishedAt,
    rejected_at: release.rejectedAt,
    withdrawn_at: release.withdrawnAt,
    created_at: release.createdAt,
  };
}

function optionalNumber(value: bigint | undefined): number | undefined {
  return value === undefined ? undefined : Number(value);
}
