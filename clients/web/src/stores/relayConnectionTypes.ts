export type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error";

export interface ConnectionHandle {
  send: (data: string) => void;
  unsubscribe: () => void;
}

export type RelayStatusInfo = {
  status: ConnectionStatus | "none";
  runnerDisconnected: boolean;
  controlLease: ControlLeaseInfo;
};

export type ControlLeaseStatus =
  | "observer"
  | "granted"
  | "busy"
  | "released"
  | "expired"
  | "control_required";

export type ControlLeaseInfo = {
  status: ControlLeaseStatus;
  leaseId?: string;
  expiresAt?: number;
};

export type StatusListener = (info: RelayStatusInfo) => void;
