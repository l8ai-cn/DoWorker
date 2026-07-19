export async function describeCreateError(res: Response): Promise<string> {
  try {
    const body: unknown = await res.json();
    if (body && typeof body === "object") {
      const data = body as Record<string, unknown>;
      if (typeof data.detail === "string") return data.detail;
      if (
        Array.isArray(data.detail) &&
        data.detail.length > 0 &&
        typeof (data.detail[0] as Record<string, unknown>)?.msg === "string"
      ) {
        return (data.detail[0] as Record<string, unknown>).msg as string;
      }
      if (typeof data.message === "string") return data.message;
      const err = data.error;
      if (typeof err === "string") return err;
      if (
        err &&
        typeof err === "object" &&
        typeof (err as Record<string, unknown>).message === "string"
      ) {
        return (err as Record<string, unknown>).message as string;
      }
    }
  } catch {
    return `Couldn't create the session (HTTP ${res.status}).`;
  }
  return `Couldn't create the session (HTTP ${res.status}).`;
}
