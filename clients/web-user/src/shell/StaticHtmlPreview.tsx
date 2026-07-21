import {
  STATIC_HTML_REFERRER_POLICY,
  STATIC_HTML_SANDBOX,
  staticHtmlDocument,
} from "@agent-cloud/agent-ui";

interface StaticHtmlPreviewProps {
  content: string;
}

export function StaticHtmlPreview({ content }: StaticHtmlPreviewProps) {
  return (
    <iframe
      srcDoc={staticHtmlDocument(content)}
      sandbox={STATIC_HTML_SANDBOX}
      referrerPolicy={STATIC_HTML_REFERRER_POLICY}
      title="HTML preview"
      className="h-full w-full border-0"
    />
  );
}
