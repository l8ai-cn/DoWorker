import React from "react";
import { Circle, CheckCircle2, Clock, AlertCircle } from "lucide-react";
import type { TicketStatus, TicketPriority } from "@/stores/ticket";

export interface TicketsSidebarContentProps {
  className?: string;
}

export const statusIcons: Record<TicketStatus, React.ReactNode> = {
  backlog: <Circle className="w-3 h-3 text-muted-foreground" />,
  todo: <Circle className="w-3 h-3 text-info" />,
  in_progress: <Clock className="w-3 h-3 text-warning" />,
  in_review: <AlertCircle className="w-3 h-3 text-primary" />,
  done: <CheckCircle2 className="w-3 h-3 text-success" />,
};

export const statusOptions: TicketStatus[] = ["backlog", "todo", "in_progress", "in_review", "done"];
export const priorityOptions: TicketPriority[] = ["urgent", "high", "medium", "low", "none"];

export interface TicketFilterState {
  searchQuery: string;
  selectedStatuses: TicketStatus[];
  selectedPriorities: TicketPriority[];
  selectedRepositoryIds: number[];
}

export interface TicketFilterActions {
  setSearchQuery: (query: string) => void;
  toggleStatus: (status: TicketStatus) => void;
  togglePriority: (priority: TicketPriority) => void;
  toggleRepository: (id: number) => void;
  clearAllFilters: () => void;
  hasActiveFilters: boolean;
}
