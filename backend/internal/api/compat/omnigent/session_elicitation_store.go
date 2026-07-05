package omnigent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

type ElicitationRecord struct {
	ID        string
	RequestID string
	Message   string
	Phase     string
	Status    string
	RawEvent  map[string]any
}

type ElicitationStore struct {
	mu       sync.RWMutex
	bySession map[string]map[string]*ElicitationRecord
}

func NewElicitationStore() *ElicitationStore {
	return &ElicitationStore{bySession: make(map[string]map[string]*ElicitationRecord)}
}

func NewElicitID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("elicit id: %w", err)
	}
	return "elicit_" + hex.EncodeToString(b), nil
}

func (s *ElicitationStore) Add(sessionID string, rec *ElicitationRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bySession[sessionID] == nil {
		s.bySession[sessionID] = make(map[string]*ElicitationRecord)
	}
	s.bySession[sessionID][rec.ID] = rec
}

func (s *ElicitationStore) Get(sessionID, elicitID string) (*ElicitationRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.bySession[sessionID]; ok {
		rec, found := m[elicitID]
		return rec, found
	}
	return nil, false
}

func (s *ElicitationStore) Resolve(sessionID, elicitID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.bySession[sessionID][elicitID]
	if !ok {
		return false
	}
	rec.Status = "resolved"
	return true
}

func (s *ElicitationStore) PendingPayloads(sessionID string) []map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m := s.bySession[sessionID]
	if len(m) == 0 {
		return nil
	}
	out := make([]map[string]any, 0)
	for _, rec := range m {
		if rec.Status == "pending" && rec.RawEvent != nil {
			out = append(out, rec.RawEvent)
		}
	}
	return out
}

func (s *ElicitationStore) PendingCount(sessionID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, rec := range s.bySession[sessionID] {
		if rec.Status == "pending" {
			n++
		}
	}
	return n
}
