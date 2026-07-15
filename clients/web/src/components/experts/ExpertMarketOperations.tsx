"use client";

import { ExpertMarketSubmissionPanel } from "./ExpertMarketSubmissionPanel";
import { ExpertMarketUpgradePanel } from "./ExpertMarketUpgradePanel";

interface ExpertMarketOperationsProps {
  expertID: number;
  expertSlug: string;
  installedFromMarket: boolean;
  submissionReady: boolean;
  onUpgraded: () => void | Promise<void>;
}

export function ExpertMarketOperations({
  expertID,
  expertSlug,
  installedFromMarket,
  submissionReady,
  onUpgraded,
}: ExpertMarketOperationsProps) {
  if (installedFromMarket) {
    return (
      <ExpertMarketUpgradePanel
        expertSlug={expertSlug}
        onUpgraded={onUpgraded}
      />
    );
  }

  return (
    <ExpertMarketSubmissionPanel
      expertID={expertID}
      expertSlug={expertSlug}
      submissionReady={submissionReady}
    />
  );
}
