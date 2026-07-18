import { describe, expect, it, vi } from "vitest";

import { uploadEmbeddedAttachment } from "./embed-attachment-api";

describe("uploadEmbeddedAttachment", () => {
  it("uses the embed session file protocol", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "file_12345678",
          metadata: { bytes: 7 },
          name: "notes.txt",
        }),
        { status: 200 },
      ),
    );
    const file = new File(["content"], "notes.txt", { type: "text/plain" });

    await expect(
      uploadEmbeddedAttachment(request, "/v1/embed/sessions/conv_embed", file),
    ).resolves.toEqual({
      id: "file_12345678",
      name: "notes.txt",
      mediaType: "text/plain",
      bytes: 7,
    });
    expect(request).toHaveBeenCalledWith(
      "/v1/embed/sessions/conv_embed/resources/files",
      expect.objectContaining({ method: "POST" }),
    );
    const init = request.mock.calls[0]?.[1] as RequestInit;
    expect(init.body).toBeInstanceOf(FormData);
    expect((init.body as FormData).get("file")).toBe(file);
  });

  it("does not accept a response without a session file id", async () => {
    const request = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify({ name: "notes.txt" }), { status: 200 }));

    await expect(
      uploadEmbeddedAttachment(
        request,
        "/v1/embed/sessions/conv_embed",
        new File(["content"], "notes.txt"),
      ),
    ).rejects.toThrow("agent_attachment_upload_invalid");
  });
});
