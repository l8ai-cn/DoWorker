export type NewChatLandingDraft = {
  message: string;
  files: File[];
  pickedAgentId: string | null;
  selectedHostId: string | null;
  sandboxSelected: boolean;
  sandboxRepoUrl: string;
  sandboxRepoBranch: string;
  workspace: string;
  branchName: string;
  baseBranch: string;
};

let draft: NewChatLandingDraft | null = null;

export function getNewChatLandingDraft(): NewChatLandingDraft | null {
  return draft;
}

export function preserveNewChatLandingDraft(next: NewChatLandingDraft | null): void {
  draft = next;
}

export function resetLandingDraft(): void {
  draft = null;
}
