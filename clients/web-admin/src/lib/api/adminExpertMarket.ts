import {
  ApproveExpertMarketReleaseRequestSchema,
  ExpertMarketReleaseSchema,
  GetExpertMarketReleaseRequestSchema,
  ListExpertMarketReleasesRequestSchema,
  ListExpertMarketReleasesResponseSchema,
  RejectExpertMarketReleaseRequestSchema,
} from "@proto/admin/v1/admin_pb";

import { callConnect } from "@/lib/connect/transport";
import type {
  ExpertMarketRelease,
  ExpertMarketReleaseList,
  ExpertMarketReleaseListParams,
} from "./adminExpertMarketTypes";
import { fromProtoExpertMarketRelease } from "./expertMarketConvert";

const SERVICE = "proto.admin.v1.AdminService";

export async function listExpertMarketReleases(
  params: ExpertMarketReleaseListParams,
): Promise<ExpertMarketReleaseList> {
  const response = await callConnect(
    SERVICE,
    "ListExpertMarketReleases",
    ListExpertMarketReleasesRequestSchema,
    ListExpertMarketReleasesResponseSchema,
    params,
  );
  return {
    items: response.items.map(fromProtoExpertMarketRelease),
    total: Number(response.total),
    limit: response.limit,
    offset: response.offset,
  };
}

export async function getExpertMarketRelease(
  releaseId: number,
): Promise<ExpertMarketRelease> {
  const response = await callConnect(
    SERVICE,
    "GetExpertMarketRelease",
    GetExpertMarketReleaseRequestSchema,
    ExpertMarketReleaseSchema,
    { releaseId: BigInt(releaseId) },
  );
  return fromProtoExpertMarketRelease(response);
}

export async function approveExpertMarketRelease(
  releaseId: number,
): Promise<ExpertMarketRelease> {
  const response = await callConnect(
    SERVICE,
    "ApproveExpertMarketRelease",
    ApproveExpertMarketReleaseRequestSchema,
    ExpertMarketReleaseSchema,
    { releaseId: BigInt(releaseId) },
  );
  return fromProtoExpertMarketRelease(response);
}

export async function rejectExpertMarketRelease(
  releaseId: number,
  reason: string,
): Promise<ExpertMarketRelease> {
  const response = await callConnect(
    SERVICE,
    "RejectExpertMarketRelease",
    RejectExpertMarketReleaseRequestSchema,
    ExpertMarketReleaseSchema,
    { releaseId: BigInt(releaseId), reason },
  );
  return fromProtoExpertMarketRelease(response);
}
