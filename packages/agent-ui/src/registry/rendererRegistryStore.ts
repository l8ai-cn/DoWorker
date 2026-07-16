export interface RendererReplacement {
  expectedSourceId: string;
  sourceId: string;
}

interface RendererEntry<TRenderer> {
  renderer: TRenderer;
  sourceId: string;
}

export class RendererRegistryStore<TKey, TRenderer> {
  private readonly entries = new Map<string, RendererEntry<TRenderer>>();
  private readonly keyId: (key: TKey) => string;

  constructor(keyId: (key: TKey) => string) {
    this.keyId = keyId;
  }

  register(key: TKey, renderer: TRenderer, sourceId: string): void {
    const id = this.keyId(key);
    const existing = this.entries.get(id);
    if (existing) {
      throw new Error(
        `renderer_key_conflict: key=${id} existing_source=${existing.sourceId} new_source=${sourceId}`,
      );
    }
    this.entries.set(id, { renderer, sourceId });
  }

  lookup(key: TKey): TRenderer | undefined {
    return this.entries.get(this.keyId(key))?.renderer;
  }

  replace(
    key: TKey,
    renderer: TRenderer,
    replacement: RendererReplacement,
  ): void {
    const id = this.keyId(key);
    const existing = this.entries.get(id);
    if (!existing) {
      throw new Error(`renderer_key_missing: key=${id}`);
    }
    if (existing.sourceId !== replacement.expectedSourceId) {
      throw new Error(
        `renderer_source_mismatch: key=${id} actual_source=${existing.sourceId} expected_source=${replacement.expectedSourceId}`,
      );
    }
    this.entries.set(id, {
      renderer,
      sourceId: replacement.sourceId,
    });
  }
}
