export function workerBigInt(value: number, field: string): bigint {
  if (!Number.isSafeInteger(value)) {
    throw new Error(`unsafe ${field}`);
  }
  return BigInt(value);
}

export function workerNumber(value: bigint, field: string): number {
  const converted = Number(value);
  if (!Number.isSafeInteger(converted) || BigInt(converted) !== value) {
    throw new Error(`unsafe ${field}`);
  }
  return converted;
}
