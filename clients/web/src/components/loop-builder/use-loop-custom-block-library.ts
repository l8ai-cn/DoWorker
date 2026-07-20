"use client";

import { useCallback, useEffect, useState } from "react";
import {
  createLoopCustomBlock,
  loadLoopCustomBlocks,
} from "./loop-custom-block-library";
import type { LoopCustomBlockDefinition } from "./loop-custom-block-types";

export function useLoopCustomBlockLibrary() {
  const [definitions, setDefinitions] = useState<LoopCustomBlockDefinition[]>([]);
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);

  const reload = useCallback(async () => {
    setError(undefined);
    try {
      const loaded = await loadLoopCustomBlocks();
      setDefinitions(loaded.definitions);
      return loaded.definitions;
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : "Unable to load custom blocks";
      setError(message);
      throw cause;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload().catch(() => undefined);
  }, [reload]);

  const create = useCallback(async (definition: LoopCustomBlockDefinition) => {
    setError(undefined);
    try {
      const loaded = await createLoopCustomBlock(definition);
      setDefinitions(loaded.definitions);
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : "Unable to create custom block";
      setError(message);
      throw cause;
    }
  }, []);

  return { create, definitions, error, loading, reload };
}
