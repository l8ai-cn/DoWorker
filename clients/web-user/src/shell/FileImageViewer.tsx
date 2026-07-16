import { useEffect, useState } from "react";
import { useLightbox } from "@/components/ImageLightbox";
import {
  type FileContentResponse,
  fileContentToBlob,
} from "@/hooks/useFileContent";
import { TruncatedBanner } from "./TruncatedBanner";

const CHECKERBOARD_STYLE: React.CSSProperties = {
  backgroundImage:
    "linear-gradient(45deg, rgba(128,128,128,0.15) 25%, transparent 25%)," +
    "linear-gradient(-45deg, rgba(128,128,128,0.15) 25%, transparent 25%)," +
    "linear-gradient(45deg, transparent 75%, rgba(128,128,128,0.15) 75%)," +
    "linear-gradient(-45deg, transparent 75%, rgba(128,128,128,0.15) 75%)",
  backgroundSize: "16px 16px",
  backgroundPosition: "0 0, 0 8px, 8px -8px, -8px 0",
};

interface FileImageViewerProps {
  data: FileContentResponse;
  path: string;
}

export function FileImageViewer({ data, path }: FileImageViewerProps) {
  const [url, setUrl] = useState<string | null>(null);
  const [errored, setErrored] = useState(false);
  const { open } = useLightbox();

  useEffect(() => {
    if (data.truncated) {
      setUrl(null);
      setErrored(true);
      return;
    }
    setErrored(false);
    const objectUrl = URL.createObjectURL(fileContentToBlob(data));
    setUrl(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [data]);

  const filename = path.split("/").pop() ?? path;
  const body = errored ? (
    <div className="flex items-center justify-center p-8 text-muted-foreground text-sm">
      {data.truncated
        ? "Image is too large to preview (truncated by the server)."
        : "Unable to render image."}
    </div>
  ) : (
    <div className="flex min-h-0 flex-1 items-center justify-center overflow-auto p-4">
      {url && (
        <img
          src={url}
          alt={filename}
          onError={() => setErrored(true)}
          onClick={() => open({ src: url, alt: filename })}
          className="max-h-full max-w-full cursor-zoom-in object-contain"
          style={CHECKERBOARD_STYLE}
          title="Click to zoom"
        />
      )}
    </div>
  );

  if (!data.truncated) return body;
  return (
    <div className="flex h-full flex-col">
      <TruncatedBanner />
      {body}
    </div>
  );
}
