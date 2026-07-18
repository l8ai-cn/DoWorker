import type { EmbedSessionClient } from "@/embed-session-api";
import {
  loadEmbeddedRuntimeHydration,
  type EmbeddedRuntimeHydration,
} from "./embeddedRuntimeHydration";

interface EmbeddedHydrationResult {
  hydration: EmbeddedRuntimeHydration;
  preserveStatus: boolean;
}

export class EmbeddedHydrationCoordinator {
  private sequence = 0;
  private statusRevision = 0;

  reset(): void {
    this.sequence = 0;
    this.statusRevision = 0;
  }

  markStatusChanged(): void {
    this.statusRevision += 1;
  }

  async hydrate(
    client: EmbedSessionClient,
    signal: AbortSignal,
  ): Promise<EmbeddedHydrationResult | null> {
    const sequence = ++this.sequence;
    const statusRevision = this.statusRevision;
    const hydration = await loadEmbeddedRuntimeHydration(client);
    if (signal.aborted || sequence !== this.sequence) return null;
    return {
      hydration,
      preserveStatus: this.statusRevision !== statusRevision,
    };
  }
}
