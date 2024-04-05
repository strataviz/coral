package agent

import "sync"

type Semaphore struct {
	s map[string]bool
	sync.Mutex
}

func NewSemaphore() *Semaphore {
	return &Semaphore{
		s: make(map[string]bool),
	}
}

func (s *Semaphore) Acquire(key string) bool {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.s[key]; ok {
		return false
	}

	s.s[key] = true
	return true
}

func (s *Semaphore) Release(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.s, key)
}

func (s *Semaphore) Acquired(key string) bool {
	s.Lock()
	defer s.Unlock()

	_, ok := s.s[key]
	return ok
}
