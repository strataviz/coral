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

package agent

import (
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Operation int

const (
	Pull Operation = iota
	Remove
)

type Event struct {
	Image     string
	Auth      []*runtime.AuthConfig
	Operation Operation
}

type EventQueue chan *Event

func NewEventQueue() EventQueue {
	return make(chan *Event)
}

func (eq EventQueue) Close() {
	close(eq)
}
