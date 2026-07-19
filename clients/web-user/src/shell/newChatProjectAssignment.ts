import type { QueryClient } from "@tanstack/react-query";
import { PROJECT_LABEL_KEY } from "@/hooks/useConversations";
import { authenticatedFetch } from "@/lib/identity";

export async function assignNewChatProject(
  sessionId: string,
  projectId: string,
  queryClient: QueryClient,
) {
  if (!projectId) return;
  const response = await authenticatedFetch(`/v1/sessions/${sessionId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ labels: { [PROJECT_LABEL_KEY]: projectId } }),
  });
  if (!response.ok) {
    const detail = await response.text().catch(() => "");
    throw new Error(detail.trim() || `Could not assign project (${response.status})`);
  }
  void queryClient.invalidateQueries({ queryKey: ["projects"] });
  void queryClient.invalidateQueries({ queryKey: ["project-sessions"] });
}
