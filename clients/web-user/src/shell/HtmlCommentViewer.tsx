import { TruncatedBanner } from "./TruncatedBanner";
import { StaticHtmlPreview } from "./StaticHtmlPreview";

interface HtmlCommentViewerProps {
  content: string;
  truncated: boolean;
}

export function HtmlCommentViewer({ content, truncated }: HtmlCommentViewerProps) {
  return (
    <div className="flex h-full flex-col">
      {truncated && <TruncatedBanner />}
      <div className="min-h-0 flex-1">
        <StaticHtmlPreview content={content} />
      </div>
    </div>
  );
}
