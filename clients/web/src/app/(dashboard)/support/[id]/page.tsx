"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  TicketStatusBadge,
  TicketCategoryBadge,
  TicketPriorityBadge,
} from "@/components/support/ticket-status-badge";
import { MessageList } from "@/components/support/message-list";
import { SupportReplyForm } from "@/components/support/SupportReplyForm";
import { addSupportTicketMessage } from "@/lib/api/facade/support-ticket";
import type { SupportTicketDetail } from "@/lib/api/facade/supportTicketConnect";
import { getSupportTicketDetail } from "@/lib/api/facade/supportTicketConnect";

export default function SupportTicketDetailPage() {
  const params = useParams();
  const router = useRouter();
  const t = useTranslations();
  const ticketId = Number(params.id);
  const maxFileSize = 10 * 1024 * 1024;

  const [data, setData] = useState<SupportTicketDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [replyContent, setReplyContent] = useState("");
  const [replyFiles, setReplyFiles] = useState<File[]>([]);
  const [isSending, setIsSending] = useState(false);
  const [sendError, setSendError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const fetchDetail = useCallback(async () => {
    try {
      const result = await getSupportTicketDetail(ticketId);
      setData(result);
      setError(null);
    } catch {
      setError(t("support.error.loadFailed"));
    } finally {
      setIsLoading(false);
    }
  }, [ticketId, t]);

  useEffect(() => {
    fetchDetail();
    const interval = setInterval(() => {
      if (!document.hidden) fetchDetail();
    }, 15000);
    return () => clearInterval(interval);
  }, [fetchDetail]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [data?.messages?.length]);

  const handleSend = async () => {
    if (!replyContent.trim()) return;
    setIsSending(true);
    setSendError(null);
    try {
      await addSupportTicketMessage(
        ticketId,
        replyContent.trim(),
        replyFiles.length > 0 ? replyFiles : undefined,
      );
      setReplyContent("");
      setReplyFiles([]);
      await fetchDetail();
    } catch {
      setSendError(t("support.error.sendFailed"));
    } finally {
      setIsSending(false);
    }
  };

  const handleFileSelect = (newFiles: File[]) => {
    const oversized = newFiles.filter((f) => f.size > maxFileSize);
    if (oversized.length > 0) {
      toast.error(t("support.fileTooLarge", { max: "10MB" }));
      return;
    }
    setReplyFiles((prev) => [...prev, ...newFiles]);
  };

  if (isLoading) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 space-y-4">
        <div className="h-8 w-48 animate-pulse rounded bg-muted" />
        <div className="h-24 animate-pulse rounded-lg bg-muted/30" />
        <div className="h-64 animate-pulse rounded-lg bg-muted/30" />
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6">
        <div className="flex flex-col items-center py-16">
          <p className="text-destructive">{error}</p>
          <Button variant="outline" onClick={fetchDetail} className="mt-4">
            {t("support.retry")}
          </Button>
        </div>
      </div>
    );
  }

  if (!data?.ticket) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6">
        <div className="flex flex-col items-center py-16">
          <p className="text-muted-foreground">{t("support.notFound")}</p>
          <Button variant="outline" onClick={() => router.push("/support")} className="mt-4">
            {t("support.backToList")}
          </Button>
        </div>
      </div>
    );
  }

  const { ticket, messages } = data;

  return (
    <div className="mx-auto flex h-full max-w-3xl flex-col px-4 py-6">
      <div className="mb-6 flex shrink-0 items-start gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push("/support")} className="mt-0.5">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="min-w-0 flex-1">
          <h1 className="text-lg font-bold">{ticket.title}</h1>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <TicketStatusBadge status={ticket.status} />
            <TicketCategoryBadge category={ticket.category} />
            <TicketPriorityBadge priority={ticket.priority} />
            <span className="text-xs text-muted-foreground">
              #{Number(ticket.id)} · {new Date(ticket.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>

      <div className="surface-card flex min-h-0 flex-1 flex-col overflow-hidden">
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          <MessageList messages={messages || []} />
          <div ref={messagesEndRef} />
        </div>

        {ticket.status !== "closed" && (
          <SupportReplyForm
            content={replyContent}
            files={replyFiles}
            isSending={isSending}
            sendError={sendError}
            onContentChange={setReplyContent}
            onSend={handleSend}
            onFileSelect={handleFileSelect}
            onRemoveFile={(index) => setReplyFiles((prev) => prev.filter((_, i) => i !== index))}
          />
        )}
      </div>
    </div>
  );
}
