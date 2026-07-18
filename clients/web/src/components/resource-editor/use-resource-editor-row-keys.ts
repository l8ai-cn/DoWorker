import { useId, useState } from "react";

let nextAddedRowKey = 0;

export function useResourceEditorRowKeys(length: number) {
  const prefix = useId();
  const [keys, setKeys] = useState(() => Array.from(
    { length },
    (_, index) => `${prefix}-initial-${index}`,
  ));

  return {
    keys: normalizeRowKeys(keys, length, prefix),
    appendKey: () => setKeys((current) => [
      ...normalizeRowKeys(current, length, prefix),
      `${prefix}-added-${nextAddedRowKey++}`,
    ]),
    removeKey: (index: number) => {
      setKeys((current) => normalizeRowKeys(current, length, prefix)
        .filter((_, item) => item !== index));
    },
  };
}

function normalizeRowKeys(keys: string[], length: number, prefix: string) {
  return Array.from(
    { length },
    (_, index) => keys[index] ?? `${prefix}-external-${index}`,
  );
}
