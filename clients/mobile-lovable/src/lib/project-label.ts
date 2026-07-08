export const PROJECT_LABEL_KEY = "omni_project";

export function projectIdFromName(name: string): string {
  return encodeURIComponent(name.trim());
}

export function projectNameFromId(id: string): string {
  try {
    return decodeURIComponent(id);
  } catch {
    return id;
  }
}
