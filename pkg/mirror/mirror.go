package mirror

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
	"stvz.io/hashring"
)

// The controller will create a new mirror deployment.  Currently we mix the mirror and
// the image list, but in the future we should split those out to make it easy to reuse
// and control access.  For right now, I'm just going to keep them together.

type Options struct {
	Log         logr.Logger
	Namespace   string
	Name        string
	Scope       string
	Labels      labels.Selector
	Ring        *hashring.Ring
	MirrorCache *MirrorCache
}

type Mirror struct {
	log         logr.Logger
	namespace   string
	scope       string
	labels      labels.Selector
	name        string
	ring        *hashring.Ring
	mirrorCache *MirrorCache
	authCache   *AuthCache
}

func New(opts *Options) *Mirror {
	return &Mirror{
		name:        opts.Name,
		labels:      opts.Labels,
		scope:       opts.Scope,
		namespace:   opts.Namespace,
		log:         opts.Log.WithValues("host", opts.Name),
		ring:        opts.Ring,
		mirrorCache: opts.MirrorCache,
		authCache:   NewAuthCache(),
	}
}

func (m *Mirror) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	sem := NewSemaphore()
	wq := NewWorkQueue()

	// TODO: congigurable worker count.
	for i := 0; i < 1; i++ {
		wg.Add(1)
		worker := NewWorker(i).WithLogger(m.log)
		go func(worker *Worker) {
			defer wg.Done()
			worker.Start(ctx, wq, sem)
		}(worker)
	}

	c, err := m.cache(ctx)
	if err != nil {
		return err
	}

	go c.Start(ctx)

	if !c.WaitForCacheSync(ctx) {
		return ctx.Err()
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

func (m *Mirror) cache(ctx context.Context) (cache.Cache, error) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = stvziov1.AddToScheme(scheme)

	c, err := cache.New(ctrl.GetConfigOrDie(), cache.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	err = m.informers(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (m *Mirror) kubeClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = stvziov1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	c, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (m *Mirror) informers(ctx context.Context, c cache.Cache) error {
	// Make sure to block on the pod informer sync so we fully populate
	// the hashring with all the servers before we start processing mirrors.
	// I'm not entirely sure this will work like I think it will, but lets
	// give it a shot.
	pi, err := c.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	})
	if err != nil {
		return err
	}

	_, err = pi.AddEventHandler(&PodHandler{
		Log:       m.log,
		Ring:      m.ring,
		Labels:    m.labels,
		Namespace: m.namespace,
	})
	if err != nil {
		return err
	}

	cli, err := m.kubeClient()
	if err != nil {
		return err
	}

	mi, err := c.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "stvz.io",
		Version: "v1",
		Kind:    "Mirror",
	})
	if err != nil {
		return err
	}

	_, err = mi.AddEventHandler(&MirrorHandler{
		Log:         m.log,
		MirrorCache: m.mirrorCache,
		Client:      cli,
		AuthCache:   m.authCache,
	})
	if err != nil {
		return err
	}

	si, err := c.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})
	if err != nil {
		return err
	}

	_, err = si.AddEventHandler(&SecretHandler{
		Log:       m.log,
		AuthCache: m.authCache,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Mirror) process(ctx context.Context, wq WorkQueue, sem *Semaphore) {
	for _, mirror := range m.mirrorCache.Get() {
		log := m.log.WithValues("mirror", mirror.Name)

		registry := mirror.Spec.Registry

		for _, repo := range mirror.Spec.Repositories {
			log := log.WithValues("repo", *repo.Name, "registry", registry)

			// Normalize repo name without tags.
			nm, err := stvziov1.NormalizeRepoTag(*repo.Name, "")
			if err != nil {
				log.Error(err, "failed to create explicit repo name")
				continue
			}

			tags, err := GetRepositoryTags(ctx, nil, registry, nm)
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
				if m.ring.Mine(m.name, normalized) && !sem.Acquired(normalized) {
					log.V(4).Info("queueing image", "image", normalized)
					wq <- &Item{
						Image:    normalized,
						Registry: registry,
						Auth:     m.authCache.Lookup(normalized),
					}
				}
			}
		}
	}
}
