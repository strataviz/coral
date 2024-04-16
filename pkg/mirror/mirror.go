package mirror

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/log"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	informer "stvz.io/coral/pkg/informer/mirror"
	"stvz.io/coral/pkg/util"
)

// The controller will create a new mirror deployment.  Currently we mix the mirror and
// the image list, but in the future we should split those out to make it easy to reuse
// and control access.  For right now, I'm just going to keep them together.

type Options struct {
	Namespace string
	Name      string
	Scope     string
	Labels    labels.Selector
	Informer  *informer.Informer
}

type Mirror struct {
	namespace string
	scope     string
	labels    labels.Selector
	name      string
	informer  *informer.Informer
	log       logr.Logger
}

func New(opts *Options) *Mirror {
	return &Mirror{
		name:      opts.Name,
		labels:    opts.Labels,
		scope:     opts.Scope,
		namespace: opts.Namespace,
		informer:  opts.Informer,
	}
}

func (m *Mirror) Start(ctx context.Context) error {
	m.log = log.FromContext(ctx)
	var wg sync.WaitGroup
	sem := NewSemaphore()
	wq := NewWorkQueue()

	// TODO: I don't think I actually want workers because we want to be able to use
	// deployments to scale the mirror.
	for i := 0; i < 1; i++ {
		wg.Add(1)
		worker := NewWorker(i, m.informer.Keyring)
		go func(worker *Worker) {
			defer wg.Done()
			worker.Start(ctx, wq, sem)
		}(worker)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("stopping agent")
			close(wq)
			wg.Wait()
			return nil
		case <-ticker.C:
			m.process(ctx, wq, sem)
		}
	}
}

// TODO: Refactor for simplicity.
func (m *Mirror) process(ctx context.Context, wq WorkQueue, sem *Semaphore) { //nolint:gocognit
	for _, mirror := range m.informer.Mirrors {
		log := m.log.WithValues("mirror", mirror.Name)

		registry := mirror.Spec.Registry

		for _, repo := range mirror.Spec.Repositories {
			log := log.WithValues("repo", *repo.Name, "registry", registry) //nolint:govet
			log.V(8).Info("processing repo")

			// Normalize repo name without tags.
			nm, err := stvziov1.NormalizeRepoTag(*repo.Name, "")
			if err != nil {
				log.Error(err, "failed to create explicit repo name")
				continue
			}

			tags, err := GetRepositoryTags(ctx, nil, registry.URL(), nm)
			if err != nil {
				log.Error(err, "failed to list tags")
				continue
			}

			missing := util.ListDiff(repo.Tags, tags)
			log.V(8).Info("processing tags", "missing", missing, "found", tags)

			for _, tag := range missing {
				// TODO: We can use the normalized repo name here.
				normalized, err := stvziov1.NormalizeRepoTag(*repo.Name, tag)
				if err != nil {
					log.Error(err, "failed to create explicit tag", "tag", tag)
					continue
				}
				// TODO: checksum normalized to prevent hotspots in the ring.
				if m.informer.ServerRing.Mine(m.name, normalized) && !sem.Acquired(normalized) {
					log.V(4).Info("queueing image", "image", normalized)
					wq <- &Item{
						Registry: registry.URL(),
						Image:    normalized,
					}
				} else {
					log.V(8).Info("skipping image", "image", normalized)
				}
			}
		}
	}
}
