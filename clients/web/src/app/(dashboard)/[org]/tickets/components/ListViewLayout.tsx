"use client";

import { Ticket } from "@/stores/ticket";
import { VirtualizedTicketList } from "@/components/tickets/VirtualizedTicketList";
import { TicketsPageHeader } from "@/components/tickets";
import { EmptyState } from "@/components/ui/empty-state";
import { TicketListView } from "./TicketListView";

const VIRTUALIZATION_THRESHOLD = 50;

interface ListViewLayoutProps {
  tickets: Ticket[];
  selectedSlug: string | null;
  onTicketClick: (ticket: Ticket) => void;
  t: (key: string) => string;
}

export function ListViewLayout({
  tickets,
  selectedSlug,
  onTicketClick,
  t,
}: ListViewLayoutProps) {
  const useVirtualization = tickets.length > VIRTUALIZATION_THRESHOLD;

  return (
    <div className="flex h-full flex-col">
      <TicketsPageHeader />
      <div className="min-h-0 flex-1 overflow-hidden p-4">
        {tickets.length === 0 ? (
          <EmptyState
            size="full"
            title={t("tickets.emptyState.title")}
            description={t("tickets.emptyState.createFirst")}
          />
        ) : useVirtualization ? (
          <VirtualizedTicketList
            tickets={tickets}
            selectedSlug={selectedSlug}
            onTicketClick={onTicketClick}
            t={t}
          />
        ) : (
          <TicketListView
            tickets={tickets}
            selectedSlug={selectedSlug}
            onTicketClick={onTicketClick}
            t={t}
          />
        )}
      </div>
    </div>
  );
}
