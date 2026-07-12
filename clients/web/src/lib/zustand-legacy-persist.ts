import { createJSONStorage } from "zustand/middleware";

/** Zustand persist storage that migrates a legacy localStorage key once. */
export function legacyPersistStorage(legacyKey: string) {
  return createJSONStorage(() => ({
    getItem: (name: string): string | null => {
      try {
        const value = readStoredJson(name);
        if (value !== null) return value;
        const legacy = readStoredJson(legacyKey);
        if (legacy !== null) {
          localStorage.setItem(name, legacy);
          localStorage.removeItem(legacyKey);
          return legacy;
        }
      } catch (error) {
        console.warn(`Unable to read persisted UI state "${name}":`, error);
      }
      return null;
    },
    setItem: (name: string, value: string): void => {
      localStorage.setItem(name, value);
    },
    removeItem: (name: string): void => {
      localStorage.removeItem(name);
      localStorage.removeItem(legacyKey);
    },
  }));
}

function readStoredJson(name: string): string | null {
  const value = localStorage.getItem(name);
  if (value === null) return null;
  try {
    JSON.parse(value);
    return value;
  } catch {
    console.warn(`Discarding malformed persisted UI state "${name}".`);
    localStorage.removeItem(name);
    return null;
  }
}
