"use client";

import { useEffect, useMemo, useState } from "react";
import { Bot, Loader2, Play, Sparkles, Video } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  getPodWorkerContext,
  type PodWorkerContext,
} from "@/lib/api/podWorkerContext";
import {
  sendPodPrompt,
  updatePodPreviewConfig,
} from "@/lib/api/facade/podConnect";
import type { PodData } from "@/lib/api/facade/pod";
import { usePodStore } from "@/stores/pod";

const VIDEO_PREVIEW_PORT = 4173;
const VIDEO_PREVIEW_PATH = "/oilan-video/index.html";

function videoDeliveryContract(task: string): string {
  return `${task.trim()}

视频交付合同：
1. 最终成片必须输出为 delivery/oilan-video-preview.mp4。
2. 使用 ffprobe 和完整解码检查验证 MP4，失败时不得报告完成。
3. 创建 delivery/oilan-video/index.html，页面必须包含可见标题“OILAN 视频成片”、原生 video controls 播放器，并加载 ../oilan-video-preview.mp4。
4. 重启预览服务：若 delivery/preview-server.pid 存在则终止旧进程；随后执行 nohup python3 -m http.server ${VIDEO_PREVIEW_PORT} --bind 127.0.0.1 --directory delivery > delivery/preview-server.log 2>&1 &，把 PID 写入 delivery/preview-server.pid。
5. 确认 http://127.0.0.1:${VIDEO_PREVIEW_PORT}/ 可访问后，报告成片规格、校验结果和预览入口已就绪。`;
}

function isActivePod(status: PodData["status"]): boolean {
  return ["initializing", "running", "paused", "disconnected"].includes(status);
}

interface WorkerTabContentProps {
  selectedPodKey: string | null;
  pod: PodData | null;
  orgSlug: string;
  t: (key: string, params?: Record<string, string | number>) => string;
}

export function WorkerTabContent({
  selectedPodKey,
  pod,
  orgSlug,
  t,
}: WorkerTabContentProps) {
  const [worker, setWorker] = useState<PodWorkerContext | null>(null);
  const [loadingExpert, setLoadingExpert] = useState(false);
  const [task, setTask] = useState("");
  const [sending, setSending] = useState(false);

  useEffect(() => {
    if (!pod?.worker_spec_snapshot_id || !selectedPodKey || !orgSlug) {
      setWorker(null);
      return;
    }
    let cancelled = false;
    setLoadingExpert(true);
    getPodWorkerContext(orgSlug, selectedPodKey).then(
      (context) => {
        if (!cancelled) setWorker(context);
      },
      () => {
        if (!cancelled) setWorker(null);
      },
    ).finally(() => {
      if (!cancelled) setLoadingExpert(false);
    });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, pod?.worker_spec_snapshot_id, selectedPodKey]);

  const canSend = useMemo(
    () => Boolean(pod && orgSlug && task.trim() && isActivePod(pod.status)),
    [orgSlug, pod, task],
  );

  if (!selectedPodKey || !pod) {
    return (
      <div className="flex h-full items-center justify-center text-xs text-muted-foreground">
        {t("videoWorker.notFound")}
      </div>
    );
  }

  const handleSubmit = async () => {
    if (!canSend) return;
    setSending(true);
    try {
      await sendPodPrompt(orgSlug, pod.pod_key, videoDeliveryContract(task));
      const updated = await updatePodPreviewConfig(
        orgSlug,
        pod.pod_key,
        VIDEO_PREVIEW_PORT,
        VIDEO_PREVIEW_PATH,
      );
      usePodStore.getState().upsertPod(updated);
      setTask("");
      toast.success(t("videoWorker.taskSent"));
    } catch (cause) {
      toast.error(
        cause instanceof Error
          ? cause.message
          : t("videoWorker.taskFailed"),
      );
    } finally {
      setSending(false);
    }
  };

  return (
    <div className="grid h-full gap-3 overflow-auto lg:grid-cols-[minmax(220px,0.8fr)_minmax(320px,1.2fr)]">
      <section className="rounded-md border border-border/60 bg-surface-muted/30 p-3">
        <div className="flex items-start gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10">
            <Bot className="h-4 w-4 text-primary" />
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">
              {worker?.expert?.name ?? worker?.alias ?? pod.alias ?? pod.title ?? pod.pod_key}
            </p>
            <p className="font-mono text-[11px] text-muted-foreground">
              {worker?.expert?.slug ?? `worker-snapshot-${worker?.snapshot_id ?? pod.worker_spec_snapshot_id}`}
            </p>
          </div>
        </div>

        <div className="mt-3 flex items-center gap-1.5 text-xs font-medium">
          <Sparkles className="h-3.5 w-3.5 text-muted-foreground" />
          {t("videoWorker.skills")}
        </div>
        <div className="mt-2 flex flex-wrap gap-1.5">
          {loadingExpert && <Loader2 className="h-4 w-4 animate-spin" />}
          {!loadingExpert && (worker?.skill_slugs.length ?? 0) === 0 && (
            <span className="text-xs text-muted-foreground">
              {t("videoWorker.noSkills")}
            </span>
          )}
          {worker?.skill_slugs.map((skill) => (
            <Badge key={skill} variant="info" className="font-normal">
              {skill}
            </Badge>
          ))}
        </div>

        <div className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
          <Video className="h-3.5 w-3.5" />
          {pod.preview_port
            ? t("videoWorker.previewEnabled", {
                port: pod.preview_port,
              })
            : t("videoWorker.previewEnabled", {
                port: VIDEO_PREVIEW_PORT,
              })}
        </div>
      </section>

      <section className="flex min-h-0 flex-col rounded-md border border-border/60 bg-background p-3">
        <label htmlFor="worker-video-task" className="text-xs font-medium">
          {t("videoWorker.taskTitle")}
        </label>
        <Textarea
          id="worker-video-task"
          value={task}
          onChange={(event) => setTask(event.target.value)}
          placeholder={t("videoWorker.taskPlaceholder")}
          className="mt-2 min-h-20 flex-1 resize-none text-sm"
          disabled={sending || !isActivePod(pod.status)}
        />
        <div className="mt-2 flex justify-end">
          <Button size="sm" onClick={handleSubmit} disabled={!canSend || sending}>
            {sending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Play className="h-3.5 w-3.5" />
            )}
            {sending
              ? t("videoWorker.sendingTask")
              : t("videoWorker.sendTask")}
          </Button>
        </div>
      </section>
    </div>
  );
}
