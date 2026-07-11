package channel

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/protocol"
)

func TestChannel_ControlLeaseGatesWriterMessages(t *testing.T) {
	_, publisher, first, second := controlLeaseChannel(t, 2*time.Second)

	writeBinary(t, second, protocol.EncodeInput([]byte("blocked")))
	requireControlStatus(t, second, protocol.ControlLeaseStatusRequired)

	leaseID := acquireControl(t, first)
	writeBinary(t, first, protocol.EncodeInput([]byte("allowed")))
	requireMessageType(t, publisher, protocol.MsgTypeInput)

	for _, frame := range [][]byte{
		protocol.EncodeResize(100, 30),
		protocol.EncodeMessage(protocol.MsgTypeAcpCommand, []byte(`{"type":"prompt"}`)),
	} {
		writeBinary(t, second, frame)
		requireControlStatus(t, second, protocol.ControlLeaseStatusRequired)
	}
	writeBinary(t, first, protocol.EncodeInput([]byte("still-allowed")))
	requireMessageType(t, publisher, protocol.MsgTypeInput)

	writeControl(t, first, protocol.ControlLeaseRequest{
		Type:    protocol.ControlLeaseType,
		Action:  protocol.ControlLeaseActionRelease,
		LeaseID: leaseID,
	})
	requireControlStatus(t, first, protocol.ControlLeaseStatusReleased)
}

func TestChannel_ControlLeaseIsExclusiveAndValidatesLeaseID(t *testing.T) {
	ch, _, first, second := controlLeaseChannel(t, 2*time.Second)
	firstLease := acquireControl(t, first)

	writeControl(t, second, protocol.ControlLeaseRequest{
		Type:   protocol.ControlLeaseType,
		Action: protocol.ControlLeaseActionAcquire,
	})
	requireControlStatus(t, second, protocol.ControlLeaseStatusBusy)

	writeControl(t, first, protocol.ControlLeaseRequest{
		Type:    protocol.ControlLeaseType,
		Action:  protocol.ControlLeaseActionRenew,
		LeaseID: "wrong",
	})
	requireControlStatus(t, first, protocol.ControlLeaseStatusRequired)

	writeControl(t, first, protocol.ControlLeaseRequest{
		Type:    protocol.ControlLeaseType,
		Action:  protocol.ControlLeaseActionRenew,
		LeaseID: firstLease,
	})
	status := requireControlStatus(t, first, protocol.ControlLeaseStatusGranted)
	if status.LeaseID != firstLease {
		t.Fatalf("renewed lease ID = %q, want %q", status.LeaseID, firstLease)
	}

	writeControl(t, second, protocol.ControlLeaseRequest{
		Type:    protocol.ControlLeaseType,
		Action:  protocol.ControlLeaseActionRelease,
		LeaseID: firstLease,
	})
	requireControlStatus(t, second, protocol.ControlLeaseStatusRequired)

	ch.RemoveSubscriber("first")
	requireControlStatus(t, second, protocol.ControlLeaseStatusReleased)
	if leaseID := acquireControl(t, second); leaseID == "" {
		t.Fatal("second subscriber did not acquire after owner disconnect")
	}
}

func TestChannel_ControlLeaseExpiresAndAllowsAnotherController(t *testing.T) {
	_, _, first, second := controlLeaseChannel(t, 40*time.Millisecond)
	acquireControl(t, first)
	writeControl(t, second, protocol.ControlLeaseRequest{
		Type:   protocol.ControlLeaseType,
		Action: protocol.ControlLeaseActionAcquire,
	})
	requireControlStatus(t, second, protocol.ControlLeaseStatusBusy)
	requireControlStatus(t, first, protocol.ControlLeaseStatusExpired)
	requireControlStatus(t, second, protocol.ControlLeaseStatusExpired)
	if leaseID := acquireControl(t, second); leaseID == "" {
		t.Fatal("second subscriber did not acquire expired lease")
	}
}

