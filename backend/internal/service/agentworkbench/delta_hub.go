package agentworkbench

import (
	"errors"
	"sync"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

var ErrSubscriberLagged = errors.New("agent workbench subscriber lagged")

type DeltaSubscription struct {
	Deltas <-chan *agentworkbenchv2.SessionDeltaBatch
	Errors <-chan error
	close  func()
	once   sync.Once
}

func (subscription *DeltaSubscription) Close() {
	subscription.once.Do(subscription.close)
}

type deltaSubscriber struct {
	deltas chan *agentworkbenchv2.SessionDeltaBatch
	errors chan error
}

type DeltaHub struct {
	mu          sync.Mutex
	capacity    int
	nextID      uint64
	subscribers map[string]map[uint64]*deltaSubscriber
}

func NewDeltaHub(capacity int) *DeltaHub {
	if capacity <= 0 {
		panic("agent workbench delta hub capacity must be positive")
	}
	return &DeltaHub{
		capacity:    capacity,
		subscribers: make(map[string]map[uint64]*deltaSubscriber),
	}
}

func (hub *DeltaHub) Subscribe(sessionID string) *DeltaSubscription {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.nextID++
	id := hub.nextID
	subscriber := &deltaSubscriber{
		deltas: make(chan *agentworkbenchv2.SessionDeltaBatch, hub.capacity),
		errors: make(chan error, 1),
	}
	if hub.subscribers[sessionID] == nil {
		hub.subscribers[sessionID] = make(map[uint64]*deltaSubscriber)
	}
	hub.subscribers[sessionID][id] = subscriber
	return &DeltaSubscription{
		Deltas: subscriber.deltas,
		Errors: subscriber.errors,
		close: func() {
			hub.remove(sessionID, id, nil)
		},
	}
}

func (hub *DeltaHub) Publish(
	sessionID string,
	delta *agentworkbenchv2.SessionDeltaBatch,
) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for id, subscriber := range hub.subscribers[sessionID] {
		cloned := proto.Clone(delta).(*agentworkbenchv2.SessionDeltaBatch)
		select {
		case subscriber.deltas <- cloned:
		default:
			subscriber.errors <- ErrSubscriberLagged
			close(subscriber.errors)
			close(subscriber.deltas)
			delete(hub.subscribers[sessionID], id)
		}
	}
	if len(hub.subscribers[sessionID]) == 0 {
		delete(hub.subscribers, sessionID)
	}
}

func (hub *DeltaHub) remove(sessionID string, id uint64, failure error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	subscriber := hub.subscribers[sessionID][id]
	if subscriber == nil {
		return
	}
	if failure != nil {
		subscriber.errors <- failure
	}
	close(subscriber.errors)
	close(subscriber.deltas)
	delete(hub.subscribers[sessionID], id)
	if len(hub.subscribers[sessionID]) == 0 {
		delete(hub.subscribers, sessionID)
	}
}
