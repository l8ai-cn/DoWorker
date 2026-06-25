const PALETTE = [
  { bg: "bg-info-bg", text: "text-info" },
  { bg: "bg-success-bg", text: "text-success" },
  { bg: "bg-accent", text: "text-primary" },
  { bg: "bg-warning-bg", text: "text-warning" },
  { bg: "bg-danger-bg", text: "text-danger" },
  { bg: "bg-info-bg", text: "text-info" },
  { bg: "bg-accent", text: "text-accent-foreground" },
  { bg: "bg-accent", text: "text-primary" },
] as const;

export type SkillAccent = (typeof PALETTE)[number];

export function skillAccent(seed: string): SkillAccent {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0;
  }
  return PALETTE[hash % PALETTE.length];
}
