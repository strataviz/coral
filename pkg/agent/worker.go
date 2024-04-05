package agent

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type WorkerError string

func (e WorkerError) Error() string { return string(e) }

const (
	ErrImageNotFound WorkerError = "image not found"
	ErrTimeout       WorkerError = "timeout"
	WaitTimeout      WorkerError = "image wait timeout"

	DefaultWaitPollInterval = 2 * time.Second
	DefaultWaitTimeout      = 5 * time.Minute
)

type Worker struct {
	log logr.Logger
	ims runtime.ImageServiceClient
	rts runtime.RuntimeServiceClient

	authCache map[string]*runtime.AuthConfig
}

func NewWorker(id int, options *AgentOptions) *Worker {
	return &Worker{
		log:       options.Log.WithValues("worker", id),
		ims:       options.ImageServiceClient,
		rts:       options.RuntimeServiceClient,
		authCache: make(map[string]*runtime.AuthConfig),
	}
}

func (w *Worker) Start(ctx context.Context, eq <-chan *Event, sem *Semaphore) {
	for event := range eq {
		w.process(ctx, event, sem)
	}
}

func (w *Worker) process(ctx context.Context, event *Event, sem *Semaphore) {
	// Make sure we only have one worker operating on an image at a time.
	do := sem.Acquire(event.Image)
	defer sem.Release(event.Image)

	if !do {
		w.log.V(10).Info("failed to acquire semaphore, skipping", "image", event.Image)
		return
	}

	// nolint: gocritic
	switch event.Operation {
	case Pull:
		w.log.V(10).Info("pulling image", "image", event.Image)
		err := w.pull(ctx, event)
		if err != nil {
			w.log.Error(err, "failed to pull image", "image", event.Image)
		}
		w.log.V(8).Info("image pulled", "image", event.Image)
	}
}

// pull will fetch the image from the registry. If the image requires authentication,
// it will attempt to authenticate with the first valid set of credentials.  Once
// successfully authenticated it will cache the credentials for future use.  If the
// cached credentials fail to authenticate, they will be removed from the cache and
// it will attempt to authenticate with the provided credentials.
func (w *Worker) pull(ctx context.Context, event *Event) error {
	if len(event.Auth) == 0 {
		w.log.V(4).Info("attempting to pull image without credentials", "image", event.Image)
		return w.pullImage(ctx, event.Image, nil)
	}

	if auth, ok := w.authCache[event.Image]; ok {
		err := w.pullImage(ctx, event.Image, auth)
		// TODO: differentiate between auth errors and other errors.
		if err != nil {
			w.log.V(8).Error(err, "failed to pull image with cached credentials", "image", event.Image)
			delete(w.authCache, event.Image)
		}
	}

	for _, auth := range event.Auth {
		w.log.V(4).Info("attempting to pull image with provided credentials", "image", event.Image, "username", auth.Username)
		err := w.pullImage(ctx, event.Image, auth)
		if err != nil {
			continue
		} else {
			w.authCache[event.Image] = auth
			return nil
		}
	}

	return nil
}

func (w *Worker) pullImage(ctx context.Context, image string, auth *runtime.AuthConfig) error {
	_, err := w.ims.PullImage(ctx, &runtime.PullImageRequest{
		Image: &runtime.ImageSpec{
			Image: image,
		},
		Auth: auth,
	})
	if err != nil {
		return err
	}

	return w.waitImage(ctx, image, true)
}

func (w *Worker) waitImage(ctx context.Context, image string, has bool) error {
	timer := time.NewTicker(DefaultWaitPollInterval)
	defer timer.Stop()

	for timeout := time.After(DefaultWaitTimeout); ; {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			ids, err := GetImageIdentifiers(ctx, w.ims)
			if err != nil {
				w.log.Error(err, "failed to get image identifiers")
				continue
			}

			_, ok := ids[image]
			if (ok && has) || (!ok && !has) {
				return nil
			}
		case <-timeout:
			w.log.V(6).Info("timed out waiting for image", "image", image, "has", has)
			return WaitTimeout
		}
	}
}
