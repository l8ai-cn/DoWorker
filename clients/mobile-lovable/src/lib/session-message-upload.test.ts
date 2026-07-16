import { beforeEach, describe, expect, it, vi } from "vitest";
import { uploadSessionFile } from "./files-api";
import { postMessageContent } from "./sessions-api";
import { postSessionMessageWithFiles } from "./session-message-upload";

vi.mock("./files-api", () => ({ uploadSessionFile: vi.fn() }));
vi.mock("./sessions-api", () => ({ postMessageContent: vi.fn() }));

describe("postSessionMessageWithFiles", () => {
  beforeEach(() => {
    vi.mocked(uploadSessionFile).mockReset();
    vi.mocked(postMessageContent).mockReset();
  });

  it("uploads files before sending their server file IDs", async () => {
    const image = new File(["image"], "design.png", { type: "image/png" });
    const document = new File(["notes"], "notes.pdf", { type: "application/pdf" });
    vi.mocked(uploadSessionFile)
      .mockResolvedValueOnce({ id: "file_image", filename: "design.png", bytes: 5 })
      .mockResolvedValueOnce({ id: "file_pdf", filename: "notes.pdf", bytes: 5 });

    await postSessionMessageWithFiles("session-1", "Review these files.", [
      { file: image },
      { file: document },
    ]);

    expect(postMessageContent).toHaveBeenCalledWith("session-1", [
      { type: "input_image", file_id: "file_image", filename: "design.png" },
      { type: "input_file", file_id: "file_pdf", filename: "notes.pdf" },
      { type: "input_text", text: "Review these files." },
    ]);
  });

  it("reuses uploaded file IDs after message delivery fails", async () => {
    const image = new File(["image"], "design.png", { type: "image/png" });
    const attachment = { file: image };
    vi.mocked(uploadSessionFile).mockResolvedValue({
      id: "file_image",
      filename: "design.png",
      bytes: 5,
    });
    vi.mocked(postMessageContent)
      .mockRejectedValueOnce(new Error("runner unavailable"))
      .mockResolvedValueOnce(undefined);

    await expect(
      postSessionMessageWithFiles("session-1", "Review this file.", [attachment]),
    ).rejects.toThrow("runner unavailable");
    await postSessionMessageWithFiles("session-1", "Review this file.", [attachment]);

    expect(uploadSessionFile).toHaveBeenCalledTimes(1);
    expect(postMessageContent).toHaveBeenCalledTimes(2);
  });
});
