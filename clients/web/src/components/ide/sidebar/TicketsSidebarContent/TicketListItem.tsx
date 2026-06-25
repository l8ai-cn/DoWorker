"use client";

import type { Ticket } from "@/stores/ticket";
import { statusIcons } from "./types";

interface TicketListItemProps {
  ticket: Ticket;
  onClick: (slug: string) => void;
}

export function TicketListItem({ ticket, onClick }: TicketListItemProps) {
  return (
    <div
      className="group nav-row pressable items-start gap-2 rounded-lg motion-interactive hover:bg-surface-muted cursor-pointer"
      onClick={() => onClick(ticket.slug)}
    >
      {/* Status icon */}
      <div className="mt-0.5">
        {statusIcons[ticket.status]}
      </div>

      {/* Ticket info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-muted-foreground font-mono">
            {ticket.slug}
          </span>
          {ticket.priority === "urgent" && (
            <span className="text-danger text-xs">!</span>
          )}
          {ticket.priority === "high" && (
            <span className="text-warning text-xs">!!</span>
          )}
        </div>
        <p className="text-sm truncate">{ticket.title}</p>
      </div>
    </div>
  );
}

export default TicketListItem;
