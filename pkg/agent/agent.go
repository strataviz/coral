// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
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
	log     logr.Logger
	options *AgentOptions
	client  client.Client
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
	sem := NewSemaphore()

	// Start the process workers.
	eq := NewEventQueue()
	for i := 0; i < a.options.WorkerProcesses; i++ {
		wg.Add(1)
		worker := NewWorker(i, a.options)
		go func(worker *Worker) {
			defer wg.Done()
			worker.Start(ctx, eq, sem)
		}(worker)
	}

	// TODO: pull logging out of the function and return descriptive errors.
	err := a.intervalRun(ctx, eq, sem)
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
			if err := a.intervalRun(ctx, eq, sem); err != nil {
				a.log.Error(err, "interval run failed")
			}
		}
	}
}

func (a *Agent) intervalRun(ctx context.Context, eq EventQueue, sem *Semaphore) error {
	// Get the node labels.
	node, err := GetNode(ctx, a.options.NodeName, a.client)
	if err != nil {
		agentError.WithLabelValues("get_node").Inc()
		return err
	}

	err = a.process(ctx, eq, sem, node)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) process(ctx context.Context, eq EventQueue, sem *Semaphore, node *Node) error { // nolint:funlen
	a.log.V(8).Info("processing images", "node", node.GetName())
	// Get all the matched images from the cache.
	images, err := ListImages(ctx, a.client, a.options.Namespace, node.GetLabels())
	if err != nil {
		agentError.WithLabelValues("list_images").Inc()
		return err
	}

	managedImages := make(map[string]string)
	authMap := make(map[string][]*runtime.AuthConfig)

	for _, image := range images {
		for _, data := range image.Status.Data {
			managedImages[data.Name] = data.Label
			authMap[data.Name] = image.RuntimeAuthLookup(data.Name)
		}
	}

	// I think the ImageMap should replace GetNodeImages.
	nodeImages, err := ImageMap(ctx, a.options.ImageServiceClient)
	if err != nil {
		return err
	}

	state := UpdateState(nodeImages, managedImages)
	labels := ReplaceImageLabels(node.GetLabels(), state)
	err = node.UpdateLabels(ctx, a.client, labels)
	if err != nil {
		agentError.WithLabelValues("update_labels").Inc()
		return err
	}

	// NOTE:
	// Originally I had deletion in here, but it may be better behavior to leave
	// the image be and let the kubelet GC it naturally.  If the injector has been
	// activated, the controller will keep new resources off of the node and should
	// trigger cleanup of the image once the pod is removed.
	for name, state := range state {
		auth, ok := authMap[name]
		if !ok {
			a.log.Error(nil, "server error, auth not found for image", "name", name)
			agentError.WithLabelValues("auth_not_found").Inc()
			continue
		}

		if sem.Acquired(name) {
			a.log.V(10).Info("image is already being processed, skipping", "image", name)
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
		case string(stvziov1.ImageStateAvailable):
			a.log.V(8).Info("image is available, skipping", "name", name)
		}
	}

	return nil
}
