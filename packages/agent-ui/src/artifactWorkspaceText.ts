export interface ArtifactWorkspaceText {
  generatedArtifact: string;
  generationFailed: string;
  loadingUnavailable: string;
  downloadNotAuthorized: string;
  loadFailed: string;
  loading(filename: string): string;
  load(filename: string): string;
  retry(filename: string): string;
  preview(filename: string): string;
  audioPreview(filename: string): string;
  csvPreview(filename: string): string;
  csvPreviewFailed(filename: string): string;
  previewLimited(bytes: number): string;
  pdfPreview(filename: string): string;
  videoPreview(filename: string): string;
  videoPlaybackFailed(filename: string): string;
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
  downloadNotAuthorized: "You do not have permission to access this artifact",
  loadFailed: "Artifact loading failed. Try again.",
  loading: (filename) => `Loading ${filename}`,
  load: (filename) => `Load ${filename}`,
  retry: (filename) => `Retry loading ${filename}`,
  preview: (filename) => `Preview ${filename}`,
  audioPreview: (filename) => `Audio preview for ${filename}`,
  csvPreview: (filename) => `CSV preview for ${filename}`,
  csvPreviewFailed: (filename) => `Unable to preview ${filename} as CSV.`,
  previewLimited: (bytes) =>
    `Preview limited to the first ${Math.floor(bytes / 1024 / 1024)} MiB.`,
  pdfPreview: (filename) => `PDF preview for ${filename}`,
  videoPreview: (filename) => `Video preview for ${filename}`,
  videoPlaybackFailed: (filename) =>
    `Unable to play ${filename}. Try loading it again.`,
  open: (filename) => `Open ${filename}`,
  download: (filename) => `Download ${filename}`,
};

const zhCN: ArtifactWorkspaceText = {
  generatedArtifact: "生成的成果",
  generationFailed: "成果生成失败",
  loadingUnavailable: "当前无法加载成果",
  downloadNotAuthorized: "你没有权限访问此成果",
  loadFailed: "成果加载失败，请重试。",
  loading: (filename) => `正在加载 ${filename}`,
  load: (filename) => `加载 ${filename}`,
  retry: (filename) => `重新加载 ${filename}`,
  preview: (filename) => `预览 ${filename}`,
  audioPreview: (filename) => `${filename} 的音频预览`,
  csvPreview: (filename) => `${filename} 的 CSV 预览`,
  csvPreviewFailed: (filename) => `${filename} 无法作为 CSV 预览。`,
  previewLimited: (bytes) =>
    `仅预览前 ${Math.floor(bytes / 1024 / 1024)} MiB 内容。`,
  pdfPreview: (filename) => `${filename} 的 PDF 预览`,
  videoPreview: (filename) => `${filename} 的视频预览`,
  videoPlaybackFailed: (filename) => `${filename} 无法播放，请重新加载。`,
  open: (filename) => `打开 ${filename}`,
  download: (filename) => `下载 ${filename}`,
};
