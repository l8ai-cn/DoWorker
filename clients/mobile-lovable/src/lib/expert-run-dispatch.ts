import { createSession, type SessionCreationAgent } from "./sessions-api";
import { getLiveExpert, runLiveExpert, sessionIdFromPod, type LiveExpert } from "./experts-api";

export type ExpertDispatchResult =
  { kind: "session"; sessionId: string } | { kind: "pod_only"; podKey: string; message: string };

function promptForExpert(expert: LiveExpert, prompt: string): string {
  const base = prompt.trim();
  if (expert.prompt?.trim()) return `${expert.prompt.trim()}\n\n${base}`;
  return base;
}

export async function dispatchLiveExpertTask(
  expert: LiveExpert,
  prompt: string,
  worker: SessionCreationAgent,
): Promise<ExpertDispatchResult> {
  if (worker.id !== expert.agent_slug) {
    throw new Error(`专家 ${expert.slug} 与 Worker ${worker.id} 不匹配`);
  }
  const text = promptForExpert(expert, prompt);
  const title = `${expert.name} · ${prompt.trim().slice(0, 48)}`.slice(0, 80);

  if (expert.interaction_mode === "acp") {
    const session = await createSession(worker, title, text);
    return { kind: "session", sessionId: session.id };
  }

  const { pod, warning } = await runLiveExpert(expert.slug, text);
  const sessionId = sessionIdFromPod(pod);
  if (sessionId) {
    return { kind: "session", sessionId };
  }
  return {
    kind: "pod_only",
    podKey: pod.pod_key,
    message: warning ?? `Pod ${pod.pod_key} 已创建（PTY 模式）。会话链接就绪后可在首页看到。`,
  };
}

export async function dispatchExpertBySlug(
  slug: string,
  prompt: string,
  worker: SessionCreationAgent,
): Promise<ExpertDispatchResult | null> {
  const expert = await getLiveExpert(slug);
  if (!expert) return null;
  return dispatchLiveExpertTask(expert, prompt, worker);
}
