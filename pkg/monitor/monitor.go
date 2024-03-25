package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/util"
)

type Monitor struct {
	client client.Client
	cache  cache.Cache
	image  *stvziov1.Image
	log    logr.Logger

	syncOnce sync.Once
	stopChan chan struct{}
}

func NewMonitor(c client.Client, cc cache.Cache, image *stvziov1.Image) *Monitor {
	return &Monitor{
		client:   c,
		cache:    cc,
		image:    image,
		log:      logr.Discard(),
		stopChan: make(chan struct{}),
	}
}

func (m *Monitor) WithLogger(log logr.Logger) *Monitor {
	m.log = log.WithName("monitor").WithValues("image", m.image.GetName())
	return m
}

func (m *Monitor) Start(ctx context.Context) {
	m.syncOnce.Do(func() {
		go m.monitor(ctx)
	})
}

func (m *Monitor) Stop() {
	m.syncOnce.Do(func() {
		close(m.stopChan)
	})
}

func (m *Monitor) monitor(ctx context.Context) {
	// TODO: make me configurable?
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()

	m.run(ctx)
	for {
		select {
		case <-m.stopChan:
			return
		case <-timer.C:
			m.run(ctx)
		}
	}
}

func (m *Monitor) run(ctx context.Context) {
	m.log.V(8).Info("starting monitoring run")
	s := labels.NewSelector()
	for _, selector := range m.image.Spec.Selector {
		req, err := labels.NewRequirement(selector.Key, selector.Operator, selector.Values)
		if err != nil {
			m.log.Error(err, "failed to create requirement")
			return
		}
		s = s.Add(*req)
	}

	// Exclude control plane nodes.
	req, err := labels.NewRequirement("node-role.kubernetes.io/control-plane", selection.DoesNotExist, nil)
	if err != nil {
		m.log.Error(err, "failed to create requirement")
		return
	}
	s = s.Add(*req)

	// TODO: Capture errors better
	total, err := m.getTotalNodes(ctx, s)
	if err != nil {
		m.log.Error(err, "failed to get total nodes")
		return
	}
	pending, err := m.getNodesState(ctx, "pending", m.image.Spec.Images, s)
	if err != nil {
		m.log.Error(err, "failed to get pending nodes")
		return
	}

	available, err := m.getNodesState(ctx, "available", m.image.Spec.Images, s)
	if err != nil {
		m.log.Error(err, "failed to get available nodes")
		return
	}

	deleting, err := m.getNodesState(ctx, "deleting", m.image.Spec.Images, s)
	if err != nil {
		m.log.Error(err, "failed to get deleting nodes")
		return
	}

	unknown, err := m.getNodesState(ctx, "unknown", m.image.Spec.Images, s)
	if err != nil {
		m.log.Error(err, "failed to get unknown nodes")
		return
	}

	image := new(stvziov1.Image)
	err = m.client.Get(ctx, client.ObjectKeyFromObject(m.image), image)
	if err != nil {
		m.log.Error(err, "failed to get image")
		return
	}

	m.log.V(8).Info("status", "total", total, "pending", pending, "available", available, "deleting", deleting, "unknown", unknown)

	image = image.DeepCopy()
	image.Status.TotalNodes = total
	image.Status.PendingNodes = pending
	image.Status.AvailableNodes = available
	image.Status.DeletingNodes = deleting
	image.Status.UnknownNodes = unknown

	err = m.client.Status().Update(ctx, image)
	if err != nil {
		m.log.Error(err, "failed to update image status")
	}
}

// getTotalNodes returns the number of nodes that the selector matches.
func (m *Monitor) getTotalNodes(ctx context.Context, s labels.Selector) (int, error) {
	nodes, err := m.getNodes(ctx, s)
	if err != nil {
		return 0, err
	}

	return len(nodes.Items), nil
}

// getPendingNodes returns the number of nodes that have at least one tag pending.
func (m *Monitor) getNodesState(ctx context.Context, state string, images []stvziov1.ImageSpecImages, s labels.Selector) (int, error) {
	found := make(map[string]int)

	for _, image := range images {
		for _, tag := range image.Tags {
			hash := util.ImageHasher(fmt.Sprintf("%s:%s", *image.Name, tag))
			label := util.ImageLabelKey(hash)
			s, err := m.addStateRequirement(s, label, state)
			if err != nil {
				return 0, err
			}

			nodes, err := m.getNodes(ctx, s)
			if err != nil {
				return 0, err
			}

			for _, node := range nodes.Items {
				found[node.Name]++
			}
		}
	}

	return len(found), nil
}

func (m *Monitor) addStateRequirement(s labels.Selector, label string, state string) (labels.Selector, error) {
	s = s.DeepCopySelector()
	req, err := labels.NewRequirement(label, selection.Equals, []string{state})
	if err != nil {
		m.log.Error(err, "failed to create requirement")
		return nil, err
	}
	return s.Add(*req), nil
}

// getNodes returns the nodes that match the selector.
func (m *Monitor) getNodes(ctx context.Context, selector labels.Selector) (*corev1.NodeList, error) {
	nodes := new(corev1.NodeList)
	err := m.cache.List(ctx, nodes, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
