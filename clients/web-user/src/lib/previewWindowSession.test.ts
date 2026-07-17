import { describe, expect, it } from "vitest";
import { readPreviewWindowSessionUrl } from "./previewWindowSession";

const SESSION_URL =
  "https://pod1.preview.example.test/preview/pod1/__session?token=JWT";

describe("readPreviewWindowSessionUrl", () => {
  it("reads a validated session URL from the fragment", () => {
    expect(
      readPreviewWindowSessionUrl(
        `#${encodeURIComponent(SESSION_URL)}`,
        "https://preview.example.test",
      ),
    ).toBe(SESSION_URL);
  });

  it("rejects an untrusted preview origin", () => {
    expect(() =>
      readPreviewWindowSessionUrl(
        `#${encodeURIComponent(
          "https://evil.example.test/preview/pod1/__session?token=JWT",
        )}`,
        "https://preview.example.test",
      ),
    ).toThrow("预览地址无效");
  });
});
