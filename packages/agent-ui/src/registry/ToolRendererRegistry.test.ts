import { ToolRendererRegistry } from "./ToolRendererRegistry";

const key = {
  namespace: "openai.codex",
  semanticKey: "image.edit",
  schemaVersion: "1",
};

describe("ToolRendererRegistry", () => {
  it("uses an exact namespace, semantic key, and schema version", () => {
    const registry = new ToolRendererRegistry<string>();
    registry.register(key, "image editor", "builtin");

    expect(registry.lookup(key)).toBe("image editor");
    expect(registry.lookup({ ...key, schemaVersion: "2" })).toBeUndefined();
    expect(
      registry.lookup({ ...key, semanticKey: "image.generate" }),
    ).toBeUndefined();
  });

  it("rejects empty required key fields at runtime", () => {
    const registry = new ToolRendererRegistry<string>();

    expect(() =>
      registry.lookup({ ...key, namespace: "" }),
    ).toThrowError(/renderer_key_invalid.*namespace/);
  });

  it("rejects duplicate registration instead of using last-write-wins", () => {
    const registry = new ToolRendererRegistry<string>();
    registry.register(key, "builtin editor", "builtin");

    expect(() =>
      registry.register(key, "host editor", "host"),
    ).toThrowError(/renderer_key_conflict.*builtin.*host/);
    expect(registry.lookup(key)).toBe("builtin editor");
  });

  it("requires the expected current source for an explicit replacement", () => {
    const registry = new ToolRendererRegistry<string>();
    registry.register(key, "builtin editor", "builtin");

    expect(() =>
      registry.replace(key, "host editor", {
        expectedSourceId: "plugin",
        sourceId: "host",
      }),
    ).toThrowError(/renderer_source_mismatch.*builtin.*plugin/);

    registry.replace(key, "host editor", {
      expectedSourceId: "builtin",
      sourceId: "host",
    });
    expect(registry.lookup(key)).toBe("host editor");
  });
});
