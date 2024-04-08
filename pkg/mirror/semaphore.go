// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mirror

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
