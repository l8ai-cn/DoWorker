import { apiFetch } from "./api-fetch";

export interface UploadedFile {
  id: string;
  filename: string;
  bytes: number;
}

export async function uploadSessionFile(sessionId: string, file: File): Promise<UploadedFile> {
  const form = new FormData();
  form.append("file", file, file.name || "image.png");
  const res = await apiFetch(
    `/v1/sessions/${encodeURIComponent(sessionId)}/resources/files`,
    { method: "POST", body: form },
  );
  if (!res.ok) throw new Error(await res.text());
  const resource = (await res.json()) as {
    id: string;
    name?: string;
    metadata?: { filename?: string; bytes?: number };
  };
  return {
    id: resource.id,
    filename: resource.metadata?.filename ?? resource.name ?? (file.name || "image.png"),
    bytes: resource.metadata?.bytes ?? file.size,
  };
}
