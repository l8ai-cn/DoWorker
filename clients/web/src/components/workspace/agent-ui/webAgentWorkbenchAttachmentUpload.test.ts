import { describe, expect, it, vi } from "vitest";

import { uploadWebAgentWorkbenchAttachment } from "./webAgentWorkbenchAttachmentUpload";

describe("uploadWebAgentWorkbenchAttachment", () => {
  it.each([
    ["image", "brief.png", "image/png"],
    ["CSV", "sales.csv", "text/csv"],
    [
      "Word document",
      "brief.docx",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ],
  ])("uploads a %s through the session-file endpoint", async (
    _kind,
    name,
    mediaType,
  ) => {
    const file = new File(["data"], name, { type: mediaType });
    const fetcher = vi.fn(async () =>
      Response.json({
        id: "file_1",
        metadata: { bytes: file.size },
        name,
        object: "file",
        session_id: "session-1",
        type: "file",
      }),
    );

    const attachment = await uploadWebAgentWorkbenchAttachment(
      {
        access: { bearerToken: "token-1", orgSlug: "acme" },
        file,
        sessionId: "session-1",
      },
      fetcher,
    );

    const [url, init] = fetcher.mock.calls[0]!;
    expect(new URL(String(url)).pathname).toBe(
      "/v1/sessions/session-1/resources/files",
    );
    expect(init).toMatchObject({
      headers: {
        Authorization: "Bearer token-1",
        "X-Organization-Slug": "acme",
      },
      method: "POST",
    });
    const body = init?.body as FormData;
    expect(body.get("file")).toMatchObject({ name });
    expect(attachment).toEqual({
      bytes: 4,
      id: "file_1",
      mediaType,
      name,
    });
  });

  it("reports a rejected MIME type explicitly", async () => {
    await expect(
      uploadWebAgentWorkbenchAttachment(
        {
          access: { bearerToken: "token-1", orgSlug: "acme" },
          file: new File(["data"], "unsupported.bin", {
            type: "application/octet-stream",
          }),
          sessionId: "session-1",
        },
        vi.fn(async () =>
          Response.json(
            { error: "file type not allowed: application/octet-stream" },
            { status: 400 },
          ),
        ),
      ),
    ).rejects.toThrow("agent_attachment_upload_unsupported");
  });

  it("reports other upload failures without inventing an attachment", async () => {
    await expect(
      uploadWebAgentWorkbenchAttachment(
        {
          access: { bearerToken: "token-1", orgSlug: "acme" },
          file: new File(["data"], "brief.png", { type: "image/png" }),
          sessionId: "session-1",
        },
        vi.fn(async () => new Response(null, { status: 503 })),
      ),
    ).rejects.toThrow("agent_attachment_upload_failed");
  });
});
