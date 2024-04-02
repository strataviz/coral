package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

const (
	// DefaultMonitorQueueSize is the default size of the monitor queue.  Potentially
	// make this configurable in the future.
	DefaultMonitorQueueSize = 1000
	DefaultMonitorWorkers   = 1
	DefaultMonitorInterval  = 5 * time.Second
)

type Monitor struct {
	client    client.Client
	log       logr.Logger
	namespace string
	stopOnce  sync.Once
	stopChan  chan struct{}
	sync.Mutex
}

func NewMonitor(c client.Client, ns string, log logr.Logger) *Monitor {
	return &Monitor{
		client:    c,
		namespace: ns,
		log:       log,
		stopChan:  make(chan struct{}),
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.Lock()
	defer m.Unlock()

	ch := make(chan types.NamespacedName, DefaultMonitorQueueSize)

	var wg sync.WaitGroup
	// Start workers to monitor the images.
	for i := 0; i < DefaultMonitorWorkers; i++ {
		wg.Add(1)
		worker := NewWorker(m.client).WithLogger(m.log)
		go func() {
			defer wg.Done()
			worker.Start(ctx, ch)
		}()
	}

	go func() {
		timer := time.NewTicker(DefaultMonitorInterval)
		defer timer.Stop()

		for {
			select {
			case <-m.stopChan:
				close(ch)
				wg.Wait()
				return
			case <-ctx.Done():
				m.Stop()
			case <-timer.C:
				err := m.send(ctx, ch)
				if err != nil {
					m.log.Error(err, "failed to gather images")
					monitorError.WithLabelValues("send").Inc()
					continue
				}
			}
		}
	}()
}

func (m *Monitor) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
}

// send gets all the images that it knows about and sends them to the work queue.  It
// supports both namespaced and cluster-scoped images depending on the controller
// configuration.  The conversion  of the image to a namespaced name is done
// intentionally to force the worker to get the image prior to monitoring and
// updating the status in case we end up blocking for an extended period of time on
// the channel.
func (m *Monitor) send(ctx context.Context, ch chan<- types.NamespacedName) error {
	// Get all the images in the namespace.
	m.log.V(10).Info("gathering images")
	images := &stvziov1.ImageList{}
	err := m.client.List(ctx, images, client.InNamespace(m.namespace))
	if err != nil {
		return err
	}

	for _, image := range images.Items {
		ch <- types.NamespacedName{
			Namespace: image.Namespace,
			Name:      image.Name,
		}

	}
	return nil
}
