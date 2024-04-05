package monitor

import (
	"context"
	"math"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type Worker struct {
	client client.Client
	log    logr.Logger
}

func NewWorker(c client.Client) *Worker {
	return &Worker{
		client: c,
		log:    logr.Discard(),
	}
}

func (m *Worker) WithLogger(log logr.Logger) *Worker {
	m.log = log.WithName("worker")
	return m
}

func (m *Worker) Start(ctx context.Context, ch <-chan types.NamespacedName) {
	m.log.V(10).Info("starting monitor worker")
	for nns := range ch {
		m.run(ctx, nns)
	}
	m.log.V(10).Info("stopping monitor worker")
}

func (m *Worker) run(ctx context.Context, nns types.NamespacedName) {
	image := new(stvziov1.Image)
	err := m.client.Get(ctx, nns, image)
	if err != nil {
		// Ignore if the image is not found.  It's been deleted and we should stop monitoring
		//  it. The controller will remove the monitor when the nodes are cleaned up.
		if client.IgnoreNotFound(err) != nil {
			return
		}
		m.log.Error(err, "failed to get image")
		return
	}

	// If we don't have any of the image data yet, just return.  The object
	// hasn't been fully reconciled yet.
	if len(image.Status.Data) == 0 {
		return
	}

	img, err := m.updateStates(ctx, image)
	if err != nil {
		m.log.Error(err, "failed to get image states")
		return
	}

	err = m.client.Status().Update(ctx, img)
	if err != nil {
		m.log.Error(err, "failed to update image status")
		return
	}
}

// getPendingNodes returns the number of nodes that have at least one tag pending.
// TODO: the and logic doesn't get the correct number of nodes unless they are all
// pending.  can we use or logic for all the image labels?
func (m *Worker) updateStates(ctx context.Context, image *stvziov1.Image) (*stvziov1.Image, error) {
	s := labels.NewSelector()
	for _, selector := range image.Spec.Selector {
		req, err := labels.NewRequirement(selector.Key, selector.Operator, selector.Values)
		if err != nil {
			return nil, err
		}
		s = s.Add(*req)
	}

	// Exclude control plane nodes.
	req, err := labels.NewRequirement("node-role.kubernetes.io/control-plane", selection.DoesNotExist, nil)
	if err != nil {
		return nil, err
	}
	s = s.Add(*req)

	// Get all matching nodes.  We only care about the nodes that match our selector
	// (if any).  Once we have the nodes, we can filter out our labels and count the
	// states.
	nodes := new(corev1.NodeList)
	err = m.client.List(ctx, nodes, &client.ListOptions{
		LabelSelector: s,
	})
	if err != nil {
		return nil, err
	}

	numNodes := len(nodes.Items)
	monitorNodesTotal.WithLabelValues(image.Name, image.Namespace).Set(float64(numNodes))

	// Calculate the hashes for the images.
	keys := make([][]string, 0)
	total := 0

	for _, data := range image.Status.Data {
		keys = append(keys, []string{data.Name, data.Label})
		total++
	}

	// Filter managed labels and count the states.
	state := map[string]int{
		"pending":   0,
		"available": 0,
		"unknown":   0,
	}

	for _, node := range nodes.Items {
		labels := node.GetLabels()
		for _, key := range keys {
			if s, ok := labels[key[1]]; ok {
				state[s]++
			} else {
				state["unknown"]++
			}
		}
	}

	img := image.DeepCopy()

	condition := stvziov1.ImageCondition{
		Available: floor(state["available"], numNodes),
		Pending:   floor(state["pending"], numNodes),
		Unknown:   floor(state["unknown"], numNodes),
	}

	img.Status.Condition = condition
	img.Status.TotalImages = total
	img.Status.TotalNodes = numNodes

	monitorImagesAvailable.WithLabelValues(image.Name, image.Namespace).Set(float64(state["available"]))
	monitorImagesPending.WithLabelValues(image.Name, image.Namespace).Set(float64(state["pending"]))
	monitorImagesUnknown.WithLabelValues(image.Name, image.Namespace).Set(float64(state["unknown"]))

	return img, nil
}

func floor(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}
