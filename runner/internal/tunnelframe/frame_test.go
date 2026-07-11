package tunnelframe

import "testing"

func TestHelloAckFrameType(t *testing.T) {
	if TypeHelloAck != 0x04 {
		t.Fatalf("TypeHelloAck = 0x%02x, want 0x04", TypeHelloAck)
	}
}
