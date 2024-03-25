package worker

import (
	"context"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Puller struct {
	id  int
	log logr.Logger
	ims runtime.ImageServiceClient
	rts runtime.RuntimeServiceClient
}

func NewPuller(id int, log logr.Logger, ims runtime.ImageServiceClient, rts runtime.RuntimeServiceClient) *Puller {
	log.V(8).Info("initializing puller", "id", id)
	return &Puller{
		id:  id,
		log: log,
		ims: ims,
		rts: rts,
	}
}

func (p *Puller) Start(ctx context.Context, wq <-chan string) {
	p.log.V(8).Info("starting puller", "id", p.id)
	for img := range wq {
		p.log.V(8).Info("pulling image", "image", img)
		if err := p.pull(ctx, img); err != nil {
			p.log.Error(err, "failed to pull image", "image", img)
		}
	}
}

// pull will fetch the image from the registry.
func (p *Puller) pull(ctx context.Context, image string) error {
	_, err := p.ims.PullImage(ctx, &runtime.PullImageRequest{
		Image: &runtime.ImageSpec{
			Image: image,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
