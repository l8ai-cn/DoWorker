import { createJSONStorage } from "zustand/middleware";

/** Zustand persist storage that migrates a legacy localStorage key once. */
export function legacyPersistStorage(legacyKey: string) {
  return createJSONStorage(() => ({
    getItem: (name: string): string | null => {
      try {
        const storage = browserStorage();
        if (!storage) return null;
        const value = readStoredJson(name);
        if (value !== null) return value;
        const legacy = readStoredJson(legacyKey);
        if (legacy !== null) {
          storage.setItem(name, legacy);
          storage.removeItem(legacyKey);
          return legacy;
        }
      } catch (error) {
        console.warn(`Unable to read persisted UI state "${name}":`, error);
      }
      return null;
    },
    setItem: (name: string, value: string): void => {
      browserStorage()?.setItem(name, value);
    },
    removeItem: (name: string): void => {
      const storage = browserStorage();
      storage?.removeItem(name);
      storage?.removeItem(legacyKey);
    },
  }));
}

function readStoredJson(name: string): string | null {
  const storage = browserStorage();
  if (!storage) return null;
  const value = storage.getItem(name);
  if (value === null) return null;
  try {
    JSON.parse(value);
    return value;
  } catch {
    console.warn(`Discarding malformed persisted UI state "${name}".`);
    storage.removeItem(name);
    return null;
  }
}

function browserStorage(): Storage | null {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}
