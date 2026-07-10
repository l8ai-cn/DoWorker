export function getNextMissionStep(current: number, total: number): number {
  return total > 0 ? (current + 1) % total : 0
}
