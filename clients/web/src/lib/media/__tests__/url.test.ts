import { describe, it, expect } from "vitest";
import {
  classifyMediaUrl,
  buildEmbedURL,
  isSafeImageSrc,
  isSafeRenderableSrc,
  extOf,
} from "../url";

describe("extOf", () => {
  it("extracts lowercase extension ignoring query and hash", () => {
    expect(extOf("https://x.com/a/photo.PNG?sig=1#frag")).toBe("png");
    expect(extOf("/relative/video.mp4")).toBe("mp4");
    expect(extOf("https://x.com/no-extension")).toBe("");
  });
});

describe("classifyMediaUrl", () => {
  it("classifies image / video / audio / html by extension", () => {
    expect(classifyMediaUrl("https://cdn.example.com/shot.png")).toBe("image");
    expect(classifyMediaUrl("https://cdn.example.com/demo.mp4")).toBe("video");
    expect(classifyMediaUrl("https://cdn.example.com/voice.mp3")).toBe("audio");
    expect(classifyMediaUrl("https://cdn.example.com/report.html")).toBe("html");
  });

  it("classifies whitelisted providers by host", () => {
    expect(classifyMediaUrl("https://www.youtube.com/watch?v=dQw4w9WgXcQ")).toBe("youtube");
    expect(classifyMediaUrl("https://youtu.be/dQw4w9WgXcQ")).toBe("youtube");
    expect(classifyMediaUrl("https://vimeo.com/12345678")).toBe("vimeo");
    expect(classifyMediaUrl("https://www.loom.com/share/abc123def")).toBe("loom");
    expect(classifyMediaUrl("https://www.figma.com/file/xyz/My-Design")).toBe("figma");
    expect(classifyMediaUrl("https://codesandbox.io/s/happy-tree-abc123")).toBe("codesandbox");
  });

  it("falls back to link for unknown or unsafe URLs", () => {
    expect(classifyMediaUrl("https://example.com/article")).toBe("link");
     
    expect(classifyMediaUrl("javascript:alert(1)")).toBe("link");
    expect(classifyMediaUrl("data:text/html,<script>alert(1)</script>")).toBe("link");
    expect(classifyMediaUrl("")).toBe("link");
  });

  it("classifies same-origin relative paths by extension", () => {
    expect(classifyMediaUrl("/uploads/pic.webp")).toBe("image");
    expect(classifyMediaUrl("uploads/clip.webm")).toBe("video");
    expect(classifyMediaUrl("//evil.com/pic.png")).toBe("link");
  });
});

describe("buildEmbedURL", () => {
  it("builds youtube embed URLs from watch and short links", () => {
    expect(buildEmbedURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube")).toBe(
      "https://www.youtube.com/embed/dQw4w9WgXcQ",
    );
    expect(buildEmbedURL("https://youtu.be/dQw4w9WgXcQ", "youtube")).toBe(
      "https://www.youtube.com/embed/dQw4w9WgXcQ",
    );
  });

  it("builds vimeo and loom embeds", () => {
    expect(buildEmbedURL("https://vimeo.com/12345678", "vimeo")).toBe(
      "https://player.vimeo.com/video/12345678",
    );
    expect(buildEmbedURL("https://www.loom.com/share/abc123def", "loom")).toBe(
      "https://www.loom.com/embed/abc123def",
    );
  });

  it("returns null when the URL doesn't match the provider shape", () => {
    expect(buildEmbedURL("https://vimeo.com/about", "vimeo")).toBeNull();
    expect(buildEmbedURL("https://example.com/x.mp4", "video")).toBeNull();
  });
});

describe("isSafeRenderableSrc", () => {
  it("accepts http(s) and relative paths, rejects other schemes", () => {
    expect(isSafeRenderableSrc("https://x.com/a.html")).toBe(true);
    expect(isSafeRenderableSrc("/a.html")).toBe(true);
     
    expect(isSafeRenderableSrc("javascript:alert(1)")).toBe(false);
    expect(isSafeRenderableSrc("data:text/html,x")).toBe(false);
    expect(isSafeRenderableSrc("//evil.com/a.html")).toBe(false);
  });
});

describe("isSafeImageSrc", () => {
  it("additionally allows data:image URIs", () => {
    expect(isSafeImageSrc("data:image/png;base64,iVBORw0KGgo=")).toBe(true);
    expect(isSafeImageSrc("data:image/svg+xml;base64,PHN2Zz4=")).toBe(true);
    expect(isSafeImageSrc("data:text/html;base64,PHNjcmlwdD4=")).toBe(false);
    expect(isSafeImageSrc("https://x.com/a.png")).toBe(true);
     
    expect(isSafeImageSrc("javascript:alert(1)")).toBe(false);
  });
});
