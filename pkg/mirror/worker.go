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
	"context"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Worker struct {
	id  int
	log logr.Logger

	authCache map[string]*runtime.AuthConfig
}

func NewWorker(id int) *Worker {
	return &Worker{
		id:        id,
		log:       logr.Discard(),
		authCache: make(map[string]*runtime.AuthConfig),
	}
}

func (w *Worker) WithLogger(log logr.Logger) *Worker {
	w.log = log.WithName("mirror-worker").WithValues("worker", w.id)
	return w
}

func (w *Worker) Start(ctx context.Context, wq WorkQueue, sem *Semaphore) {
	for item := range wq {
		w.process(ctx, item, sem)
	}
}

func (w *Worker) process(ctx context.Context, item *Item, sem *Semaphore) {
	// Make sure we only have one worker operating on an image at a time.
	do := sem.Acquire(item.Image)
	defer sem.Release(item.Image)

	if !do {
		w.log.V(10).Info("failed to acquire semaphore, skipping", "image", item.Image)
		return
	}

	// Sync the image.
	w.log.V(4).Info("syncing image", "image", item.Image)
	err := w.sync(ctx, item)
	if err != nil {
		w.log.Error(err, "failed to sync image", "image", item.Image)
	}
}

func (w *Worker) sync(ctx context.Context, item *Item) error {
	if len(item.Auth) == 0 {
		w.log.V(6).Info("attempting to sync image without credentials", "image", item.Image, "registry", item.Registry)
		return Copy(ctx, nil, item.Image, item.Registry)
	}

	if auth, ok := w.authCache[item.Image]; ok {
		// TODO: convert auth.
		err := Copy(ctx, auth, item.Image, item.Registry)
		// TODO: differentiate between auth errors and other errors.
		if err != nil {
			w.log.V(8).Error(err, "failed to pull image with cached credentials", "image", item.Image)
			delete(w.authCache, item.Image)
		}
	}

	for _, auth := range item.Auth {
		w.log.V(4).Info("attempting to pull image with provided credentials", "image", item.Image, "username", auth.Username)
		// TODO: convert auth.
		err := Copy(ctx, auth, item.Image, item.Registry)
		if err != nil {
			continue
		} else {
			w.authCache[item.Image] = auth
			return nil
		}
	}

	return nil
}
