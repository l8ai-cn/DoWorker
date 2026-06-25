export type ChangelogChangeType = "added" | "changed" | "fixed" | "removed";

export interface ChangelogEntry {
  version: string;
  date: string;
  changes: {
    type: ChangelogChangeType;
    items: string[];
  }[];
}
