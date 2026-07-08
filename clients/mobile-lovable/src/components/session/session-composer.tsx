import { Paperclip, Send, Square, X } from "lucide-react";
import { useRef, useState } from "react";
import { SlashMenu, detectSlashToken, type SlashCommand } from "@/components/slash-menu";
import { useSessionActions } from "@/lib/session-action-context";

export function SessionComposer() {
  const actions = useSessionActions();
  const [value, setValue] = useState("");
  const [token, setToken] = useState<{ start: number; token: string } | null>(null);
  const [attachments, setAttachments] = useState<
    { id: string; name: string; size: number; kind: "image" | "file"; url?: string }[]
  >([]);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const onChange = (v: string, caret: number) => {
    setValue(v);
    setToken(detectSlashToken(v, caret));
  };

  const applySlash = (cmd: SlashCommand) => {
    if (!token) return;
    const before = value.slice(0, token.start);
    const after = value.slice(token.start + token.token.length);
    const next = before + cmd.cmd + " " + after;
    setValue(next);
    setToken(null);
    requestAnimationFrame(() => {
      const el = inputRef.current;
      if (!el) return;
      el.focus();
      const pos = (before + cmd.cmd + " ").length;
      el.setSelectionRange(pos, pos);
    });
  };

  const onPickFiles = (files: FileList | null) => {
    if (!files) return;
    const next = Array.from(files).map((f) => ({
      id: `${f.name}-${f.size}-${Math.random().toString(36).slice(2, 7)}`,
      name: f.name,
      size: f.size,
      kind: f.type.startsWith("image/") ? ("image" as const) : ("file" as const),
      url: f.type.startsWith("image/") ? URL.createObjectURL(f) : undefined,
    }));
    setAttachments((prev) => [...prev, ...next]);
  };

  const removeAttachment = (id: string) => {
    setAttachments((prev) => {
      const target = prev.find((a) => a.id === id);
      if (target?.url) URL.revokeObjectURL(target.url);
      return prev.filter((a) => a.id !== id);
    });
  };

  const send = () => {
    const text = value.trim();
    if (!text || !actions.onSend) return;
    void actions.onSend(text).then(() => setValue(""));
  };

  return (
    <div className="border-t border-border/60 bg-background/95 px-3 pt-2.5 pb-2.5 backdrop-blur-xl">
      {token && (
        <div className="mb-2">
          <SlashMenu token={token.token} onPick={applySlash} />
        </div>
      )}

      {attachments.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {attachments.map((a) => (
            <div
              key={a.id}
              className="group relative flex items-center gap-1.5 rounded-lg bg-surface-2 py-1 pl-1 pr-6 ring-1 ring-border/50"
            >
              {a.kind === "image" && a.url ? (
                <img src={a.url} alt={a.name} className="h-6 w-6 rounded object-cover" />
              ) : (
                <span className="flex h-6 w-6 items-center justify-center rounded bg-surface text-muted-foreground">
                  <Paperclip className="h-3 w-3" />
                </span>
              )}
              <span className="max-w-[120px] truncate text-[11px]">{a.name}</span>
              <button
                onClick={() => removeAttachment(a.id)}
                className="absolute right-1 top-1/2 -translate-y-1/2 rounded-full p-0.5 text-muted-foreground hover:bg-surface hover:text-foreground"
                aria-label="移除附件"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
        </div>
      )}

      <input
        ref={fileRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e) => {
          onPickFiles(e.target.files);
          e.target.value = "";
        }}
      />

      <div className="flex items-end gap-2 rounded-2xl bg-surface px-2 py-2 ring-1 ring-border/50">
        <button
          onClick={() => fileRef.current?.click()}
          className="mb-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-muted-foreground hover:bg-surface-2 hover:text-primary"
          aria-label="上传文件"
        >
          <Paperclip className="h-4 w-4" />
        </button>
        <textarea
          ref={inputRef}
          rows={1}
          value={value}
          onChange={(e) => onChange(e.target.value, e.target.selectionStart ?? e.target.value.length)}
          onKeyUp={(e) => {
            const el = e.currentTarget;
            setToken(detectSlashToken(el.value, el.selectionStart ?? el.value.length));
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              send();
            }
          }}
          onBlur={() => setTimeout(() => setToken(null), 150)}
          placeholder="回复 agent，/ 唤起命令…"
          className="max-h-28 min-h-[36px] flex-1 resize-none bg-transparent py-1.5 text-[13px] leading-snug outline-none placeholder:text-muted-foreground"
        />
        {actions.onStop && (
          <button
            onClick={() => void actions.onStop?.()}
            className="mb-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            aria-label="停止 agent"
          >
            <Square className="h-3.5 w-3.5 fill-current" />
          </button>
        )}
        <button
          onClick={send}
          disabled={!value.trim()}
          className="mb-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground disabled:opacity-40"
          aria-label="发送"
        >
          <Send className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  );
}
