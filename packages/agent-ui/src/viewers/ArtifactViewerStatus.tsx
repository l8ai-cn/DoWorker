import { ArtifactError } from "../GenericArtifactCard";

export function ArtifactViewerLoading({ filename }: { filename: string }) {
  return (
    <article
      className="rounded-md border border-border bg-muted/30 px-3 py-3 text-sm text-muted-foreground"
      role="status"
    >
      正在加载 {filename}
    </article>
  );
}

export function ArtifactViewerError({
  filename,
  message,
}: {
  filename: string;
  message: string;
}) {
  return <ArtifactError filename={filename} message={message} />;
}
