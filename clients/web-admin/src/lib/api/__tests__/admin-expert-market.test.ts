import { beforeEach, describe, expect, it, vi } from "vitest";

const mockCallConnect = vi.fn();
vi.mock("@/lib/connect/transport", () => ({
  callConnect: (...args: unknown[]) => mockCallConnect(...args),
}));

import {
  approveExpertMarketRelease,
  getExpertMarketRelease,
  listExpertMarketReleases,
  rejectExpertMarketRelease,
} from "../admin";

const protoRelease = {
  id: 12n,
  applicationId: 22n,
  applicationSlug: "video-expert",
  sourceExpertId: 32n,
  publisherOrganizationId: 42n,
  publisherUserId: 52n,
  version: 3,
  status: "pending",
  name: "Video Expert",
  summary: "Build production videos",
  description: "A detailed expert description",
  category: "media",
  icon: "video",
  tags: ["video", "editing"],
  outcomes: ["Published video"],
  featured: false,
  expertSnapshotJson: "{\"model\":\"codex\"}",
  workerSpecSnapshotJson: "{\"runtime\":\"runner\"}",
  skillDependenciesJson: "[{\"skill_id\":7,\"slug\":\"remotion\",\"version\":1}]",
  reviewerUserId: 62n,
  submittedAt: "2026-07-14T08:00:00Z",
  createdAt: "2026-07-14T07:00:00Z",
};

describe("Admin API - Expert marketplace", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("lists releases and converts proto fields", async () => {
    mockCallConnect.mockResolvedValue({
      items: [protoRelease],
      total: 1n,
      limit: 20,
      offset: 0,
    });

    const result = await listExpertMarketReleases({
      status: "published",
      limit: 20,
      offset: 0,
    });

    expect(mockCallConnect.mock.calls[0][0]).toBe("proto.admin.v1.AdminService");
    expect(mockCallConnect.mock.calls[0][1]).toBe("ListExpertMarketReleases");
    expect(mockCallConnect.mock.calls[0][4]).toEqual({
      status: "published",
      limit: 20,
      offset: 0,
    });
    expect(result).toMatchObject({
      total: 1,
      limit: 20,
      offset: 0,
      items: [{
        id: 12,
        application_id: 22,
        application_slug: "video-expert",
        source_expert_id: 32,
        publisher_organization_id: 42,
        publisher_user_id: 52,
        reviewer_user_id: 62,
        expert_snapshot_json: "{\"model\":\"codex\"}",
        skill_dependencies_json: "[{\"skill_id\":7,\"slug\":\"remotion\",\"version\":1}]",
      }],
    });
  });

  it("gets, approves, and rejects releases with converted responses", async () => {
    mockCallConnect.mockResolvedValue(protoRelease);

    const detail = await getExpertMarketRelease(12);
    await approveExpertMarketRelease(12);
    await rejectExpertMarketRelease(12, "Missing license");

    expect(detail.id).toBe(12);
    expect(mockCallConnect.mock.calls[0][1]).toBe("GetExpertMarketRelease");
    expect(mockCallConnect.mock.calls[0][4]).toEqual({ releaseId: 12n });
    expect(mockCallConnect.mock.calls[1][1]).toBe("ApproveExpertMarketRelease");
    expect(mockCallConnect.mock.calls[1][4]).toEqual({ releaseId: 12n });
    expect(mockCallConnect.mock.calls[2][1]).toBe("RejectExpertMarketRelease");
    expect(mockCallConnect.mock.calls[2][4]).toEqual({
      releaseId: 12n,
      reason: "Missing license",
    });
  });
});
