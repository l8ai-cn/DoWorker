import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

interface LandingMessages {
  landing: {
    nav: Record<string, unknown>;
    pricing?: unknown;
    finalCta: Record<string, unknown>;
    footer: {
      product: Record<string, unknown>;
    };
  };
}

interface DocsMessages {
  docs: {
    faq: {
      categories: Record<string, unknown>;
      items: Record<string, unknown>;
    };
  };
}

function collectCopy(value: unknown): string[] {
  if (typeof value === "string") {
    return [value];
  }
  if (Array.isArray(value)) {
    return value.flatMap(collectCopy);
  }
  if (value && typeof value === "object") {
    return Object.values(value).flatMap(collectCopy);
  }
  return [];
}

const retiredPricingCopy =
  /sign up free|start free|free tier|no credit card|kostenlos|gratis|gratuit|無料|무료|免费|信用卡/i;

describe("marketing pricing messages", () => {
  it("does not ship retired pricing copy in any locale", () => {
    const messagesRoot = resolve(__dirname, "../../../messages");
    const locales = readdirSync(messagesRoot);

    for (const locale of locales) {
      const file = resolve(messagesRoot, locale, "landing.json");
      const messages = JSON.parse(readFileSync(file, "utf8")) as LandingMessages;

      expect(messages.landing.nav).not.toHaveProperty("pricing");
      expect(messages.landing).not.toHaveProperty("pricing");
      expect(messages.landing.finalCta).not.toHaveProperty("getStartedFree");
      expect(messages.landing.finalCta).not.toHaveProperty("freeTier");
      expect(messages.landing.finalCta).not.toHaveProperty("noCreditCard");
      expect(messages.landing.footer.product).not.toHaveProperty("pricing");
      expect(collectCopy(messages.landing).join("\n")).not.toMatch(retiredPricingCopy);

      const workforceFile = resolve(messagesRoot, locale, "workforce.json");
      const workforceMessages = JSON.parse(readFileSync(workforceFile, "utf8")) as unknown;
      expect(collectCopy(workforceMessages).join("\n")).not.toMatch(retiredPricingCopy);

      const docsFile = resolve(messagesRoot, locale, "docs.json");
      const docsMessages = JSON.parse(readFileSync(docsFile, "utf8")) as DocsMessages;
      expect(docsMessages.docs.faq.categories).not.toHaveProperty("billing");
      expect(docsMessages.docs.faq.items).not.toHaveProperty("billingBYOK");
      expect(docsMessages.docs.faq.items).not.toHaveProperty("billingFree");
    }
  });
});
