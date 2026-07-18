export type VideoTaskStatusText = Record<
  | "failed"
  | "partial"
  | "processing"
  | "task_failed"
  | "verified"
  | "verification_failed",
  { detail: string; title: string }
>;

export const englishVideoTaskStatus: VideoTaskStatusText = {
  failed: {
    title: "Video generation failed",
    detail: "No verified playable video was produced.",
  },
  partial: {
    title: "Video file available, but the task did not finish cleanly",
    detail: "The published file remains available; a later step failed.",
  },
  processing: {
    title: "Creating video",
    detail: "The result will appear after generation and file verification.",
  },
  task_failed: {
    title: "Task failed",
    detail: "No verified result is available.",
  },
  verified: {
    title: "Video file published and integrity checked",
    detail: "The platform verified the published playable file, not its provider origin.",
  },
  verification_failed: {
    title: "Video result failed verification",
    detail: "The preview is hidden because the playable file metadata is incomplete.",
  },
};

export const chineseVideoTaskStatus: VideoTaskStatusText = {
  failed: {
    title: "视频生成失败",
    detail: "没有生成通过校验的可播放视频。",
  },
  partial: {
    title: "视频文件可用，但任务未完整结束",
    detail: "已发布的视频文件仍可使用；后续步骤执行失败。",
  },
  processing: {
    title: "正在生成视频",
    detail: "生成完成并通过文件校验后，成果会显示在右侧。",
  },
  task_failed: {
    title: "任务执行失败",
    detail: "当前没有可验证的成果。",
  },
  verified: {
    title: "视频文件已发布并通过完整性校验",
    detail: "平台已校验发布文件；此状态不代表供应商来源已被证明。",
  },
  verification_failed: {
    title: "视频成果校验未通过",
    detail: "可播放文件信息不完整，平台已停止展示预览。",
  },
};
