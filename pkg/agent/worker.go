package agent

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
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

func (w *Worker) Start(ctx context.Context, eq <-chan *Event) {
	for event := range eq {
		w.process(ctx, event)
	}
}

func (w *Worker) process(ctx context.Context, event *Event) {
	switch event.Operation {
	case Pull:
		w.log.V(10).Info("pulling image", "image", event.Image)
		err := w.pull(ctx, event)
		if err != nil {
			w.log.Error(err, "failed to pull image", "image", event.Image)
		}
	case Remove:
		w.log.V(10).Info("removing image", "image", event.Image)
		err := w.remove(ctx, event)
		if err != nil {
			w.log.Error(err, "failed to remove image", "image", event.Image)
		}
	}
}

// pull will fetch the image from the registry.
// TODO: clean this up.  It's a bit of a mess.
func (w *Worker) pull(ctx context.Context, event *Event) error {
	if len(event.Auth) == 0 {
		_, err := w.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
		})
		if err != nil {
			w.log.Error(err, "failed to pull image", "image", event.Image)
		}
		return w.WaitImage(ctx, event.Image, true)
	}

	// TODO: I'm not sure what the behavior will be here.  I'm assuming that it will
	// send back an error if the auth fails, but since it's a non-blocking call,
	// I'm unsure how it will be handled.

	// Check the cache first.  If the creds work, we are good, if not, we'll need to
	// try all the creds that have been passed.
	if auth, ok := w.authCache[event.Image]; ok {
		w.log.V(10).Info("image pulled with existing auth reauthenticating", "image", event.Image)
		_, err := w.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
			Auth: auth,
		})
		if err == nil {
			return w.WaitImage(ctx, event.Image, true)
		} else {
			w.log.V(10).Info("failed to pull image with existing auth", "image", event.Image)
			delete(w.authCache, event.Image)
		}
	}

	w.log.V(8).Info("reauthenticating", "image", event.Image, "auths", event.Auth)

	for _, auth := range event.Auth {
		_, err := w.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
			Auth: auth,
		})
		if err != nil {
			w.log.Error(err, "unable to authenticate", "image", event.Image)
			continue
		} else {
			w.authCache[event.Image] = auth
			return w.WaitImage(ctx, event.Image, true)
		}
	}

	return nil
}

func (w *Worker) remove(ctx context.Context, event *Event) error {
	_, err := w.ims.RemoveImage(ctx, &runtime.RemoveImageRequest{
		Image: &runtime.ImageSpec{
			Image: event.Image,
		},
	})
	if err != nil {
		// Non-fatal. Ignore these errors for now to prevent the cleaner from stopping.
		w.log.Error(err, "failed to remove image", "name", event.Image)
	}

	return w.WaitImage(ctx, event.Image, false)
}

func (w *Worker) WaitImage(ctx context.Context, image string, has bool) error {
	timer := time.NewTicker(2 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			ids, err := GetImageIdentifiers(ctx, w.ims)
			if err != nil {
				w.log.Error(err, "failed to get image identifiers")
				continue
			}

			i, ok := ids[image]
			w.log.V(10).Info("waiting for image", "image", image, "has", has, "ok", ok, "id", i)
			if (ok && has) || (!ok && !has) {
				return nil
			}
		case <-time.After(5 * time.Minute):
			w.log.V(6).Info("timed out waiting for image", "image", image, "has", has)
			return nil // maybe we should return an error here?
		}
	}
}
