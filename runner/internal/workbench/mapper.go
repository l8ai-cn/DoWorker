package workbench

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/google/uuid"
)

type timelineState struct {
	itemID string
	text   string
}

type Mapper struct {
	mu              sync.Mutex
	podKey          string
	adapterID       string
	sourceProtocol  string
	epoch           string
	externalSession string
	sequence        uint64
	itemSequence    uint64
	messages        map[string]*timelineState
	thinking        *timelineState
	planSeen        bool
	tools           map[string]*agentworkbenchv2.ToolExecution
}

func NewMapper(podKey, adapterID string) *Mapper {
	return &Mapper{
		podKey:         podKey,
		adapterID:      adapterID,
		sourceProtocol: sourceProtocol(adapterID),
		epoch:          uuid.NewString(),
		messages:       make(map[string]*timelineState),
		tools:          make(map[string]*agentworkbenchv2.ToolExecution),
	}
}

func (m *Mapper) SetExternalSessionID(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sessionID != "" {
		m.externalSession = sessionID
	}
}

func (m *Mapper) Unsupported(
	semanticKey string,
	source any,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.batchLocked(
		source,
		m.unsupportedMutationLocked(semanticKey, stringPayload(source)),
	)
}

func (m *Mapper) batchLocked(
	source any,
	mutations ...*agentworkbenchv2.RunnerWorkbenchMutation,
) *agentworkbenchv2.RunnerWorkbenchEventBatch {
	payload := structuredPayload("application/json", source)
	for _, mutation := range mutations {
		m.sequence++
		mutation.Source = &agentworkbenchv2.RunnerSourceEvent{
			StableEventId:  fmt.Sprintf("%s:%d", m.epoch, m.sequence),
			SourceSequence: m.sequence,
			OccurredAt:     time.Now().UTC().Format(time.RFC3339Nano),
			SourcePayload:  payload,
		}
	}
	batch := &agentworkbenchv2.RunnerWorkbenchEventBatch{
		PodKey:                m.podKey,
		AdapterId:             m.adapterID,
		SourceProtocolVersion: "agentcloud-normalized-acp/1",
		RunnerSessionEpoch:    m.epoch,
		Mutations:             mutations,
	}
	if m.externalSession != "" {
		batch.ExternalSessionId = stringPointer(m.externalSession)
	}
	return batch
}

func (m *Mapper) nextItemIDLocked(kind string) string {
	m.itemSequence++
	return fmt.Sprintf("%s:%s:%d", m.podKey, kind, m.itemSequence)
}

func structuredPayload(mediaType string, value any) *agentworkbenchv2.StructuredPayload {
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"marshal_error":%q}`, err.Error()))
	}
	return &agentworkbenchv2.StructuredPayload{MediaType: mediaType, Data: data}
}

func rawPayload(mediaType, value string) *agentworkbenchv2.StructuredPayload {
	return &agentworkbenchv2.StructuredPayload{
		MediaType: mediaType,
		Data:      []byte(value),
	}
}

func stringPointer(value string) *string {
	return &value
}
