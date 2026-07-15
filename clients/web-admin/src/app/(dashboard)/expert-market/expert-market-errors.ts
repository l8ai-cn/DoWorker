export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (error && typeof error === "object" && "error" in error) {
    return String((error as { error: unknown }).error);
  }
  return "请求失败，请重试";
}
