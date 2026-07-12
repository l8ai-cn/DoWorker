import { apiFetch } from "./api-fetch";

interface ModelResource {
  id: number;
  is_default: boolean;
}

interface ModelResourceList {
  object: "list";
  data: ModelResource[];
}

async function readJson<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const message = (await res.text()).trim();
    throw new Error(message || `Failed to load AI resources (${res.status})`);
  }
  return res.json() as Promise<T>;
}

export async function resolveDefaultModelResourceId(): Promise<number> {
  const response = await apiFetch("/v1/model-resources");
  const resources = await readJson<ModelResourceList>(response);
  const defaults = (resources.data ?? []).filter(
    (resource) => resource.is_default && resource.id > 0,
  );
  if (defaults.length !== 1) {
    throw new Error("No default model resource is configured");
  }
  return defaults[0].id;
}
