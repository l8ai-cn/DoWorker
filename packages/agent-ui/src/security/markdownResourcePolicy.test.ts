import { markdownImageSource } from "./markdownResourcePolicy";

describe("markdownImageSource", () => {
  it("blocks remote and relative image sources", () => {
    expect(markdownImageSource("https://tracker.test/p.png")).toBeUndefined();
    expect(markdownImageSource("http://tracker.test/p.png")).toBeUndefined();
    expect(markdownImageSource("/images/p.png")).toBeUndefined();
  });

  it("allows blob image sources", () => {
    expect(markdownImageSource("blob:https://app.test/id")).toBe(
      "blob:https://app.test/id",
    );
    expect(markdownImageSource("BLOB:https://app.test/id")).toBe(
      "BLOB:https://app.test/id",
    );
  });

  it("allows data image sources but blocks other data URLs", () => {
    expect(markdownImageSource("data:image/png;base64,AA==")).toBe(
      "data:image/png;base64,AA==",
    );
    expect(markdownImageSource("data:text/html,<h1>unsafe</h1>")).toBeUndefined();
    expect(markdownImageSource(" data:image/png;base64,AA==")).toBeUndefined();
    expect(markdownImageSource("data:image")).toBeUndefined();
  });
});
