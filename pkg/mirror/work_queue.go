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

import (
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Item struct {
	Image    string
	Registry string
	Auth     []*runtime.AuthConfig
}

type WorkQueue chan *Item

func NewWorkQueue() WorkQueue {
	return make(chan *Item)
}

func (wq WorkQueue) Close() {
	close(wq)
}
