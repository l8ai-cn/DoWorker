import { Maximize2, Minimize2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
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
  const surfaceRef = useRef<HTMLDivElement>(null);
  const [fullscreen, setFullscreen] = useState(false);
  const [fullscreenError, setFullscreenError] = useState(false);
  const fullscreenSupported =
    typeof HTMLElement !== "undefined" &&
    typeof HTMLElement.prototype.requestFullscreen === "function";

  useEffect(() => {
    const syncFullscreen = () => {
      setFullscreen(document.fullscreenElement === surfaceRef.current);
    };
    document.addEventListener("fullscreenchange", syncFullscreen);
    return () => document.removeEventListener("fullscreenchange", syncFullscreen);
  }, []);

  const toggleFullscreen = async () => {
    setFullscreenError(false);
    try {
      if (document.fullscreenElement === surfaceRef.current) {
        await document.exitFullscreen();
      } else {
        await surfaceRef.current?.requestFullscreen();
      }
    } catch {
      setFullscreenError(true);
    }
  };

  return (
    <div
      className={`relative bg-black ${fullscreen ? "flex h-full items-center justify-center" : ""}`}
      ref={surfaceRef}
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
      {fullscreenSupported && (
        <button
          aria-label={fullscreen ? text.exitFullscreen : text.fullscreenVideo}
          className="absolute right-3 top-3 inline-flex size-11 items-center justify-center rounded-md bg-black/70 text-white outline-none hover:bg-black/85 focus-visible:ring-2 focus-visible:ring-white"
          onClick={() => void toggleFullscreen()}
          title={fullscreen ? text.exitFullscreen : text.fullscreenVideo}
          type="button"
        >
          {fullscreen ? (
            <Minimize2 aria-hidden="true" className="size-4" />
          ) : (
            <Maximize2 aria-hidden="true" className="size-4" />
          )}
        </button>
      )}
      {fullscreenError && (
        <div
          className="absolute inset-x-3 bottom-3 rounded-md bg-black/80 px-3 py-2 text-sm text-white"
          role="alert"
        >
          {text.fullscreenUnavailable}
        </div>
      )}
    </div>
  );
}
