package protocol

import (
	"encoding/json"
	"testing"
)

func TestControlLeaseRequestRoundTrip(t *testing.T) {
	payload, err := json.Marshal(ControlLeaseRequest{
		Type:        ControlLeaseType,
		Action:      ControlLeaseActionAcquire,
		ClientLabel: "mobile",
	})
	if err != nil {
		t.Fatal(err)
	}
	request, matched, err := DecodeControlLeaseRequest(payload)
	if err != nil || !matched || request.Action != ControlLeaseActionAcquire {
		t.Fatalf("request=%+v matched=%v err=%v", request, matched, err)
	}
}

func TestDecodeControlLeaseRequestIgnoresOtherControlMessages(t *testing.T) {
	_, matched, err := DecodeControlLeaseRequest([]byte(`{"type":"pod_resized"}`))
	if err != nil || matched {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
}

func TestEncodeControlLeaseStatus(t *testing.T) {
	msg, err := DecodeMessage(EncodeControlLeaseStatus(ControlLeaseStatusGranted, "lease", 123))
	if err != nil || msg.Type != MsgTypeControl {
		t.Fatalf("message=%+v err=%v", msg, err)
	}
	var status ControlLeaseStatus
	if err := json.Unmarshal(msg.Payload, &status); err != nil {
		t.Fatal(err)
	}
	if status.Status != ControlLeaseStatusGranted || status.LeaseID != "lease" || status.ExpiresAt != 123 {
		t.Fatalf("status=%+v", status)
	}
}
