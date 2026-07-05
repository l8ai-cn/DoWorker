import type { ReleaseAsset, RunnerAsset } from "./asset-types";

const RUNNER_ARCHIVE = /^do-worker-runner_[\d.]+_(linux|darwin|windows)_(amd64|arm64)\.(tar\.gz|zip)$/i;

export function classifyRunner(asset: ReleaseAsset): RunnerAsset | null {
  const m = asset.name.match(RUNNER_ARCHIVE);
  if (!m) return null;
  const platform = m[1].toLowerCase() === "darwin" ? "macos" : (m[1].toLowerCase() as "linux" | "windows");
  const arch = m[2].toLowerCase() === "amd64" ? "x64" : "arm64";
  const kind = m[3].toLowerCase() === "zip" ? "zip" : "tarball";
  return { ...asset, platform, arch, kind };
}
