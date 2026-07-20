import type {
  UserVideoExecutionDetail,
  UserVideoExecutionStepId,
} from "./userVideoExecutionTrace";

export interface VideoExecutionTraceText {
  detail: Record<Exclude<UserVideoExecutionDetail, "rendering">, string>;
  label: string;
  rendering(progress?: number): string;
  step: Record<UserVideoExecutionStepId, string>;
}

export const englishVideoExecutionTrace: VideoExecutionTraceText = {
  label: "Video creation progress",
  step: {
    request: "Request received",
    generation: "Generate video",
    preview: "Prepare preview",
    verification: "Verify and publish",
  },
  detail: {
    generation_failed: "Video generation did not finish",
    generation_ready: "Video frames are ready",
    preview_failed: "Preview preparation did not finish",
    preview_ready: "Playable file is ready",
    preparing_preview: "Preparing a playable file",
    provider_auth_failed: "Video credentials were rejected",
    provider_unavailable: "No video account is currently available",
    published: "File checked and published",
    queued: "Waiting to start",
    request_received: "Video creation request accepted",
    verification_failed: "File verification did not pass",
    verification_incomplete: "Task ended before file verification",
    verifying: "Checking the playable file",
    waiting: "Waiting for the previous step",
  },
  rendering: (progress) =>
    progress === undefined
      ? "Generating video"
      : `Generating video (${Math.round(progress * 100)}%)`,
};

export const chineseVideoExecutionTrace: VideoExecutionTraceText = {
  label: "视频创作进度",
  step: {
    request: "接收创作请求",
    generation: "生成视频画面",
    preview: "准备播放文件",
    verification: "校验并发布结果",
  },
  detail: {
    generation_failed: "视频生成未完成",
    generation_ready: "视频画面已生成",
    preview_failed: "播放文件准备未完成",
    preview_ready: "可播放文件已准备",
    preparing_preview: "正在准备可播放文件",
    provider_auth_failed: "视频服务凭据校验失败",
    provider_unavailable: "视频服务当前没有可用账号",
    published: "文件已校验并发布",
    queued: "等待开始生成",
    request_received: "已接收视频创作请求",
    verification_failed: "文件校验未通过",
    verification_incomplete: "任务结束前未完成文件校验",
    verifying: "正在校验可播放文件",
    waiting: "等待上一步完成",
  },
  rendering: (progress) =>
    progress === undefined
      ? "正在生成视频"
      : `正在生成视频（${Math.round(progress * 100)}%）`,
};
