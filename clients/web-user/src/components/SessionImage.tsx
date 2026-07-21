import { SessionResourceImage } from "@agent-cloud/agent-ui";

import { authenticatedFetch } from "@/lib/identity";

export interface SessionImageProps {
  path?: string;
  alt: string;
  className?: string;
}

export function SessionImage({ path, alt, className }: SessionImageProps) {
  return (
    <SessionResourceImage
      alt={alt}
      className={className}
      loadBlob={loadSessionImageBlob}
      path={path}
    />
  );
}

async function loadSessionImageBlob(path: string): Promise<Blob> {
  const response = await authenticatedFetch(path);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.blob();
}
