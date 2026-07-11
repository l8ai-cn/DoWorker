import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

vi.mock("@/lib/identity", () => ({
  authenticatedFetch: vi.fn(),
}));
vi.mock("@/lib/do-worker", () => ({
  readDoWorkerOrgSlug: vi.fn(),
}));

import { authenticatedFetch } from "@/lib/identity";
import { readDoWorkerOrgSlug } from "@/lib/do-worker";
import { buildPreviewSrc, usePodPreview } from "./usePodPreview";

const fetchMock = vi.mocked(authenticatedFetch);
const orgSlugMock = vi.mocked(readDoWorkerOrgSlug);

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "",
    json: async () => body,
  } as unknown as Response;
}

function Wrap({ children }: { children: ReactNode }) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: 0 } },
  });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

beforeEach(() => {
  fetchMock.mockReset();
  orgSlugMock.mockReset();
  orgSlugMock.mockReturnValue("acme");
});

afterEach(() => {
  cleanup();
});

describe("buildPreviewSrc", () => {
  it("uses the session url as iframe source", () => {
    const src = buildPreviewSrc({
      preview_base_url: "https://d/preview/pod1/",
      session_url: "https://d/preview/pod1/__session?token=JWT",
      expires_at: "",
    });
    expect(src).toBe("https://d/preview/pod1/__session?token=JWT");
  });
});

describe("usePodPreview", () => {
  it("fetches the org-scoped preview endpoint and returns only the preview contract", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(200, {
        preview_base_url: "https://d/preview/pod1/",
        session_url: "https://d/preview/pod1/__session?token=JWT",
        expires_at: new Date(Date.now() + 30 * 60_000).toISOString(),
      }),
    );

    const { result } = renderHook(() => usePodPreview("pod1"), { wrapper: Wrap });

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data?.session_url).toContain("__session");
    expect(Object.keys(result.current.data ?? {}).sort()).toEqual([
      "expires_at",
      "preview_base_url",
      "session_url",
    ]);
    const [url] = fetchMock.mock.calls[0];
    expect(url).toBe("/api/v1/orgs/acme/pods/pod1/preview");
  });

  it("normalizes legacy preview payloads by dropping unknown fields", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(200, {
        preview_base_url: "https://d/preview/pod1/",
        session_url: "https://d/preview/pod1/__session?token=legacy",
        expires_at: "2026-07-12T00:00:00Z",
        token: "legacy",
        debug: "should be dropped",
      }),
    );

    const { result } = renderHook(() => usePodPreview("pod1"), { wrapper: Wrap });

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data).toEqual({
      preview_base_url: "https://d/preview/pod1/",
      session_url: "https://d/preview/pod1/__session?token=legacy",
      expires_at: "2026-07-12T00:00:00Z",
    });
    expect(Object.prototype.hasOwnProperty.call(result.current.data, "token")).toBe(false);
  });

  it("does not fetch when disabled or missing an org slug", () => {
    orgSlugMock.mockReturnValue(null);
    renderHook(() => usePodPreview("pod1"), { wrapper: Wrap });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("surfaces a 403 as an error rather than throwing an opaque failure", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(403, { error: "forbidden" }));

    const { result } = renderHook(() => usePodPreview("pod1"), { wrapper: Wrap });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as Error & { status?: number }).status).toBe(403);
  });

  it("schedules a refresh before the token expires", async () => {
    vi.useFakeTimers();
    try {
      const now = Date.now();
      fetchMock.mockResolvedValue(
        jsonResponse(200, {
          preview_base_url: "https://d/preview/pod1/",
          session_url: "https://d/preview/pod1/__session?token=JWT",
          expires_at: new Date(now + 5 * 60_000).toISOString(),
        }),
      );

      renderHook(() => usePodPreview("pod1"), { wrapper: Wrap });

      await vi.waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));

      await vi.advanceTimersByTimeAsync(5 * 60_000);

      await vi.waitFor(() => expect(fetchMock.mock.calls.length).toBeGreaterThanOrEqual(2));
    } finally {
      vi.useRealTimers();
    }
  });
});
