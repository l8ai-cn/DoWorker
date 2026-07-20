import { test as setup } from "@playwright/test";
import { clearAuthRateLimit } from "../helpers/redis";
import { authenticateE2ETestUser } from "../helpers/test-user-auth";

setup("authenticate as test user", async ({ browser }) => {
  clearAuthRateLimit();
  await authenticateE2ETestUser(browser);
});