func TestChannel_ConcurrentControlAcquireHasSingleWinner(t *testing.T) {
	_, _, first, second := controlLeaseChannel(t, 2*time.Second)
	request := protocol.ControlLeaseRequest{
		Type:   protocol.ControlLeaseType,
		Action: protocol.ControlLeaseActionAcquire,
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	frame := protocol.EncodeMessage(protocol.MsgTypeControl, payload)

	var writes sync.WaitGroup
	writeErrors := make(chan error, 2)
	writes.Add(2)
	go func() {
		defer writes.Done()
		writeErrors <- first.WriteMessage(websocket.BinaryMessage, frame)
	}()
	go func() {
		defer writes.Done()
		writeErrors <- second.WriteMessage(websocket.BinaryMessage, frame)
	}()
	writes.Wait()
	close(writeErrors)
	for err := range writeErrors {
		if err != nil {
			t.Fatal(err)
		}
	}

	statuses := []protocol.ControlLeaseStatus{
		readControlStatus(t, first),
		readControlStatus(t, second),
	}
	granted := 0
	busy := 0
	for _, status := range statuses {
		switch status.Status {
		case protocol.ControlLeaseStatusGranted:
			granted++
		case protocol.ControlLeaseStatusBusy:
			busy++
		}
	}
	if granted != 1 || busy != 1 {
		t.Fatalf("statuses=%+v, want one granted and one busy", statuses)
	}
}

func TestChannel_ObserverStillReceivesOutputAndCanResync(t *testing.T) {
	_, publisher, _, observer := controlLeaseChannel(t, 200*time.Millisecond)

	output := protocol.EncodeOutput([]byte("visible"))
	writeBinary(t, publisher, output)
	requireMessageType(t, observer, protocol.MsgTypeOutput)

	resync := protocol.EncodeMessage(protocol.MsgTypeResync, nil)
	writeBinary(t, observer, resync)
	requireMessageType(t, publisher, protocol.MsgTypeResync)
}

func controlLeaseChannel(
	t *testing.T,
	duration time.Duration,
) (*Channel, *websocket.Conn, *websocket.Conn, *websocket.Conn) {
	t.Helper()
	cfg := testChannelConfig()
	cfg.ControlLeaseDuration = duration
	ch := NewChannelWithConfig("pod-control", cfg, nil, nil)
	pubServer, pubClient := createWSPair(t)
	firstServer, firstClient := createWSPair(t)
	secondServer, secondClient := createWSPair(t)
	ch.SetPublisher(pubServer)
	ch.AddSubscriber("first", firstServer)
	ch.AddSubscriber("second", secondServer)
	t.Cleanup(ch.Close)
	return ch, pubClient, firstClient, secondClient
}

func acquireControl(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	writeControl(t, conn, protocol.ControlLeaseRequest{
		Type:        protocol.ControlLeaseType,
		Action:      protocol.ControlLeaseActionAcquire,
		ClientLabel: "test",
	})
	status := requireControlStatus(t, conn, protocol.ControlLeaseStatusGranted)
	if status.LeaseID == "" || status.ExpiresAt == 0 {
		t.Fatalf("incomplete granted status: %+v", status)
	}
	return status.LeaseID
}

func writeControl(t *testing.T, conn *websocket.Conn, request protocol.ControlLeaseRequest) {
	t.Helper()
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	writeBinary(t, conn, protocol.EncodeMessage(protocol.MsgTypeControl, payload))
}

func writeBinary(t *testing.T, conn *websocket.Conn, data []byte) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatal(err)
	}
}

func requireControlStatus(
	t *testing.T,
	conn *websocket.Conn,
	want string,
) protocol.ControlLeaseStatus {
	t.Helper()
	status := readControlStatus(t, conn)
	if status.Status != want {
		t.Fatalf("control status = %q, want %q", status.Status, want)
	}
	return status
}

func readControlStatus(t *testing.T, conn *websocket.Conn) protocol.ControlLeaseStatus {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read control status: %v", err)
	}
	msg, err := protocol.DecodeMessage(data)
	if err != nil || msg.Type != protocol.MsgTypeControl {
		t.Fatalf("control status frame = %v, err=%v", msg, err)
	}
	var status protocol.ControlLeaseStatus
	if err := json.Unmarshal(msg.Payload, &status); err != nil {
		t.Fatal(err)
	}
	return status
}

func requireMessageType(t *testing.T, conn *websocket.Conn, want byte) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message type 0x%02x: %v", want, err)
	}
	msg, err := protocol.DecodeMessage(data)
	if err != nil || msg.Type != want {
		t.Fatalf("message = %v, err=%v, want type=0x%02x", msg, err, want)
	}
}
