package monitor

import (
	"context"
	"fmt"
	"math"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
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
	for _, img := range image.Spec.Images {
		for _, tag := range img.Tags {
			name := fmt.Sprintf("%s:%s", *img.Name, tag)
			label := util.HashedImageLabelKey(name)
			keys = append(keys, []string{name, label})
			total++
		}
	}

	// Filter managed labels and count the states.
	state := map[string]int{
		"pending":   0,
		"available": 0,
		"deleting":  0,
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
	img.Status.AvailableImages = floor(state["available"], numNodes)
	monitorImagesAvailable.WithLabelValues(image.Name, image.Namespace).Set(float64(state["available"]))
	img.Status.PendingImages = floor(state["pending"], numNodes)
	monitorImagesPending.WithLabelValues(image.Name, image.Namespace).Set(float64(state["pending"]))
	img.Status.DeletingImages = floor(state["deleting"], numNodes)
	monitorImagesDeleting.WithLabelValues(image.Name, image.Namespace).Set(float64(state["deleting"]))
	img.Status.UnknownImages = floor(state["unknown"], numNodes)
	img.Status.TotalImages = total
	img.Status.TotalNodes = numNodes

	return img, nil
}

func floor(a, b int) int {
	return int(math.Floor(float64(a) / float64(b)))
}

// Get filter the labels to those that the image is managing.
// Make a

// for _, image := range images {
// 	for _, tag := range image.Tags {
// 		hash := util.ImageHasher(fmt.Sprintf("%s:%s", *image.Name, tag))
// 		label := util.ImageLabelKey(hash)
// 		req, err := labels.NewRequirement(label, selection.Equals, []string{state})
// 		if err != nil {
// 			m.log.Error(err, "failed to create requirement")
// 			return -1, err
// 		}
// 		selectors = selectors.Add(*req)

// 		nodes := new(corev1.NodeList)
// 		err = m.client.List(ctx, nodes, &client.ListOptions{
// 			LabelSelector: selectors,
// 		})
// 		if err != nil {
// 			return -1, err
// 		}
// 	}
// }

// return len(nodes.Items), nil
// }

// func (m *Worker) addStateRequirement(s labels.Selector, label string, state string) (labels.Selector, error) {
// 	s = s.DeepCopySelector()
// 	req, err := labels.NewRequirement(label, selection.Equals, []string{state})
// 	if err != nil {
// 		m.log.Error(err, "failed to create requirement")
// 		return nil, err
// 	}
// 	return s.Add(*req), nil
// }

// getNodes returns the nodes that match the selector.
// func (m *Worker) getNodes(ctx context.Context, selector labels.Selector) (*corev1.NodeList, error) {
// 	nodes := new(corev1.NodeList)
// 	err := m.cache.List(ctx, nodes, &client.ListOptions{
// 		LabelSelector: selector,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return nodes, nil
// }
