package idempotency

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

// Backend interface matches what we added to RedisStore
type Backend interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Store struct {
	backend Backend
	// In-memory fallback
	cache sync.Map
}

type entry struct {
	Resp      Response
	Timestamp time.Time
}

func NewStore(backend Backend) *Store {
	return &Store{
		backend: backend,
	}
}

func (s *Store) Get(ctx context.Context, key string) (Response, bool) {
	if s.backend != nil {
		val, err := s.backend.Get(ctx, key)
		if err != nil {
			log.Printf("Idempotency: Redis error getting %s: %v", key, err)
			return Response{}, false
		}
		if val == "" {
			return Response{}, false
		}
		var e entry
		if err := json.Unmarshal([]byte(val), &e); err != nil {
			return Response{}, false
		}
		return e.Resp, true
	}

	// Memory Fallback
	val, ok := s.cache.Load(key)
	if !ok {
		return Response{}, false
	}
	e := val.(entry)
	// Simple TTL check for memory (1 hour default)
	if time.Since(e.Timestamp) > 1*time.Hour {
		s.cache.Delete(key)
		return Response{}, false
	}
	return e.Resp, true
}

func (s *Store) Set(ctx context.Context, key string, resp Response) {
	e := entry{
		Resp:      resp,
		Timestamp: time.Now(),
	}

	if s.backend != nil {
		bytes, _ := json.Marshal(e)
		// TTL 24 hours for idempotency
		if err := s.backend.Set(ctx, key, string(bytes), 24*time.Hour); err != nil {
			log.Printf("Idempotency: Redis error setting %s: %v", key, err)
		}
		return
	}

	s.cache.Store(key, e)
}
