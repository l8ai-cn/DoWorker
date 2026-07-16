interface TerminalLeaseRenewalInput {
  expiresAt: number;
  leaseId: string;
  now?: () => number;
  onError?: (error: unknown) => void;
  renew: (leaseId: string) => Promise<void>;
}

export function startTerminalLeaseRenewal({
  expiresAt,
  leaseId,
  now = Date.now,
  onError = () => undefined,
  renew,
}: TerminalLeaseRenewalInput): () => void {
  const interval = Math.max(expiresAt - now() - 10_000, 10_000);
  const timer = window.setInterval(() => {
    void renew(leaseId).catch(onError);
  }, interval);
  return () => window.clearInterval(timer);
}
