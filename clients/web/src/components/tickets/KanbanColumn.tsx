"use client";

import { useMemo } from "react";
import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy, useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { TicketCard } from "./TicketCard";
import { Ticket, TicketStatus } from "@/stores/ticket";
import { cn } from "@/lib/utils";
import { GripVertical, Loader2, ChevronLeft } from "lucide-react";

type Status = TicketStatus;

export const statusConfig: { status: Status; labelKey: string; topColor: string; dotColor: string }[] = [
  { status: "backlog", labelKey: "tickets.status.backlog", topColor: "bg-muted-foreground", dotColor: "bg-muted-foreground" },
  { status: "todo", labelKey: "tickets.status.todo", topColor: "bg-info", dotColor: "bg-info" },
  { status: "in_progress", labelKey: "tickets.status.in_progress", topColor: "bg-warning", dotColor: "bg-warning" },
  { status: "in_review", labelKey: "tickets.status.in_review", topColor: "bg-primary", dotColor: "bg-primary" },
  { status: "done", labelKey: "tickets.status.done", topColor: "bg-success", dotColor: "bg-success" },
];

interface SortableTicketProps {
  ticket: Ticket;
  onTicketClick?: (ticket: Ticket) => void;
  onMouseEnter: () => void;
  onMouseLeave: () => void;
}

export function SortableTicket({ ticket, onTicketClick, onMouseEnter, onMouseLeave }: SortableTicketProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: ticket.slug });
  const style = { transform: CSS.Transform.toString(transform), transition };

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}
      className={cn("transition-all duration-200 cursor-grab active:cursor-grabbing",
        isDragging ? "opacity-40 scale-[0.97] z-50" : "hover:scale-[1.01] hover:shadow-sm")}
      onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave}>
      <TicketCard ticket={ticket} onClick={() => onTicketClick?.(ticket)} showRepository={false} showStatus={false} />
    </div>
  );
}

interface DroppableColumnProps {
  status: Status;
  labelKey: string;
  topColor: string;
  dotColor: string;
  tickets: Ticket[];
  totalCount?: number;
  hasMore?: boolean;
  loadingMore?: boolean;
  sentinelRef?: React.RefObject<HTMLDivElement | null>;
  isOver: boolean;
  onTicketClick?: (ticket: Ticket) => void;
  onCollapse?: () => void;
  prefetchOnHover: (slug: string) => void;
  cancelPrefetch: () => void;
  t: (key: string) => string;
}

export function DroppableColumn({
  status, labelKey, topColor, dotColor, tickets, totalCount,
  loadingMore, sentinelRef, isOver, onTicketClick, onCollapse,
  prefetchOnHover, cancelPrefetch, t,
}: DroppableColumnProps) {
  const ticketIds = useMemo(() => tickets.map((t) => t.slug), [tickets]);
  const { setNodeRef, isOver: isDroppableOver } = useDroppable({ id: status });
  const highlighted = isOver || isDroppableOver;

  return (
    <div ref={setNodeRef}
      data-testid="kanban-column"
      data-column-status={status}
      className={cn("flex-shrink-0 w-80 flex flex-col rounded-xl bg-surface-muted/70 shadow-xs ring-1 ring-border/30 transition-all duration-200 overflow-hidden",
        highlighted && "ring-2 ring-primary/45 bg-primary/5 shadow-[var(--shadow-soft)]")}>
      <div className={cn("h-1 w-full", topColor)} />
      <div className="flex items-center justify-between px-3 py-3">
        <div className="flex items-center gap-2">
          <div className={cn("w-2 h-2 rounded-full", dotColor)} />
          <h3 className="font-medium text-sm">{t(labelKey)}</h3>
          <span className="text-xs text-muted-foreground font-mono">
            {totalCount !== undefined ? totalCount : tickets.length}
          </span>
        </div>
        {onCollapse && (
          <button onClick={onCollapse}
            className="p-0.5 rounded hover:bg-muted text-muted-foreground/50 hover:text-foreground transition-colors">
            <ChevronLeft className="w-3.5 h-3.5" />
          </button>
        )}
      </div>
      <div className="flex-1 overflow-y-auto px-2.5 pb-2.5 space-y-2 min-h-[100px]">
        <SortableContext items={ticketIds} strategy={verticalListSortingStrategy}>
          {tickets.map((ticket) => (
            <SortableTicket key={ticket.slug} ticket={ticket} onTicketClick={onTicketClick}
              onMouseEnter={() => prefetchOnHover(ticket.slug)} onMouseLeave={cancelPrefetch} />
          ))}
        </SortableContext>
        {sentinelRef && <div ref={sentinelRef} className="h-1 shrink-0" />}
        {loadingMore && (
          <div className="flex justify-center py-2">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          </div>
        )}
        {tickets.length === 0 && (
          <div className={cn("flex flex-col items-center justify-center py-10 text-muted-foreground/50 transition-colors rounded-lg bg-surface-raised/45 ring-1 ring-transparent",
            highlighted && "text-primary/60 ring-primary/25")}>
            <GripVertical className="h-5 w-5 mb-2" />
            <span className="text-xs font-medium">
              {highlighted ? (t("tickets.kanban.dropHere") || "Drop here") : t("tickets.kanban.noTickets")}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
