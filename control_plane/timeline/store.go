package timeline

import (
	"sync"
	"time"
)

type ReconcileEvent struct {
	ReqID     string            `json:"req_id"`
	Stage     string            `json:"stage"` // CREATED, QUEUED, SCHEDULED, WORKER_ASSIGNED, EXEC_STARTED, EXEC_FINISHED, STATE_COMMITTED, FAILED
	Timestamp time.Time         `json:"timestamp"`
	NodeID    string            `json:"node_id"`
	TenantID  string            `json:"tenant_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Store struct {
	events []ReconcileEvent
	mu     sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		events: make([]ReconcileEvent, 0),
	}
}

func (s *Store) Record(e ReconcileEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	s.events = append(s.events, e)
}

func (s *Store) GetEvents(reqID string) []ReconcileEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []ReconcileEvent
	for _, e := range s.events {
		if e.ReqID == reqID {
			results = append(results, e)
		}
	}
	return results
}

func (s *Store) GetEventsByStateID(stateID string) []ReconcileEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []ReconcileEvent
	for _, e := range s.events {
		if e.Metadata != nil && e.Metadata["state_id"] == stateID {
			results = append(results, e)
		}
	}
	return results
}

// GetAllEvents returns specific range of events (simple implementation for debug snapshot)
func (s *Store) GetAllEvents() []ReconcileEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return copy
	c := make([]ReconcileEvent, len(s.events))
	copy(c, s.events)
	return c
}
