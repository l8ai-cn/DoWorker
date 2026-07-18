"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import type { PodData } from "@/lib/api/facade/pod";
import type { TicketContext } from "@/components/pod/CreatePodForm";
import { useCurrentOrg } from "@/stores/auth";

interface CreatePodModalProps {
  open: boolean;
  onClose: () => void;
  onCreated: (pod?: PodData) => void;
  ticketContext?: TicketContext;
  initialAgentSlug?: string;
  initialPrompt?: string;
}

export function CreatePodModal({
  open,
  onClose,
}: CreatePodModalProps) {
  const router = useRouter();
  const currentOrg = useCurrentOrg();

  useEffect(() => {
    if (!open || !currentOrg?.slug) return;
    onClose();
    router.push(`/${currentOrg.slug}/workers/new?mode=template`);
  }, [currentOrg?.slug, onClose, open, router]);

  return null;
}

export default CreatePodModal;
