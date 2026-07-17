import { Loader2, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { useMobileAcpRelay } from "@/hooks/use-mobile-acp-relay";
import { AcpComposer } from "@/components/mobile-acp-composer";
import { AcpMessageList, AcpPermissions } from "@/components/mobile-acp-content";

export function MobileAcpPanel({ podKey }: { podKey: string }) {
  const relay = useMobileAcpRelay(podKey);
  const [prompt, setPrompt] = useState("");
  const [sending, setSending] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const hasControl = relay.lease.status === "granted";
  const isProcessing = relay.session.state === "processing";

  const sendPrompt = async () => {
    if (!prompt.trim() || sending) return;
    setSending(true);
    setActionError(null);
    try {
      await relay.sendPrompt(prompt.trim());
      setPrompt("");
    } catch (cause) {
      setActionError(cause instanceof Error ? cause.message : "消息发送失败");
    } finally {
      setSending(false);
    }
  };

  const respond = async (requestId: string, approved: boolean) => {
    setActionError(null);
    try {
      await relay.respondPermission(requestId, approved);
    } catch (cause) {
      setActionError(cause instanceof Error ? cause.message : "权限响应失败");
    }
  };

  return (
    <div className="flex min-h-0 flex-1 flex-col bg-background">
      <AcpConnectionBar
        connected={relay.connection === "connected"}
        hasControl={hasControl}
        acquiring={relay.control.acquiring}
        onAcquire={() => void relay.control.acquire()}
        onReconnect={relay.reconnect}
      />
      <AcpMessageList messages={relay.session.messages} />
      <AcpPermissions
        disabled={!hasControl}
        permissions={relay.session.pendingPermissions}
        onRespond={respond}
      />
      {(relay.error || relay.control.error || actionError) && (
        <p className="border-t border-destructive/20 bg-destructive/10 px-3 py-2 text-xs text-destructive">
          {relay.error ?? relay.control.error ?? actionError}
        </p>
      )}
      <AcpComposer
        disabled={!hasControl || relay.connection !== "connected" || sending}
        isProcessing={isProcessing}
        prompt={prompt}
        onChange={setPrompt}
        onInterrupt={() => void relay.interrupt().catch((cause) => setActionError(String(cause)))}
        onSend={() => void sendPrompt()}
      />
    </div>
  );
}

function AcpConnectionBar(props: {
  connected: boolean;
  hasControl: boolean;
  acquiring: boolean;
  onAcquire: () => void;
  onReconnect: () => void;
}) {
  if (!props.connected) {
    return (
      <div className="flex min-h-12 items-center justify-between border-b border-border/60 px-3">
        <span className="text-xs text-muted-foreground">正在连接 Worker…</span>
        <button onClick={props.onReconnect} className="text-xs font-medium text-primary">
          重试
        </button>
      </div>
    );
  }
  if (props.hasControl) {
    return (
      <div className="flex min-h-12 items-center gap-2 border-b border-border/60 px-3 text-xs">
        <ShieldCheck className="h-4 w-4 text-success" />
        <span className="font-medium text-success">输入已接管</span>
      </div>
    );
  }
  return (
    <div className="flex min-h-12 items-center justify-between border-b border-border/60 px-3">
      <span className="text-xs text-muted-foreground">只读观察</span>
      <button
        onClick={props.onAcquire}
        disabled={props.acquiring}
        className="flex min-h-9 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground disabled:opacity-50"
      >
        {props.acquiring && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
        接管输入
      </button>
    </div>
  );
}
