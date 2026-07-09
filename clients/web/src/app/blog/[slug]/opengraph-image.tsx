import { ImageResponse } from "next/og";
import { getPost } from "@/lib/blog";

export const alt = "Do Worker Blog";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = await getPost("en", slug);

  const title = post?.title ?? "Blog Post";
  const date = post?.date
    ? new Date(post.date).toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
        day: "numeric",
      })
    : "";
  const author = post?.author ?? "";

  return new ImageResponse(
    (
      <div
        style={{
          background:
            "linear-gradient(135deg, #16130f 0%, #1f1b16 50%, #16130f 100%)",
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          fontFamily: "system-ui, sans-serif",
          position: "relative",
          overflow: "hidden",
          padding: 60,
        }}
      >
        {/* Grid pattern */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            opacity: 0.08,
            backgroundImage:
              "linear-gradient(rgba(15, 118, 110, 0.35) 1px, transparent 1px), linear-gradient(90deg, rgba(15, 118, 110, 0.35) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />
        {/* Glow */}
        <div
          style={{
            position: "absolute",
            top: "30%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            width: 600,
            height: 600,
            borderRadius: "50%",
            background:
              "radial-gradient(circle, rgba(45, 212, 191, 0.14) 0%, transparent 70%)",
          }}
        />

        {/* Header: Logo + Blog label */}
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: 12,
              background: "linear-gradient(135deg, #0F766E, #0B5F59)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontSize: 24,
              fontWeight: 800,
              color: "#ffffff",
            }}
          >
            DW
          </div>
          <span style={{ fontSize: 28, fontWeight: 600, color: "#a1a1aa" }}>
            Do Worker Blog
          </span>
        </div>

        {/* Title */}
        <div
          style={{
            display: "flex",
            flex: 1,
            alignItems: "center",
          }}
        >
          <div
            style={{
              fontSize: title.length > 60 ? 40 : 48,
              fontWeight: 700,
              color: "#ededed",
              lineHeight: 1.3,
              maxWidth: 1000,
            }}
          >
            {title}
          </div>
        </div>

        {/* Footer: date + author */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 24,
            fontSize: 22,
            color: "#71717a",
          }}
        >
          {date && <span>{date}</span>}
          {date && author && <span>·</span>}
          {author && <span>{author}</span>}
        </div>
      </div>
    ),
    { ...size },
  );
}
