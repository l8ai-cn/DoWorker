package tunnelframe

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	f := Frame{Type: TypeReqBody, StreamID: 0x01020304, Payload: []byte("hello")}
	raw := Encode(f)
	if len(raw) != 1+4+5 {
		t.Fatalf("bad length %d", len(raw))
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeReqBody || got.StreamID != 0x01020304 || !bytes.Equal(got.Payload, []byte("hello")) {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestDecode_TooShort(t *testing.T) {
	if _, err := Decode([]byte{0x10, 0x00}); err == nil {
		t.Fatal("expected error on short frame")
	}
}

func TestHelloAckFrameType(t *testing.T) {
	if TypeHelloAck != 0x04 {
		t.Fatalf("TypeHelloAck = 0x%02x, want 0x04", TypeHelloAck)
	}
}
