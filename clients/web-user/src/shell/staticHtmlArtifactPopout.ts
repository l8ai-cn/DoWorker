import { openStaticHtmlInNewWindow } from "@do-worker/agent-ui";
import { showToast } from "@/components/ui/toast";

function popupBlockedMessage(): string {
  return document.documentElement.lang.toLowerCase().startsWith("zh")
    ? "浏览器阻止了新窗口，请允许此站点打开弹出窗口后重试。"
    : "The browser blocked the new window. Allow popups for this site and try again.";
}

export function openHtmlArtifactInNewTab(
  content: string,
  filename: string,
  opener: Pick<Window, "open"> = window,
): boolean {
  const result = openStaticHtmlInNewWindow(content, filename, opener);
  if (!result.opened) showToast(popupBlockedMessage());
  return result.opened;
}
