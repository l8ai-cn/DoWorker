package agentworkbench

import "google.golang.org/protobuf/proto"

func sameOptionalString(current, next *string) bool {
	return current == nil && next == nil ||
		current != nil && next != nil && *current == *next
}

func optionalStringEnriches(current, next *string) bool {
	return current == nil || next != nil && *current == *next
}

func optionalUint64Enriches(current, next *uint64) bool {
	return current == nil || next != nil && *current == *next
}

func messageEnriches(current, next proto.Message) bool {
	return !current.ProtoReflect().IsValid() ||
		next.ProtoReflect().IsValid() && proto.Equal(current, next)
}
