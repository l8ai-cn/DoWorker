import {
  RendererRegistryStore,
  type RendererReplacement,
} from "./rendererRegistryStore";
import {
  contentRendererKeyId,
  type ContentRendererKey,
} from "./rendererKeys";

export interface ContentRendererRegistration<
  TInline = unknown,
  TViewer = unknown,
  TEditor = unknown,
> {
  inline?: TInline;
  viewer: TViewer;
  editor?: TEditor;
}

export class ContentRendererRegistry<
  TRenderer = ContentRendererRegistration,
> {
  private readonly store = new RendererRegistryStore<
    ContentRendererKey,
    TRenderer
  >(contentRendererKeyId);

  register(
    key: ContentRendererKey,
    renderer: TRenderer,
    sourceId: string,
  ): void {
    this.store.register(key, renderer, sourceId);
  }

  lookup(key: ContentRendererKey): TRenderer | undefined {
    return this.store.lookup(key);
  }

  replace(
    key: ContentRendererKey,
    renderer: TRenderer,
    replacement: RendererReplacement,
  ): void {
    this.store.replace(key, renderer, replacement);
  }
}
