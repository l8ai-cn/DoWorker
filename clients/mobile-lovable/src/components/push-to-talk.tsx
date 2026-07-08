import { useEffect, useRef, useState } from "react";
import { Mic, Send, X } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * PushToTalk — 按住说话录音组件。
 *
 * 交互：
 *  - 按下（pointerdown）开始录音，显示波形和计时
 *  - 松开（pointerup）停止并进入"预览转写"状态，用户可确认发送 / 取消
 *  - 上滑取消（松开时 y 距离 > 60）
 *
 * 目前录音存本地为 blob，转写文字为 mock；接入真实 STT（例如 Lovable AI
 * openai/gpt-4o-mini-transcribe）时只需替换 transcribe()。
 */
export function PushToTalk({
  onSubmit,
  placeholder = "按住语音键说出你的需求",
}: {
  onSubmit: (text: string) => void;
  placeholder?: string;
}) {
  const [state, setState] = useState<"idle" | "recording" | "cancelling" | "preview">("idle");
  const [seconds, setSeconds] = useState(0);
  const [transcript, setTranscript] = useState("");
  const [levels, setLevels] = useState<number[]>(() => Array(24).fill(0.15));
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const rafRef = useRef<number | null>(null);
  const startYRef = useRef(0);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, []);

  function beginRecording(e: React.PointerEvent) {
    startYRef.current = e.clientY;
    setSeconds(0);
    setState("recording");
    timerRef.current = setInterval(() => setSeconds((s) => s + 1), 1000);
    const tick = () => {
      setLevels((prev) => prev.map(() => 0.15 + Math.random() * 0.85));
      rafRef.current = requestAnimationFrame(tick);
    };
    tick();
  }

  function updateRecording(e: React.PointerEvent) {
    if (state !== "recording" && state !== "cancelling") return;
    const dy = startYRef.current - e.clientY;
    setState(dy > 60 ? "cancelling" : "recording");
  }

  function endRecording() {
    if (timerRef.current) clearInterval(timerRef.current);
    if (rafRef.current) cancelAnimationFrame(rafRef.current);
    timerRef.current = null;
    rafRef.current = null;

    if (state === "cancelling") {
      setState("idle");
      setSeconds(0);
      return;
    }
    if (state !== "recording") return;

    // 模拟 STT 完成 —— 真实接入替换为 fetch('/api/transcribe', {body: audioBlob})
    const mocks = [
      "帮我 review 一下最新的 PR，重点看看 auth 那块的边界条件",
      "把 checkout 页面重新用 skeleton loader 做加载态",
      "查下昨天 3 点那次部署失败的原因，把 CI 日志摘要发我",
    ];
    setTranscript(mocks[Math.floor(Math.random() * mocks.length)]);
    setState("preview");
  }

  if (state === "preview") {
    return (
      <div className="stream-in rounded-2xl border border-primary/30 bg-primary/5 p-3">
        <div className="mb-2 flex items-center justify-between">
          <p className="text-[10.5px] font-semibold uppercase tracking-wider text-primary">
            语音转文字 · {seconds}s
          </p>
          <button
            onClick={() => {
              setState("idle");
              setTranscript("");
            }}
            className="flex h-6 w-6 items-center justify-center rounded-full hover:bg-surface-2"
          >
            <X className="h-3 w-3 text-muted-foreground" />
          </button>
        </div>
        <p className="text-[13.5px] leading-relaxed text-foreground">{transcript}</p>
        <div className="mt-3 flex gap-2">
          <button
            onClick={() => setState("idle")}
            className="flex-1 rounded-full border border-border bg-surface py-2 text-[12px] font-medium"
          >
            重录
          </button>
          <button
            onClick={() => {
              onSubmit(transcript);
              setState("idle");
              setTranscript("");
            }}
            className="flex flex-1 items-center justify-center gap-1.5 rounded-full bg-primary py-2 text-[12.5px] font-semibold text-primary-foreground glow-primary"
          >
            <Send className="h-3 w-3" />
            派发
          </button>
        </div>
      </div>
    );
  }

  const active = state === "recording" || state === "cancelling";

  return (
    <div className="flex flex-col items-center gap-2">
      {active && (
        <div
          className={cn(
            "stream-in flex w-full items-center gap-3 rounded-2xl px-4 py-3 ring-1",
            state === "cancelling"
              ? "bg-destructive/10 ring-destructive/40"
              : "bg-primary/10 ring-primary/40",
          )}
        >
          <div className="flex flex-1 items-center gap-[2px]">
            {levels.map((l, i) => (
              <span
                key={i}
                className={cn(
                  "w-1 rounded-full transition-[height]",
                  state === "cancelling" ? "bg-destructive" : "bg-primary",
                )}
                style={{ height: `${Math.max(4, l * 28)}px` }}
              />
            ))}
          </div>
          <span
            className={cn(
              "font-mono text-[12px] tabular-nums",
              state === "cancelling" ? "text-destructive" : "text-primary",
            )}
          >
            {String(Math.floor(seconds / 60)).padStart(2, "0")}:
            {String(seconds % 60).padStart(2, "0")}
          </span>
        </div>
      )}

      <button
        type="button"
        onPointerDown={beginRecording}
        onPointerMove={updateRecording}
        onPointerUp={endRecording}
        onPointerCancel={endRecording}
        onPointerLeave={(e) => e.buttons === 1 && endRecording()}
        className={cn(
          "flex w-full items-center justify-center gap-2 rounded-full py-3.5 text-[13.5px] font-semibold transition select-none touch-none",
          active
            ? state === "cancelling"
              ? "bg-destructive text-destructive-foreground scale-[1.02]"
              : "bg-primary text-primary-foreground scale-[1.02] glow-primary"
            : "bg-surface text-foreground/80 ring-1 ring-border hover:ring-primary/40",
        )}
      >
        <Mic className={cn("h-4 w-4", active && "animate-pulse")} />
        {state === "idle" && placeholder}
        {state === "recording" && "松开发送 · 上滑取消"}
        {state === "cancelling" && "松开取消"}
      </button>
      {!active && (
        <p className="text-center text-[10.5px] text-muted-foreground">
          按住说话 · 也可在下方直接输入文字
        </p>
      )}
    </div>
  );
}
