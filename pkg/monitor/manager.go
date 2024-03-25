package monitor

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type Manager struct {
	images map[string]*Monitor
	client client.Client
	cache  cache.Cache
	log    logr.Logger

	wg sync.WaitGroup
}

func NewManager(c client.Client, cc cache.Cache, log logr.Logger) *Manager {
	return &Manager{
		client: c,
		cache:  cc,
		images: make(map[string]*Monitor),
		log:    log,
	}
}

func (m *Manager) Stop() {
	for _, monitor := range m.images {
		monitor.Stop()
	}

	m.wg.Wait()
}

func (m *Manager) AddImage(ctx context.Context, image *stvziov1.Image) {
	// If the image is already being monitored, stop it first.
	if m, ok := m.images[image.GetName()]; ok {
		m.Stop()
	}

	monitor := NewMonitor(m.client, m.cache, image.DeepCopy()).WithLogger(m.log)
	m.images[image.GetName()] = monitor

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		monitor.Start(ctx)
	}()
}
