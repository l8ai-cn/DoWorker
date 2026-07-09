/** Zustand persist storage that migrates a legacy localStorage key once. */
export function legacyPersistStorage(legacyKey: string) {
  return {
    getItem: (name: string): string | null => {
      const value = localStorage.getItem(name);
      if (value !== null) return value;
      const legacy = localStorage.getItem(legacyKey);
      if (legacy !== null) {
        localStorage.setItem(name, legacy);
        localStorage.removeItem(legacyKey);
        return legacy;
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
  };
}
