import type { RefObject } from "react";
import type { Comment } from "@/hooks/useComments";
import type { useFileContent } from "@/hooks/useFileContent";

export interface ActiveSelection {
  start_index: number;
  end_index: number;
  anchor_content: string;
}

export type SaveStatus = "idle" | "unsaved" | "saving" | "saved" | "error" | "offline";

export const MONACO_SPLIT_BREAKPOINT = 900;
export const SPLIT_DIFF_MIN_WIDTH = 920;

export interface CodeViewerProps {
  conversationId: string;
  path: string;
  fileQuery: ReturnType<typeof useFileContent>;
  comments: Comment[];
  activeSelection: ActiveSelection | null;
  onSetActiveSelection: (selection: ActiveSelection | null) => void;
  panelOpen: boolean;
  searchOpen: boolean;
  setSearchOpen: (open: boolean) => void;
  searchInputRef: RefObject<HTMLInputElement | null>;
  viewMode: "editor" | "preview" | "source" | "diff";
  onDirtyChange?: (isDirty: boolean) => void;
  onSaveStatusChange?: (status: SaveStatus) => void;
  pendingBodyRef?: RefObject<string>;
}
