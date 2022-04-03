package service

import (
	"sync"
	"time"
)

type store struct {
	sync.RWMutex
	data map[string]entry
}

type entry struct {
	Requests []Request
	Created  int64
}

func (s *store) get(key string) (*[]Request, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	now := time.Now().Unix()
	if len(e.Requests) > 0 && e.Requests[0].TTL > 1 && (e.Created+int64(e.Requests[0].TTL) < now) {
		s.remove(key)
		return nil, false
	}
	return &e.Requests, ok
}

func (s *store) set(key string, reqs []Request) bool {
	changed := false
	s.Lock()
	if _, ok := s.data[key]; ok {
		e := s.data[key]
		e.Requests = reqs
		s.data[key] = e
		changed = true
	} else {
		e := entry{
			Requests: reqs,
			Created:  time.Now().Unix(),
		}
		s.data[key] = e
		changed = true
	}
	s.Unlock()
	return changed
}

func (s *store) remove(key string) bool {
	s.Lock()
	_, ok := s.data[key]
	delete(s.data, key)
	s.Unlock()
	return ok
}
