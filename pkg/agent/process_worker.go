package agent

import (
	"context"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type ProcessWorker struct {
	log logr.Logger
	ims runtime.ImageServiceClient
	rts runtime.RuntimeServiceClient

	authCache map[string]*runtime.AuthConfig
}

func NewProcessWorker(id int, options *AgentOptions) *ProcessWorker {
	return &ProcessWorker{
		log:       options.Log.WithValues("worker", id),
		ims:       options.ImageServiceClient,
		rts:       options.RuntimeServiceClient,
		authCache: make(map[string]*runtime.AuthConfig),
	}
}

func (p *ProcessWorker) Start(ctx context.Context, eq <-chan *Event) {
	for event := range eq {
		p.process(ctx, event)
	}
}

func (p *ProcessWorker) process(ctx context.Context, event *Event) {
	switch event.Operation {
	case Pull:
		p.log.V(10).Info("pulling image", "image", event.Image)
		err := p.pull(ctx, event)
		if err != nil {
			p.log.Error(err, "failed to pull image", "image", event.Image)
		}
	case Remove:
		p.log.V(10).Info("removing image", "image", event.Image)
		err := p.remove(ctx, event)
		if err != nil {
			p.log.Error(err, "failed to remove image", "image", event.Image)
		}
	}
}

// pull will fetch the image from the registry.
func (p *ProcessWorker) pull(ctx context.Context, event *Event) error {
	if len(event.Auth) == 0 {
		_, err := p.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
		})
		return err
	}

	// TODO: I'm not sure what the behavior will be here.  I'm assuming that it will
	// send back an error if the auth fails, but since it's a non-blocking call,
	// I'm unsure how it will be handled.

	// Check the cache first.  If the creds work, we are good, if not, we'll need to
	// try all the creds that have been passed.
	if auth, ok := p.authCache[event.Image]; ok {
		p.log.V(10).Info("image pulled with existing auth reauthenticating", "image", event.Image)
		_, err := p.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
			Auth: auth,
		})
		if err == nil {
			return nil
		} else {
			p.log.V(10).Info("failed to pull image with existing auth", "image", event.Image)
			delete(p.authCache, event.Image)
		}
	}

	p.log.V(8).Info("reauthenticating", "image", event.Image, "auths", event.Auth)

	for _, auth := range event.Auth {
		_, err := p.ims.PullImage(ctx, &runtime.PullImageRequest{
			Image: &runtime.ImageSpec{
				Image: event.Image,
			},
			Auth: auth,
		})
		if err != nil {
			p.log.Error(err, "unable to authenticate", "image", event.Image)
			continue
		} else {
			p.authCache[event.Image] = auth
			return nil
		}
	}

	return nil
}

func (p *ProcessWorker) remove(ctx context.Context, event *Event) error {
	_, err := p.ims.RemoveImage(ctx, &runtime.RemoveImageRequest{
		Image: &runtime.ImageSpec{
			Image: event.Image,
		},
	})
	if err != nil {
		// Non-fatal. Ignore these errors for now to prevent the cleaner from stopping.
		p.log.Error(err, "failed to remove image", "name", event.Image)
	}

	return nil
}
