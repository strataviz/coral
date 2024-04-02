package monitor

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type Manager struct {
	images map[types.NamespacedName]*Monitor
	client client.Client
	log    logr.Logger

	wg       sync.WaitGroup
	stopOnce sync.Once
	sync.Mutex
}

func NewManager(c client.Client, log logr.Logger) *Manager {
	return &Manager{
		client: c,
		images: make(map[types.NamespacedName]*Monitor),
		log:    log,
	}
}

func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		m.Lock()
		defer m.Unlock()

		for _, monitor := range m.images {
			monitor.Stop()
		}
		m.wg.Wait()
	})
}

func (m *Manager) AddImage(ctx context.Context, image *stvziov1.Image) {
	m.Lock()
	defer m.Unlock()

	nn := types.NamespacedName{
		Namespace: image.GetNamespace(),
		Name:      image.GetName(),
	}

	// If the image is already being monitored, stop it first.
	if m, ok := m.images[nn]; ok {
		m.Stop()
	}

	monitor := NewMonitor(m.client, nn).WithLogger(m.log)
	m.images[nn] = monitor

	m.log.V(4).Info("starting monitor", "image", nn)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		monitor.Start(ctx)
	}()
}

func (m *Manager) RemoveImage(image *stvziov1.Image) {
	m.Lock()
	defer m.Unlock()

	nn := types.NamespacedName{
		Namespace: image.GetNamespace(),
		Name:      image.GetName(),
	}

	if monitor, ok := m.images[nn]; ok {
		m.log.V(4).Info("stopping monitor", "image", nn)
		monitor.Stop()
		delete(m.images, nn)
	}
}

func (m *Manager) HasImage(nn types.NamespacedName) bool {
	m.Lock()
	defer m.Unlock()

	_, ok := m.images[nn]
	return ok
}
