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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"stvz.io/coral/pkg/credentials"
)

type Worker struct {
	id      int
	log     logr.Logger
	keyring *credentials.Keyring
}

func NewWorker(id int, keyring *credentials.Keyring) *Worker {
	return &Worker{
		id:      id,
		keyring: keyring,
	}
}

func (w *Worker) Start(ctx context.Context, wq WorkQueue, sem *Semaphore) {
	w.log = log.FromContext(ctx)

	w.log.V(8).Info("starting worker", "id", w.id)
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
	auth, found, err := w.keyring.Lookup(ctx, item.Image)
	if err != nil {
		return err
	}

	if !found {
		w.log.V(6).Info("attempting to sync image without credentials", "image", item.Image, "registry", item.Registry)
		return Copy(ctx, nil, item.Registry, item.Image)
	}

	for _, a := range auth {
		w.log.V(4).Info("attempting to pull image with provided credentials", "image", item.Image, "username", a.Username)
		// TODO: convert auth.
		err := Copy(ctx, a, item.Registry, item.Image)
		if err != nil {
			continue
		}
	}

	return nil
}
