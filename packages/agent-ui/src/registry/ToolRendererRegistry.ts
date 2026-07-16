import {
  RendererRegistryStore,
  type RendererReplacement,
} from "./rendererRegistryStore";
import {
  toolRendererKeyId,
  type ToolRendererKey,
} from "./rendererKeys";

export interface ToolRendererRegistration<
  TSummary = unknown,
  TDetail = unknown,
  TWorkbench = unknown,
> {
  summary?: TSummary;
  detail?: TDetail;
  workbench?: TWorkbench;
}

export class ToolRendererRegistry<
  TRenderer = ToolRendererRegistration,
> {
  private readonly store = new RendererRegistryStore<
    ToolRendererKey,
    TRenderer
  >(toolRendererKeyId);

  register(
    key: ToolRendererKey,
    renderer: TRenderer,
    sourceId: string,
  ): void {
    this.store.register(key, renderer, sourceId);
  }

  lookup(key: ToolRendererKey): TRenderer | undefined {
    return this.store.lookup(key);
  }

  replace(
    key: ToolRendererKey,
    renderer: TRenderer,
    replacement: RendererReplacement,
  ): void {
    this.store.replace(key, renderer, replacement);
  }
}
