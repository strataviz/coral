package worker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

type Cleaner struct {
	name string
	log  logr.Logger
	ims  runtime.ImageServiceClient
	rts  runtime.RuntimeServiceClient
	cli  client.Client
	poll time.Duration

	syncOnce sync.Once
	stopChan chan struct{}
}

func NewCleaner(name string, poll time.Duration, log logr.Logger, cli client.Client, ims runtime.ImageServiceClient, rts runtime.RuntimeServiceClient) *Cleaner {
	return &Cleaner{
		name: name,
		log:  log,
		ims:  ims,
		rts:  rts,
		cli:  cli,
		poll: poll,

		stopChan: make(chan struct{}),
	}
}

func (c *Cleaner) Start(ctx context.Context) {
	timer := time.NewTicker(c.poll)
	c.log.V(8).Info("starting cleaner")
	for {
		select {
		case <-c.stopChan:
			c.log.V(8).Info("cleaner shutting down")
			return
		case <-timer.C:
			c.log.V(8).Info("cleaning up images")
			if err := c.clean(ctx); err != nil {
				c.log.Error(err, "failed to clean images")
			}
		}
	}
}

func (c *Cleaner) Stop() {
	c.syncOnce.Do(func() {
		close(c.stopChan)
	})
}

func (c *Cleaner) clean(ctx context.Context) error {
	// Get labels marked for deletion
	node := &corev1.Node{}
	if err := c.cli.Get(ctx, client.ObjectKey{Name: c.name}, node); err != nil {
		return err
	}

	// Map the images to hashes.
	mapping := make(map[string]string)
	for _, ci := range node.Status.Images {
		for _, img := range ci.Names {
			mapping[util.ImageHasher(img)] = img
		}
	}

	deleting := make([]string, 0)
	c.log.V(8).Info("parsing labels", "labels", node.Labels)
	for k, v := range node.Labels {
		// This is giving me the hash in the list.  That's a bit problematic since
		// I'm not mapping it yet.  Need to centralize that functionality.
		if strings.Contains(k, LabelPrefix) && v == string(stvziov1.ImageStateDeleting) {
			if name, ok := mapping[strings.TrimPrefix(k, LabelPrefix)]; ok {
				deleting = append(deleting, name)
			} else {
				// We can't find the image in the hashes so ignore it for now.  I need to think
				// through the situations where this could occur.
				c.log.V(10).Info("failed to find image in hash map", "key", k, "value", v)
			}
		}
	}

	c.log.V(8).Info("images staged for deletion", "images", deleting)

	for _, img := range deleting {
		c.log.V(8).Info("removing image", "image", img)
		// There's a possibility that we can refuse to fail if the image is being actively
		// used by containers on the node.  We should expose some metrics about the number of
		// active/started containers using it and then give the option to force the removal.
		// It doesn't affect containers that are already running, so it's just a convienence
		// thing to prevent unanticipated removals if the image is still needed.  This of
		// of course would cause the finalizers to hang until timeout on resource deletion.

		// This is non-blocking.  It doesn't appear to have any negative impacts as we are
		// using the k8s cri api (which uses the underlying cri api) which was most likely
		// designed to handle the situation where multiple operations could come in at once.
		// Unless we see issues, we should be fine, but if watching the logs, you may see
		// the same removal attempt multiple times since we trigger label removal on the
		// absence of the container on the node.  If it is an issue we could block until
		// the specific label has been removed.
		_, err := c.ims.RemoveImage(ctx, &runtime.RemoveImageRequest{
			Image: &runtime.ImageSpec{
				Image: img,
			},
		})
		if err != nil {
			// Non-fatal. Ignore these errors for now to prevent the cleaner from stopping.
			c.log.Error(err, "failed to remove image", "image", img)
		}
	}

	return nil
}
