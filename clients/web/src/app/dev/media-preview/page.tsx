"use client";

import { Markdown } from "@/components/ui/markdown";
import { HtmlPreviewCard } from "@/components/media/HtmlPreviewCard";
import { LightboxImage } from "@/components/media/MediaLightbox";
import { VideoEmbed } from "@/components/media/VideoEmbed";
import { AttachmentCard } from "@/components/channel/AttachmentCard";
import { StructuredContent } from "@/components/channel/StructuredContent";
import type { MessageContent } from "@/lib/viewModels/channelMessage";

const SAMPLE_HTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>AI Preview</title>
    <style>
      body { font-family: system-ui, sans-serif; padding: 24px; background: linear-gradient(135deg, #667eea, #764ba2); color: white; }
      h1 { margin: 0 0 8px; }
      p { opacity: 0.9; }
    </style>
  </head>
  <body>
    <h1>Hello from AI</h1>
    <p>This page is rendered inside a sandboxed iframe.</p>
    <script>document.body.insertAdjacentHTML("beforeend", "<p id=js>JS works ✓</p>");</script>
  </body>
</html>`;

const MARKDOWN_SAMPLE = `## Media preview demo

![sample image](https://picsum.photos/seed/agentsmesh/480/240)

https://www.youtube.com/watch?v=dQw4w9WgXcQ

\`\`\`html
${SAMPLE_HTML}
\`\`\`
`;

const STRUCTURED_HTML: MessageContent = {
  kind: "markdown",
  schema_version: 1,
  blocks: [
    {
      type: "code_block",
      language: "html",
      text: SAMPLE_HTML,
    },
    {
      type: "paragraph",
      elements: [{ type: "link", url: "https://picsum.photos/seed/channel/360/200", display: "image link" }],
    },
  ],
};

export default function MediaPreviewDevPage() {
  return (
    <main className="mx-auto max-w-3xl space-y-10 p-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">Chat media preview</h1>
        <p className="text-sm text-muted-foreground">
          Local dev page for verifying lightbox, video embeds, and HTML sandbox previews.
        </p>
      </header>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">ACP Markdown (`enableMedia`)</h2>
        <div className="rounded-lg border border-border p-4">
          <Markdown content={MARKDOWN_SAMPLE} enableMedia />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">HtmlPreviewCard</h2>
        <HtmlPreviewCard html={SAMPLE_HTML} />
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">LightboxImage</h2>
        <LightboxImage
          src="https://picsum.photos/seed/lightbox/640/360"
          alt="Lightbox sample"
          className="max-w-sm"
          imgClassName="max-h-48 w-full object-cover"
        />
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">VideoEmbed</h2>
        <VideoEmbed url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" kind="youtube" />
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">AttachmentCard</h2>
        <div className="space-y-2 rounded-lg border border-border p-4">
          <AttachmentCard url="https://picsum.photos/seed/attach/400/220" />
          <AttachmentCard url="https://www.w3schools.com/html/mov_bbb.mp4" />
          <AttachmentCard url="https://www.w3schools.com/html/horse.mp3" />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-medium">StructuredContent</h2>
        <div className="rounded-lg border border-border p-4">
          <StructuredContent content={STRUCTURED_HTML} />
        </div>
      </section>
    </main>
  );
}
