"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import {
  approveExpertMarketRelease,
  getExpertMarketRelease,
  listExpertMarketReleases,
  rejectExpertMarketRelease,
  type ExpertMarketRelease,
  type ExpertMarketReleaseStatus,
} from "@/lib/api/admin";
import { errorMessage } from "./expert-market-errors";

type LoadResult = string | null | undefined;

export function useExpertMarketReview() {
  const [status, setStatus] = useState<ExpertMarketReleaseStatus>("pending");
  const [releases, setReleases] = useState<ExpertMarketRelease[]>([]);
  const [selected, setSelected] = useState<ExpertMarketRelease | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [isActing, setIsActing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isRejecting, setIsRejecting] = useState(false);
  const [rejectionReason, setRejectionReason] = useState("");
  const [rejectionError, setRejectionError] = useState("");
  const listRequestId = useRef(0);
  const detailRequestId = useRef(0);
  const activeReleaseId = useRef<number | null>(null);

  const loadReleases = useCallback(async (): Promise<LoadResult> => {
    const requestId = ++listRequestId.current;
    setIsLoading(true);
    setError(null);
    try {
      const result = await listExpertMarketReleases({
        status,
        limit: 50,
        offset: 0,
      });
      if (requestId !== listRequestId.current) return undefined;
      setReleases(result.items);
      return null;
    } catch (loadError) {
      if (requestId !== listRequestId.current) return undefined;
      const message = errorMessage(loadError);
      setError(message);
      return message;
    } finally {
      if (requestId === listRequestId.current) setIsLoading(false);
    }
  }, [status]);

  useEffect(() => {
    void loadReleases();
  }, [loadReleases]);

  const resetReviewForm = () => {
    setIsRejecting(false);
    setRejectionReason("");
    setRejectionError("");
  };

  const changeStatus = (nextStatus: ExpertMarketReleaseStatus) => {
    if (nextStatus !== status) {
      listRequestId.current += 1;
      setStatus(nextStatus);
    }
    detailRequestId.current += 1;
    activeReleaseId.current = null;
    setSelected(null);
    setIsDetailLoading(false);
    resetReviewForm();
  };

  const viewRelease = async (releaseId: number) => {
    const requestId = ++detailRequestId.current;
    activeReleaseId.current = releaseId;
    setIsDetailLoading(true);
    setSelected(null);
    resetReviewForm();
    try {
      const release = await getExpertMarketRelease(releaseId);
      if (
        requestId === detailRequestId.current &&
        activeReleaseId.current === releaseId
      ) {
        setSelected(release);
      }
    } catch (detailError) {
      if (
        requestId === detailRequestId.current &&
        activeReleaseId.current === releaseId
      ) {
        toast.error(errorMessage(detailError));
      }
    } finally {
      if (requestId === detailRequestId.current) setIsDetailLoading(false);
    }
  };

  const approve = async () => {
    if (!selected || activeReleaseId.current !== selected.id) return;
    const releaseId = selected.id;
    setIsActing(true);
    try {
      const approved = await approveExpertMarketRelease(releaseId);
      if (activeReleaseId.current !== releaseId) return;
      setSelected(approved);
      toast.success("专家发布已批准");
      await loadReleases();
    } catch (actionError) {
      if (activeReleaseId.current === releaseId) {
        toast.error(errorMessage(actionError));
      }
    } finally {
      setIsActing(false);
    }
  };

  const reject = async () => {
    const reason = rejectionReason.trim();
    if (!reason) {
      setRejectionError("请输入驳回理由");
      return;
    }
    if (!selected || activeReleaseId.current !== selected.id) return;
    const releaseId = selected.id;
    setIsActing(true);
    try {
      const rejected = await rejectExpertMarketRelease(releaseId, reason);
      if (activeReleaseId.current !== releaseId) return;
      setSelected(rejected);
      resetReviewForm();
      toast.success("专家发布已驳回");
      await loadReleases();
    } catch (actionError) {
      if (activeReleaseId.current === releaseId) {
        toast.error(errorMessage(actionError));
      }
    } finally {
      setIsActing(false);
    }
  };

  const refresh = async () => {
    const refreshError = await loadReleases();
    if (refreshError === undefined) return;
    if (refreshError) toast.error(refreshError);
    else toast.success("审核列表已刷新");
  };

  const changeRejectionReason = (value: string) => {
    setRejectionReason(value);
    setRejectionError("");
  };

  return {
    status,
    releases,
    selected,
    isLoading,
    isDetailLoading,
    isActing,
    error,
    isRejecting,
    rejectionReason,
    rejectionError,
    loadReleases,
    changeStatus,
    viewRelease,
    approve,
    reject,
    refresh,
    startRejecting: () => setIsRejecting(true),
    cancelRejecting: () => {
      setIsRejecting(false);
      setRejectionError("");
    },
    changeRejectionReason,
  };
}
