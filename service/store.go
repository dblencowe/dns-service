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
	Request request
	Created int64
}

func (s *store) get(key string) (*request, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	now := time.Now().Unix()
	if e.Request.TTL > 1 && (e.Created+int64(e.Request.TTL) < now) {
		s.remove(key)
		return nil, false
	}
	return &e.Request, ok
}

func (s *store) set(key string, req request) bool {
	changed := false
	s.Lock()
	if _, ok := s.data[key]; ok {
		e := s.data[key]
		e.Request = req
		s.data[key] = e
		changed = true
	} else {
		e := entry{
			Request: req,
			Created: time.Now().Unix(),
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
