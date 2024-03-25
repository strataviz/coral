package worker

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultParallelPullers int           = 1
	ConnectionTimeout      time.Duration = 30 * time.Second
	MaxCallRecvMsgSize     int           = 1024 * 1024 * 32
)

type Worker struct {
	name       string
	log        logr.Logger
	ims        runtime.ImageServiceClient
	rts        runtime.RuntimeServiceClient
	kubeClient client.Client
	namespace  string
	pullers    []*Puller

	pullInterval  time.Duration
	cleanInterval time.Duration
	stopChan      chan struct{}
	stopOnce      sync.Once
	sync.Mutex
}

func NewWorker() *Worker {
	return &Worker{
		stopChan: make(chan struct{}),
		pullers:  make([]*Puller, DefaultParallelPullers),
	}
}

func (w *Worker) WithLogger(log logr.Logger) *Worker {
	w.log = log
	return w
}

func (w *Worker) WithPullInterval(interval time.Duration) *Worker {
	w.pullInterval = interval
	return w
}

func (w *Worker) WithCleanInterval(interval time.Duration) *Worker {
	w.cleanInterval = interval
	return w
}

func (w *Worker) WithImageServiceClient(client runtime.ImageServiceClient) *Worker {
	w.ims = client
	return w
}

func (w *Worker) WithRuntimeServiceClient(client runtime.RuntimeServiceClient) *Worker {
	w.rts = client
	return w
}

func (w *Worker) WithKubeClient(client client.Client) *Worker {
	w.kubeClient = client
	return w
}

func (w *Worker) WithName(name string) *Worker {
	w.name = name
	return w
}

func (w *Worker) WithNamespace(ns string) *Worker {
	w.namespace = ns
	return w
}

func (w *Worker) Start(ctx context.Context) {
	cleaner := NewCleaner(w.name, w.cleanInterval, w.log, w.kubeClient, w.ims, w.rts)
	go cleaner.Start(ctx)
	defer cleaner.Stop()

	w.InitPullers()
	w.intervalRun(ctx)

	timer := time.NewTicker(w.pullInterval)
	for {
		select {
		case <-ctx.Done():
			w.Stop()
		case <-w.stopChan:
			return
		case <-timer.C:
			w.intervalRun(ctx)
		}
	}
}

func (w *Worker) InitPullers() {
	for i := 0; i < DefaultParallelPullers; i++ {
		w.pullers[i] = NewPuller(i, w.log, w.ims, w.rts)
	}
}

// Stop stops the watcher.
func (w *Worker) Stop() {
	w.Lock()
	defer w.Unlock()
	w.stopOnce.Do(func() {
		w.log.Info("shutting down coral worker")
		close(w.stopChan)
	})
}

func (w *Worker) intervalRun(ctx context.Context) {
	w.Lock()
	defer w.Unlock()

	w.log.V(8).Info("starting interval run")

	node, err := w.get(ctx)
	if err != nil {
		w.log.Error(err, "get node failed")
		return
	}

	images := NewImages().
		WithNamespace(w.namespace).
		WithLogger(w.log)

	err = images.GetImages(ctx, w.kubeClient, node.GetLabels())
	if err != nil {
		w.log.Error(err, "get images failed")
		return
	}

	// Start the pullers.
	w.log.V(8).Info("starting the pullers")
	wg := sync.WaitGroup{}
	wq := make(chan string, 1000)
	for i := 0; i < DefaultParallelPullers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w.pullers[i].Start(ctx, wq)
		}(i)
	}

	state := NewNodeState(node)
	if !state.IsReady() {
		w.log.Info("node not ready", "node", node.GetName(), "state", state)
	}

	// Add pending images to the work queue.
	w.log.V(8).Info("adding the images to the work queue")
	go func() {
		defer close(wq)
		w.log.V(8).Info("adding images to the work queue", "images", images.List())
		for _, img := range images.List() {
			if state.NeedsImage(img) {
				w.log.V(8).Info("adding image to the work queue", "image", img)
				wq <- img
			}
		}
	}()

	// Get the node again to ensure that we have the latest status.
	w.log.V(8).Info("updating the node with the new image labels")
	node, err = w.get(ctx)
	if err != nil {
		w.log.Error(err, "get node failed")
		return
	}
	state = NewNodeState(node)
	// Update the node with the new image labels.
	labels := state.GetUpdatedLabels(images.HashedMap())
	node.SetLabels(labels)
	err = w.kubeClient.Update(ctx, node)
	if err != nil {
		w.log.Error(err, "update node failed")
	}

	// Wait for the pullers to finish.
	w.log.V(8).Info("waiting for the pullers to finish")
	wg.Wait()
}

// get retrieves the node and returns a deep copy.
func (w *Worker) get(ctx context.Context) (*corev1.Node, error) {
	node := corev1.Node{}
	err := w.kubeClient.Get(ctx, client.ObjectKey{Name: w.name}, &node)
	return node.DeepCopy(), err
}
