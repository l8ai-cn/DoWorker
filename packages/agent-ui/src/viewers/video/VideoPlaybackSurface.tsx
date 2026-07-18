import { Maximize2, Minimize2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useAgentWorkspaceText } from "../../AgentWorkspaceLocaleContext";

interface VideoPlaybackSurfaceProps {
  filename: string;
  posterSrc?: string;
  src: string;
}

export function VideoPlaybackSurface({
  filename,
  posterSrc,
  src,
}: VideoPlaybackSurfaceProps) {
  const text = useAgentWorkspaceText().artifact;
  const [fullscreen, setFullscreen] = useState(false);

  useEffect(() => {
    if (!fullscreen) return;
    const exitOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setFullscreen(false);
    };
    document.addEventListener("keydown", exitOnEscape);
    return () => document.removeEventListener("keydown", exitOnEscape);
  }, [fullscreen]);

  return (
    <div
      className={`relative bg-black ${fullscreen ? "flex items-center justify-center" : ""}`}
      style={
        fullscreen
          ? {
              inset: 0,
              position: "fixed",
              zIndex: 100,
            }
          : undefined
      }
    >
      <video
        aria-label={`视频预览：${filename}`}
        className={`w-full object-contain ${
          fullscreen ? "h-full max-h-none" : "aspect-video max-h-[70vh]"
        }`}
        controls
        playsInline
        poster={posterSrc}
        preload="metadata"
        src={src}
      >
        {text.videoUnsupported}
      </video>
      <button
        aria-label={fullscreen ? text.exitFullscreen : text.fullscreenVideo}
        className="absolute right-3 top-3 inline-flex size-11 items-center justify-center rounded-md bg-black/70 text-white outline-none hover:bg-black/85 focus-visible:ring-2 focus-visible:ring-white"
        onClick={() => setFullscreen((current) => !current)}
        title={fullscreen ? text.exitFullscreen : text.fullscreenVideo}
        type="button"
      >
        {fullscreen ? (
          <Minimize2 aria-hidden="true" className="size-4" />
        ) : (
          <Maximize2 aria-hidden="true" className="size-4" />
        )}
      </button>
    </div>
  );
}
