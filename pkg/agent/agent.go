package agent

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	cutil "stvz.io/coral/pkg/util"
)

const (
	DefaultParallelPullers int           = 1
	DefaultEventQueueSize  int           = 100
	ConnectionTimeout      time.Duration = 30 * time.Second
	MaxCallRecvMsgSize     int           = 1024 * 1024 * 32
)

type AgentOptions struct {
	Log                  logr.Logger
	ImageServiceClient   runtime.ImageServiceClient
	RuntimeServiceClient runtime.RuntimeServiceClient
	Client               client.Client
	WorkerProcesses      int
	Namespace            string
	NodeName             string
	PollInterval         time.Duration
}

type Agent struct {
	log logr.Logger

	options *AgentOptions

	client client.Client
}

func NewAgent(options *AgentOptions) *Agent {
	return &Agent{
		log:     options.Log,
		client:  options.Client,
		options: options,
	}
}

func (a *Agent) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}

	// Start the process workers.
	eq := NewEventQueue(DefaultEventQueueSize)
	for i := 0; i < a.options.WorkerProcesses; i++ {
		wg.Add(1)
		worker := NewProcessWorker(i, a.options)
		go func(worker *ProcessWorker) {
			defer wg.Done()
			worker.Start(ctx, eq)
		}(worker)
	}

	a.intervalRun(ctx, eq)

	timer := time.NewTicker(a.options.PollInterval)
	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping agent")
			wg.Wait()
			return
		case <-timer.C:
			if err := a.intervalRun(ctx, eq); err != nil {
				a.log.Error(err, "interval run failed")
			}
		}
	}
}

func (a *Agent) intervalRun(ctx context.Context, eq EventQueue) error {
	// Get the node labels.
	node, err := GetNode(ctx, a.options.NodeName, a.client)
	if err != nil {
		return err
	}

	err = a.processImages(ctx, eq, node)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) processImages(ctx context.Context, eq EventQueue, node *Node) error {
	// Get all the matched images from the cache.
	images, err := ListImages(ctx, a.client, a.options.Namespace, node.GetLabels())
	if err != nil {
		return err
	}

	// TODO: I don't like this all that much, so I'll probably refactor it later, but it's
	// better than before.
	managedImages := make(map[string]string)
	authMap := make(map[string][]*runtime.AuthConfig)

	for _, image := range images {
		for _, img := range image.Spec.Images {
			for _, tag := range img.Tags {
				name := *img.Name + ":" + tag
				managedImages[cutil.HashedImageLabelKey(name)] = name
				authMap[name] = image.RuntimeAuthLookup(name)
			}
		}
	}
	nodeImages := node.ImageHashMap()
	nodeLabels := node.GetLabels()

	state := UpdateStateLabels(nodeLabels, nodeImages, managedImages)

	for hash, state := range state {
		name, ok := managedImages[hash]
		// TODO: fix me, we want to be removing in the switch...
		if !ok {
			// Get the name from the available images.
			name, ok = nodeImages[hash]
			if !ok {
				a.log.Error(nil, "server error, image not found", "hash", hash)
				continue
			}
			a.log.V(8).Info("sending remove event", "name", name)
			eq <- &Event{
				Operation: Remove,
				Image:     name,
			}
			continue
		}

		auth, ok := authMap[name]
		if !ok {
			a.log.Error(nil, "server error, auth not found for image", "name", name)
			continue
		}

		switch state {
		case string(stvziov1.ImageStatePending):
			a.log.V(8).Info("sending pull event", "name", name)
			eq <- &Event{
				Operation: Pull,
				Image:     name,
				Auth:      auth,
			}
		case string(stvziov1.ImageStateDeleting):
			a.log.V(8).Info("sending remove event", "name", name)
			eq <- &Event{
				Operation: Remove,
				Image:     name,
			}
		case string(stvziov1.ImageStateAvailable):
			a.log.V(10).Info("image is available, skipping", "name", name)
		}
	}

	// Refresh the node to make sure we get the latest updates before modifying the
	// labels.  We want to make sure that we don't overwrite any changes that have been
	// from other sources.  As for the image labels, we're the only ones that should be
	// modifying them, so we don't need to worry about that.
	err = node.Refresh(ctx, a.client)
	if err != nil {
		return err
	}

	labels := ReplaceImageLabels(node.GetLabels(), state)
	node.UpdateLabels(ctx, a.client, labels)

	return nil
}
