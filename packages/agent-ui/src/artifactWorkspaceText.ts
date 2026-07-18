export interface ArtifactWorkspaceText {
  generatedArtifact: string;
  generationFailed: string;
  loadingUnavailable: string;
  loadFailed: string;
  loading(filename: string): string;
  load(filename: string): string;
  retry(filename: string): string;
  preview(filename: string): string;
  videoPreview(filename: string): string;
  videoPlaybackFailed(filename: string): string;
  videoUnsupported: string;
  fullscreenVideo: string;
  exitFullscreen: string;
  open(filename: string): string;
  download(filename: string): string;
}

export function artifactWorkspaceText(
  locale: "en-US" | "zh-CN",
): ArtifactWorkspaceText {
  return locale === "zh-CN" ? zhCN : enUS;
}

const enUS: ArtifactWorkspaceText = {
  generatedArtifact: "Generated artifact",
  generationFailed: "Artifact generation failed",
  loadingUnavailable: "Artifact loading is unavailable",
  loadFailed: "Artifact loading failed. Try again.",
  loading: (filename) => `Loading ${filename}`,
  load: (filename) => `Load ${filename}`,
  retry: (filename) => `Retry loading ${filename}`,
  preview: (filename) => `Preview ${filename}`,
  videoPreview: (filename) => `Video preview for ${filename}`,
  videoPlaybackFailed: (filename) =>
    `Unable to play ${filename}. Try loading it again.`,
  videoUnsupported: "Your browser does not support video playback.",
  fullscreenVideo: "View video fullscreen",
  exitFullscreen: "Exit fullscreen",
  open: (filename) => `Open ${filename}`,
  download: (filename) => `Download ${filename}`,
};

const zhCN: ArtifactWorkspaceText = {
  generatedArtifact: "生成的成果",
  generationFailed: "成果生成失败",
  loadingUnavailable: "当前无法加载成果",
  loadFailed: "成果加载失败，请重试。",
  loading: (filename) => `正在加载 ${filename}`,
  load: (filename) => `加载 ${filename}`,
  retry: (filename) => `重新加载 ${filename}`,
  preview: (filename) => `预览 ${filename}`,
  videoPreview: (filename) => `${filename} 的视频预览`,
  videoPlaybackFailed: (filename) => `${filename} 无法播放，请重新加载。`,
  videoUnsupported: "当前浏览器不支持视频播放。",
  fullscreenVideo: "全屏预览视频",
  exitFullscreen: "退出全屏预览",
  open: (filename) => `打开 ${filename}`,
  download: (filename) => `下载 ${filename}`,
};
