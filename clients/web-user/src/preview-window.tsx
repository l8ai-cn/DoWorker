import { createRoot } from "react-dom/client";
import { readPreviewWindowSessionUrl } from "./lib/previewWindowSession";
import { requirePreviewPublicOrigin } from "./lib/previewSessionUrl";
import "./index.css";

function PreviewWindow() {
  let sessionUrl: string;
  try {
    sessionUrl = readPreviewWindowSessionUrl(
      window.location.hash,
      requirePreviewPublicOrigin(),
    );
  } catch {
    return (
      <main className="grid min-h-dvh place-items-center bg-background p-6 text-foreground">
        <div className="max-w-sm text-center">
          <h1 className="text-base font-semibold">无法打开预览</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            预览地址无效或已失效，请回到工作区重新打开。
          </p>
        </div>
      </main>
    );
  }

  return (
    <iframe
      title="工作区应用预览"
      src={sessionUrl}
      sandbox="allow-scripts allow-same-origin allow-forms allow-downloads"
      referrerPolicy="no-referrer"
      allow="fullscreen 'self'"
      allowFullScreen
      className="fixed inset-0 h-dvh w-full border-0 bg-background"
    />
  );
}

createRoot(document.getElementById("root")!).render(<PreviewWindow />);
