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

    await postSessionMessageWithFiles("session-1", "Review these files.", [image, document]);

    expect(postMessageContent).toHaveBeenCalledWith("session-1", [
      { type: "input_image", file_id: "file_image", filename: "design.png" },
      { type: "input_file", file_id: "file_pdf", filename: "notes.pdf" },
      { type: "input_text", text: "Review these files." },
    ]);
  });
});
