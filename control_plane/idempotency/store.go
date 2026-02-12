package idempotency

import (
	"sync"
	"time"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

type Store struct {
	cache sync.Map
}

type entry struct {
	resp      Response
	timestamp time.Time
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Get(key string) (Response, bool) {
	val, ok := s.cache.Load(key)
	if !ok {
		return Response{}, false
	}
	e := val.(entry)
	// Simple TTL: 1 hour
	if time.Since(e.timestamp) > 1*time.Hour {
		s.cache.Delete(key)
		return Response{}, false
	}
	return e.resp, true
}

func (s *Store) Set(key string, resp Response) {
	s.cache.Store(key, entry{
		resp:      resp,
		timestamp: time.Now(),
	})
}
