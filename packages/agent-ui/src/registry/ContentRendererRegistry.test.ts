import { ContentRendererRegistry } from "./ContentRendererRegistry";

const key = {
  blockKind: "artifact",
  mediaType: "video/mp4",
  role: "playable",
  schemaVersion: "1",
};

describe("ContentRendererRegistry", () => {
  it("does not wildcard missing media type, role, or schema version", () => {
    const registry = new ContentRendererRegistry<string>();
    registry.register(key, "video player", "builtin");

    expect(registry.lookup(key)).toBe("video player");
    expect(registry.lookup({ ...key, mediaType: undefined })).toBeUndefined();
    expect(registry.lookup({ ...key, role: "poster" })).toBeUndefined();
    expect(registry.lookup({ ...key, schemaVersion: "2" })).toBeUndefined();
  });

  it("keeps manifest renderer identities exact", () => {
    const registry = new ContentRendererRegistry<string>();
    const manifestKey = {
      blockKind: "artifact",
      manifestType: "image_edit",
      schemaVersion: "1",
    };
    registry.register(manifestKey, "image editor", "builtin.image-edit");

    expect(registry.lookup(manifestKey)).toBe("image editor");
    expect(
      registry.lookup({ ...manifestKey, manifestType: "presentation" }),
    ).toBeUndefined();
    expect(
      registry.lookup({ ...manifestKey, schemaVersion: "2" }),
    ).toBeUndefined();
  });

  it("rejects invalid optional key fields from serialized host input", () => {
    const registry = new ContentRendererRegistry<string>();
    const genericKey = {
      blockKind: "progress",
      schemaVersion: "1",
    };
    registry.register(genericKey, "progress", "builtin");

    expect(registry.lookup(genericKey)).toBe("progress");
    expect(
      () => registry.lookup({ ...genericKey, mediaType: "" }),
    ).toThrowError(/renderer_key_invalid.*mediaType/);
    expect(() =>
      registry.lookup({
        ...genericKey,
        mediaType: null as unknown as string,
      }),
    ).toThrowError(/renderer_key_invalid.*mediaType/);
  });

  it("uses the same explicit conflict and replacement contract", () => {
    const registry = new ContentRendererRegistry<string>();
    registry.register(key, "builtin player", "builtin");

    expect(() =>
      registry.register(key, "host player", "host"),
    ).toThrowError(/renderer_key_conflict/);
    registry.replace(key, "host player", {
      expectedSourceId: "builtin",
      sourceId: "host",
    });
    expect(registry.lookup(key)).toBe("host player");
  });
});
