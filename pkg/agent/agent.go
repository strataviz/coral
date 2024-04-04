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
	eq := NewEventQueue()
	for i := 0; i < a.options.WorkerProcesses; i++ {
		wg.Add(1)
		worker := NewWorker(i, a.options)
		go func(worker *Worker) {
			defer wg.Done()
			worker.Start(ctx, eq)
		}(worker)
	}

	// TODO: pull logging out of the function and return descriptive errors.
	err := a.intervalRun(ctx, eq)
	if err != nil {
		a.log.Error(err, "run failed")
	}

	timer := time.NewTicker(a.options.PollInterval)
	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping agent")
			close(eq)
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
		agentError.WithLabelValues("get_node").Inc()
		return err
	}

	err = a.processImages(ctx, eq, node)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) processImages(ctx context.Context, eq EventQueue, node *Node) error { // nolint:funlen
	// Get all the matched images from the cache.
	images, err := ListImages(ctx, a.client, a.options.Namespace, node.GetLabels())
	if err != nil {
		agentError.WithLabelValues("list_images").Inc()
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

	// TODO:  I'd like to get to a place where I'm keying by the name and not the
	// label hash to make things a bit more clean.  The problem is that I'd need to
	// convert the hashes from the labels to get the current node state.

	// If the image is in the list of managed images then we check available or not.
	// if not available, then it's pending and if is available then it's available.
	// if the image is not in the list of managed images and it's available, then we
	// then it's deleting.  The one thing that we do need though is the labels as that
	// is kind of how we track history.  On the other hand, could we introduce a new
	// history object that we could use to track the history of the image on the node
	// and not have to worry about the labels...  It feels like that could be a more
	// reliable way to track the images.  The monitor would be the thing that would
	// need to update the global state (pending/etc) and the worker wouldn't be relying
	// on it for state at all.  That introduces problems though as you'd need to maintain
	// a pretty large history of the images on the node or you only keep the current
	// active ones.

	// Benefits:
	// - track both the hashes and the names in one spot.
	// - track the history of the images on the node (in mvp just the current active).
	// - if tracking deletion, then the fininalizer would just update the status.
	// - only one object to load and update.

	nodeImages, err := ImageHashMap(ctx, a.options.ImageServiceClient)
	if err != nil {
		return err
	}
	nodeLabels := node.GetLabels()

	state := UpdateStateLabels(nodeLabels, nodeImages, managedImages)
	labels := ReplaceImageLabels(node.GetLabels(), state)
	err = node.UpdateLabels(ctx, a.client, labels)
	if err != nil {
		agentError.WithLabelValues("update_labels").Inc()
		return err
	}

	// TODO: Once the workers pick up an event that leaves the queue empty so the
	// loop will try to push on the same event again and then block until processed.
	// I'm not sure if this is the best way to handle this, but it's mostly just an
	// annoyance at this point.  I'll need to address this at some point.
	for hash, state := range state {
		name, ok := managedImages[hash]
		// TODO: fix me, we want to be removing in the switch...
		if !ok {
			// Get the name from the available images.
			name, ok = nodeImages[hash]
			if !ok {
				a.log.Error(nil, "server error, image not found", "hash", hash)
				agentError.WithLabelValues(name, "image_not_found").Inc()
				continue
			}
			a.log.V(8).Info("sending remove event", "name", name)
			agentImageRemovals.Inc()
			eq <- &Event{
				Operation: Remove,
				Image:     name,
			}
			continue
		}

		auth, ok := authMap[name]
		if !ok {
			a.log.Error(nil, "server error, auth not found for image", "name", name)
			agentError.WithLabelValues(name, "auth_not_found").Inc()
			continue
		}

		switch state {
		case string(stvziov1.ImageStatePending):
			a.log.V(8).Info("sending pull event", "name", name)
			agentImagePulls.Inc()
			eq <- &Event{
				Operation: Pull,
				Image:     name,
				Auth:      auth,
			}
		case string(stvziov1.ImageStateDeleting):
			a.log.V(8).Info("sending remove event", "name", name)
			agentImageRemovals.Inc()
			eq <- &Event{
				Operation: Remove,
				Image:     name,
			}
		case string(stvziov1.ImageStateAvailable):
			a.log.V(10).Info("image is available, skipping", "name", name)
		}
	}

	return nil
}
