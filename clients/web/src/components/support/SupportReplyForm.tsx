"use client";

import { useRef } from "react";
import { Send, Paperclip, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useTranslations } from "next-intl";

interface SupportReplyFormProps {
  content: string;
  files: File[];
  isSending: boolean;
  sendError: string | null;
  onContentChange: (value: string) => void;
  onSend: () => void;
  onFileSelect: (files: File[]) => void;
  onRemoveFile: (index: number) => void;
}

export function SupportReplyForm({
  content,
  files,
  isSending,
  sendError,
  onContentChange,
  onSend,
  onFileSelect,
  onRemoveFile,
}: SupportReplyFormProps) {
  const t = useTranslations();
  const fileInputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="panel-lift bg-surface-muted/40 p-4 pt-5">
      <div className="space-y-3">
        <Textarea
          value={content}
          onChange={(e) => onContentChange(e.target.value)}
          placeholder={t("support.replyPlaceholder")}
          className="min-h-[80px] resize-y"
          onKeyDown={(e) => {
            if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) onSend();
          }}
        />

        {files.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {files.map((file, i) => (
              <div key={i} className="flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs">
                <Paperclip className="h-3 w-3" />
                <span className="max-w-[150px] truncate">{file.name}</span>
                <button
                  type="button"
                  onClick={() => onRemoveFile(i)}
                  className="ml-1 text-muted-foreground hover:text-foreground motion-interactive"
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            ))}
          </div>
        )}

        {sendError && <p className="text-xs text-destructive">{sendError}</p>}

        <div className="flex items-center justify-between">
          <div>
            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept="image/*,.pdf,.txt,.log"
              onChange={(e) => {
                if (e.target.files) onFileSelect(Array.from(e.target.files));
              }}
              className="hidden"
            />
            <Button variant="ghost" size="sm" type="button" onClick={() => fileInputRef.current?.click()}>
              <Paperclip className="mr-1 h-4 w-4" />
              {t("support.attach")}
            </Button>
          </div>
          <Button type="button" onClick={onSend} disabled={!content.trim() || isSending} size="sm">
            <Send className="mr-1 h-4 w-4" />
            {isSending ? t("support.sending") : t("support.send")}
          </Button>
        </div>
      </div>
    </div>
  );
}
