export type VideoTaskStatusText = Record<
  | "failed"
  | "model_quota_exhausted"
  | "partial"
  | "processing"
  | "provider_auth_failed"
  | "provider_unavailable"
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
  model_quota_exhausted: {
    title: "Agent primary-model quota exhausted",
    detail: "No verified video file is available.",
  },
  partial: {
    title: "Video file available, but the task did not finish cleanly",
    detail: "The published file remains available; a later step failed.",
  },
  processing: {
    title: "Creating video",
    detail: "The result will appear after generation and file verification.",
  },
  provider_auth_failed: {
    title: "Video service credentials failed",
    detail: "The provider rejected the video credentials, so no video file was generated.",
  },
  provider_unavailable: {
    title: "Video service temporarily unavailable",
    detail: "The third-party video account pool has no available account, so no video file was generated.",
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
  model_quota_exhausted: {
    title: "智能体主模型额度耗尽",
    detail: "未取得可验证视频文件。",
  },
  partial: {
    title: "视频文件可用，但任务未完整结束",
    detail: "已发布的视频文件仍可使用；后续步骤执行失败。",
  },
  processing: {
    title: "正在生成视频",
    detail: "生成完成并通过文件校验后，成果会显示在右侧。",
  },
  provider_auth_failed: {
    title: "视频服务凭据无效",
    detail: "供应商拒绝了当前视频生成凭据，未生成视频文件。",
  },
  provider_unavailable: {
    title: "视频服务暂不可用",
    detail: "第三方视频账号池当前没有可用账号，未生成视频文件。",
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
