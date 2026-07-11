package protocol

import (
	"encoding/json"
	"errors"
)

const (
	ControlLeaseType = "control_lease"

	ControlLeaseActionAcquire = "acquire"
	ControlLeaseActionRenew   = "renew"
	ControlLeaseActionRelease = "release"

	ControlLeaseStatusGranted  = "granted"
	ControlLeaseStatusBusy     = "busy"
	ControlLeaseStatusReleased = "released"
	ControlLeaseStatusExpired  = "expired"
	ControlLeaseStatusRequired = "control_required"
)

var ErrInvalidControlLease = errors.New("invalid control lease message")

type ControlLeaseRequest struct {
	Type        string `json:"type"`
	Action      string `json:"action"`
	LeaseID     string `json:"lease_id,omitempty"`
	ClientLabel string `json:"client_label,omitempty"`
}

type ControlLeaseStatus struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	LeaseID   string `json:"lease_id,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

func DecodeControlLeaseRequest(payload []byte) (ControlLeaseRequest, bool, error) {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return ControlLeaseRequest{}, false, nil
	}
	if envelope.Type != ControlLeaseType {
		return ControlLeaseRequest{}, false, nil
	}

	var request ControlLeaseRequest
	if err := json.Unmarshal(payload, &request); err != nil || request.Action == "" {
		return ControlLeaseRequest{}, true, ErrInvalidControlLease
	}
	return request, true, nil
}

func EncodeControlLeaseStatus(status, leaseID string, expiresAt int64) []byte {
	payload, _ := json.Marshal(ControlLeaseStatus{
		Type:      ControlLeaseType,
		Status:    status,
		LeaseID:   leaseID,
		ExpiresAt: expiresAt,
	})
	return EncodeMessage(MsgTypeControl, payload)
}
