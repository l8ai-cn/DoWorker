"use client";

import { useCallback, useState } from "react";

const PAGE_LIMIT = 50;

interface PaginationResponse {
  total: number;
  limit: number;
  offset: number;
}

export function useExpertMarketPagination() {
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(PAGE_LIMIT);
  const [offset, setOffset] = useState(0);

  const applyResponse = useCallback((response: PaginationResponse) => {
    setTotal(response.total);
    setLimit(response.limit);
    setOffset(response.offset);
  }, []);
  const reset = useCallback(() => setOffset(0), []);
  const previousPage = useCallback(() => {
    setOffset((current) => Math.max(0, current - limit));
  }, [limit]);
  const nextPage = useCallback(() => {
    setOffset((current) =>
      current + limit < total ? current + limit : current,
    );
  }, [limit, total]);

  return {
    total,
    limit,
    offset,
    requestLimit: PAGE_LIMIT,
    applyResponse,
    reset,
    previousPage,
    nextPage,
  };
}
