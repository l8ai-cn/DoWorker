import { Send, Square } from "lucide-react";

export function AcpComposer(props: {
  disabled: boolean;
  isProcessing: boolean;
  prompt: string;
  onChange: (value: string) => void;
  onInterrupt: () => void;
  onSend: () => void;
}) {
  return (
    <div className="safe-bottom border-t border-border/60 p-3">
      <div className="flex items-end gap-2">
        <textarea
          value={props.prompt}
          onChange={(event) => props.onChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !event.shiftKey) {
              event.preventDefault();
              props.onSend();
            }
          }}
          placeholder={props.disabled ? "接管输入后可发送消息" : "输入任务或问题"}
          disabled={props.disabled}
          rows={1}
          className="min-h-11 flex-1 resize-none rounded-md border border-border bg-surface px-3 py-2 text-sm leading-5 outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 disabled:opacity-60"
        />
        {props.isProcessing ? (
          <button
            onClick={props.onInterrupt}
            aria-label="中断 Worker"
            disabled={props.disabled}
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-destructive text-destructive-foreground disabled:opacity-50"
          >
            <Square className="h-4 w-4" />
          </button>
        ) : (
          <button
            onClick={props.onSend}
            aria-label="发送消息"
            disabled={props.disabled || !props.prompt.trim()}
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-50"
          >
            <Send className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
