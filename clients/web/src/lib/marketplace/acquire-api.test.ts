import { beforeEach, describe, expect, it, vi } from "vitest";

import { sessionStorageKey } from "@/lib/light-session";
import {
  applyInstallationPlan,
  createInstallationPlan,
  MarketplaceAcquireError,
} from "./acquire-api";

describe("marketplace acquisition API", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    window.localStorage.setItem(
      sessionStorageKey(window.location.origin),
      JSON.stringify({
        access_token: "market-token",
        expires_at: Math.floor(Date.now() / 1000) + 3600,
      }),
    );
  });

  it("reuses the operation ID as the stable apply idempotency key", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ status: "succeeded" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await applyInstallationPlan({
      installation_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      operation_id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      plan: {
        plan_id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
        plan_digest: "d".repeat(64),
        expires_at: "2026-07-12T12:00:00Z",
        listing_version_id: "31",
        estimated_credits_micro: "20000000",
        required_permissions: [],
      },
    });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/installation-operations/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb/apply",
      ),
      expect.objectContaining({
        headers: expect.objectContaining({
          "Idempotency-Key": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
        }),
      }),
    );
  });

  it("creates a plan with the selected organization and bearer token", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ operation_id: "operation-1", plan: {} }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await createInstallationPlan(
      "agent-cloud-market",
      "delivery",
      "31",
      9,
      301,
      { "seedance-video": 302 },
    );

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/markets/agent-cloud-market/listings/delivery/plans"),
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer market-token",
        }),
        body: JSON.stringify({
          listing_version_id: "31",
          target_platform_organization_id: "9",
          requested_configuration: {
            model_resource_id: 301,
            tool_model_resource_ids: { "seedance-video": 302 },
          },
        }),
      }),
    );
  });

  it("surfaces stable marketplace error codes", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({
        error: { code: "QUOTA_INSUFFICIENT", message: "市场额度不足" },
      }), {
        status: 409,
        headers: { "Content-Type": "application/json" },
      }),
    );

    await expect(
      createInstallationPlan("agent-cloud-market", "delivery", "31", 9, 301, {}),
    ).rejects.toEqual(
      expect.objectContaining<Partial<MarketplaceAcquireError>>({
        code: "QUOTA_INSUFFICIENT",
        message: "市场额度不足",
      }),
    );
  });
});
