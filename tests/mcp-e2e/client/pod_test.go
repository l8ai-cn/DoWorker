package client

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodePodWireRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		wire string
		want string
	}{
		{
			name: "malformed pod id",
			wire: `{"id":"pod","runnerId":"1"}`,
			want: `decode pod id "pod"`,
		},
		{
			name: "zero pod id",
			wire: `{"id":"0","runnerId":"1"}`,
			want: `decode pod id "0": must be positive`,
		},
		{
			name: "malformed runner id",
			wire: `{"id":"1","runnerId":"runner"}`,
			want: `decode pod runnerId "runner"`,
		},
		{
			name: "missing runner id",
			wire: `{"id":"1"}`,
			want: `decode pod runnerId ""`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodePodWire(json.RawMessage(tt.wire))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("decodePodWire() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestDecodePodWireAcceptsPositiveIdentifiers(t *testing.T) {
	pod, err := decodePodWire(json.RawMessage(
		`{"id":"7","runnerId":"9","podKey":"7-standalone-a1"}`,
	))
	if err != nil {
		t.Fatalf("decodePodWire() error = %v", err)
	}
	if pod.ID != 7 || pod.RunnerID != 9 {
		t.Fatalf("decodePodWire() = %+v", pod)
	}
}
