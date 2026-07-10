import { test, expect } from "../../fixtures/index";
import { clearAuthRateLimit } from "../../helpers/redis";
import { makeConnectClient, ConnectError } from "../../helpers/connect-client";

/**
 * EnvBundle Connect-RPC regression for runtime environment variable bundles.
 * Tests verify the typed surface and 401/404 boundary cases.
 */
test.describe("EnvBundle API", () => {
  test.beforeEach(async () => {
    clearAuthRateLimit();
  });

  test("list env bundles", async ({ api }) => {
    const cc = await api.connect();
    const { items } = await cc.envBundle.listEnvBundles({}) as { items: unknown[] };
    expect(Array.isArray(items)).toBe(true);
  });

  test("list env bundles without auth returns 401", async ({ api: _api }) => {
    const cc = makeConnectClient("bad");
    await expect(cc.envBundle.listEnvBundles({})).rejects.toMatchObject({
      status: 401,
    });
  });

  test("create env bundle (runtime kind echoes values back)", async ({ api, db }) => {
    const name = `E2E Runtime ${Date.now()}`;
    db.cleanup(`DELETE FROM env_bundles WHERE name LIKE 'E2E Runtime%'`);

    const cc = await api.connect();
    const created = await cc.envBundle.createEnvBundle({
      name,
      kind: "runtime",
      data: { LOG_LEVEL: "debug" },
    }) as { kind: string; configuredValues: Record<string, string> };

    expect(created.kind).toBe("runtime");
    // Non-encrypted kinds round-trip plaintext values
    expect(created.configuredValues?.LOG_LEVEL).toBe("debug");

    db.cleanup(`DELETE FROM env_bundles WHERE name LIKE 'E2E Runtime%'`);
  });

  test("delete non-existent env bundle returns 404", async ({ api }) => {
    const cc = await api.connect();
    let caught: ConnectError | undefined;
    try {
      await cc.envBundle.deleteEnvBundle({ id: BigInt(999999) });
    } catch (err) {
      caught = err as ConnectError;
    }
    expect(caught?.status).toBe(404);
  });
});
